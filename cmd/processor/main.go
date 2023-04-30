package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/michaelprice232/image-processor-api/internal/validate-profile"
)

const (
	httpServerPort    = ":3000"
	successKafkaTopic = "success-validate-profile-image-v1"
	failedKafkaTopic  = "failed-validate-profile-image-v1"
)

func main() {
	setLogLevel()

	client, err := validate_profile.NewClient(successKafkaTopic, failedKafkaTopic)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	if log.IsLevelEnabled(log.DebugLevel) {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)
	r.Get("/health", client.HealthEndpoint)
	r.Post("/validate", client.ProcessHTTPRequest)

	log.Infof("Starting server on port: %s", httpServerPort)
	log.Fatal(http.ListenAndServe(httpServerPort, r))
}

func setLogLevel() {
	logLevelEnvar := os.Getenv("LOG_LEVEL")
	level, err := log.ParseLevel(logLevelEnvar)
	if err != nil || len(logLevelEnvar) == 0 {
		log.SetLevel(log.ErrorLevel)
	}
	log.SetLevel(level)
}
