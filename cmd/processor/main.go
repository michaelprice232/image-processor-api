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

	client, err := validate_profile.NewClient(successKafkaTopic, failedKafkaTopic, getStringEnvar("kafka-bootstrap-servers"))
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	if log.IsLevelEnabled(log.DebugLevel) {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)

	r.Get("/health", client.HealthEndpoint)

	// Protected routes
	r.Route("/validate", func(r chi.Router) {
		r.Use(validate_profile.BearerTokenAuth(getStringEnvar("api-key")))
		r.Post("/", client.ProcessHTTPRequest)
	})

	log.Infof("Starting server on port: %s", httpServerPort)
	log.Fatal(http.ListenAndServe(httpServerPort, r))
}

func setLogLevel() {
	logLevelEnvar := os.Getenv("LOG_LEVEL")
	level, err := log.ParseLevel(logLevelEnvar)
	if err != nil || len(logLevelEnvar) == 0 {
		log.SetLevel(log.ErrorLevel)
		return
	}
	log.SetLevel(level)
}

func getStringEnvar(key string) string {
	envar := os.Getenv(key)
	if len(envar) == 0 {
		log.Fatalf("unable to load envar: %v", key)
	}
	return envar
}
