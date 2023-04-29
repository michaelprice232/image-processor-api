package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/michaelprice232/image-processor-api/internal/validate-profile"
)

const httpServerPort = ":3000"

func main() {
	logLevelEnvar := os.Getenv("LOG_LEVEL")
	level, err := log.ParseLevel(logLevelEnvar)
	if err != nil || len(logLevelEnvar) == 0 {
		log.SetLevel(log.ErrorLevel)
	}
	log.SetLevel(level)

	client, err := validate_profile.NewClient()
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
	_ = http.ListenAndServe(httpServerPort, r)
}
