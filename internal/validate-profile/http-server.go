package validate_profile

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	s3ObjectCreatedSchema "github.com/michaelprice232/image-processor-api/internal/s3-object-created-schema"
)

type ErrorHTTPResponse struct {
	ErrorMessage string `json:"message"`
	Code         int    `json:"code"`
}

// HealthEndpoint is an endpoint to be used by K8s health probes
func (c *Client) HealthEndpoint(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

// ProcessHTTPRequest extracts the EventBridge (S3) event from the request payload and then calls processImage to validate image
// If any errors are encountered parsing the request body then return an HTTP 400 error
// If body parsed OK, send to 1 of 2 Kafka topics dependent upon whether image validation has passed, to allow for processing by downstream services
func (c *Client) ProcessHTTPRequest(w http.ResponseWriter, r *http.Request) {
	rawPayload, err := io.ReadAll(r.Body)
	if err != nil || len(rawPayload) == 0 {
		writeHTTPErrorResponse(fmt.Sprintf("error: unable to parse the request body: %v", err), 400, w)
		return
	}

	event, err := s3ObjectCreatedSchema.UnmarshalEvent(rawPayload)
	if err != nil {
		writeHTTPErrorResponse(fmt.Sprintf("error: unable to unmarshal request body into AWSEvent event: %v", err), 400, w)
		return
	}

	if len(event.Detail.Bucket.Name) == 0 || len(event.Detail.Object.Key) == 0 {
		writeHTTPErrorResponse("event.Detail.Bucket.Name or event.Detail.Object.Key can't be empty", 400, w)
		return
	}

	err = c.processImage(event.Detail.Bucket.Name, event.Detail.Object.Key)
	if err != nil {
		// send to the failed Kafka topic
		err = c.sendMessage(c.failedKafkaTopic, []byte(fmt.Sprintf("%s/%s", event.Detail.Bucket.Name, event.Detail.Object.Key)), event)
		if err != nil {
			writeHTTPErrorResponse(fmt.Sprintf("error sending message: %v", err), 500, w)
		}
		return
	}
	// send to the successful Kafka topic
	err = c.sendMessage(c.successKafkaTopic, []byte(fmt.Sprintf("%s/%s", event.Detail.Bucket.Name, event.Detail.Object.Key)), event)
	if err != nil {
		writeHTTPErrorResponse(fmt.Sprintf("error sending message: %v", err), 500, w)
	}
}

// writeHTTPErrorResponse writes an ErrorHTTPResponse to the http.ResponseWriter and sets the status code
func writeHTTPErrorResponse(body string, code int, w http.ResponseWriter) {
	resp := ErrorHTTPResponse{
		ErrorMessage: body,
		Code:         code,
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
