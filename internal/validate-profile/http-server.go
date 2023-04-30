package validate_profile

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	s3ObjectCreatedSchema "github.com/michaelprice232/image-processor-api/internal/s3-object-created-schema"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// HealthEndpoint is an endpoint to be used by K8s health probes
func (c *Client) HealthEndpoint(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

// ProcessHTTPRequest extracts the EventBridge event from the request payload and then calls processImage to validate image
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

	err = c.processImage(event.Detail.Bucket.Name, event.Detail.Object.Key)
	if err != nil {
		writeHTTPErrorResponse(fmt.Sprintf("error: processing event: %v", err), 400, w)
		return
	}
}

// writeHTTPErrorResponse writes an ErrorResponse to the http.ResponseWriter and sets the status code
func writeHTTPErrorResponse(body string, code int, w http.ResponseWriter) {
	resp := ErrorResponse{
		Message: body,
		Code:    code,
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("error marshalling ErrorResponse: %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(jsonResp)
	if err != nil {
		log.Errorf("error writing JSON error response: %v", err)
	}
	log.Debug(body)
}
