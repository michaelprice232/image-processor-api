package validate_profile

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	s3Object "github.com/michaelprice232/image-processor-api/internal/s3-object-created-schema"
)

type ErrorHTTPResponse struct {
	ErrorMessage string `json:"message"`
	StatusCode   int    `json:"code"`
}

// HealthEndpoint is an endpoint to be used by K8s health probes
func (c *Client) HealthEndpoint(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
	// todo: check that Kafka is available
}

// ProcessHTTPRequest extracts the EventBridge (S3) event from the request payload and then calls processImage to validate image.
// If any errors are encountered parsing the request body then return an HTTP 400 error.
// If body parsed OK, send to 1 of 2 Kafka (success vs failure) topics dependent upon whether image validation has passed, to allow for processing by downstream services.
func (c *Client) ProcessHTTPRequest(w http.ResponseWriter, r *http.Request) {
	event, err := parseRequestBody(r)
	if err != nil {
		writeHTTPErrorResponse(fmt.Sprintf("error: parsing request body: %v", err), 400, w)
	}

	err = c.processImage(event.Detail.Bucket.Name, event.Detail.Object.Key)
	if err != nil {
		// send to the failed Kafka topic and include the validation failure reason
		kafkaError := c.sendMessage(c.failedKafkaTopic, kafkaResponseEvent{
			Event:        event,
			Outcome:      "failed",
			ErrorMessage: err.Error(),
		})
		if kafkaError != nil {
			writeHTTPErrorResponse(fmt.Sprintf("error sending message to topic %s: %v", c.failedKafkaTopic, kafkaError), 500, w)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}
	// send to the successful Kafka topic
	err = c.sendMessage(c.successKafkaTopic, kafkaResponseEvent{
		Event:   event,
		Outcome: "success",
	})
	if err != nil {
		writeHTTPErrorResponse(fmt.Sprintf("error sending message to topic %s: %v", c.successKafkaTopic, err), 500, w)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// parseRequestBody extracts the HTTP request payload into an AWSEvent type
func parseRequestBody(r *http.Request) (s3Object.AWSEvent, error) {
	var event s3Object.AWSEvent

	rawPayload, err := io.ReadAll(r.Body)
	if err != nil || len(rawPayload) == 0 {
		return event, fmt.Errorf("error: unable to parse the request body: %v", err)
	}

	event, err = s3Object.UnmarshalEvent(rawPayload)
	if err != nil {
		return event, fmt.Errorf("error: unable to unmarshal request body into AWSEvent type: %v", err)
	}

	if len(event.Detail.Bucket.Name) == 0 || len(event.Detail.Object.Key) == 0 {
		return event, fmt.Errorf("error: event.Detail.Bucket.Name or event.Detail.Object.Key can't be empty")
	}
	return event, nil
}

// writeHTTPErrorResponse writes an ErrorHTTPResponse to the http.ResponseWriter and sets the status code
func writeHTTPErrorResponse(body string, code int, w http.ResponseWriter) {
	resp := ErrorHTTPResponse{
		ErrorMessage: body,
		StatusCode:   code,
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("error marshalling ErrorHTTPResponse: %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(jsonResp)
	if err != nil {
		log.Errorf("error writing JSON error response: %v", err)
	}
	log.Debug(body)
}
