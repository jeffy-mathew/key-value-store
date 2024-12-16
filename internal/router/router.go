// Package router provides the routing configuration for the HTTP server.
//
// This package defines a function to instantiate and configure a new HTTP router
// using the httprouter package. It sets up the necessary endpoints for the key-value
// store service and binds the HTTP methods to the corresponding handler functions.
//
// The New function initializes a new httprouter instance, creates a new store service
// using the provided logger, and configures the routes for setting, getting, and deleting
// keys in the key-value store.
package router

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"

	"codesignal/internal/config"
	"codesignal/internal/repository"
	"codesignal/internal/store"
)

// New instantiates a new http router and
// configures the endpoints of the service.
func New(log zerolog.Logger, repo repository.Store, appConfig *config.Config) http.Handler {
	router := httprouter.New()

	storeService := store.NewService(log, repo, store.Opts{
		MaxKeyLength: appConfig.GetMaxKeyLength(),
		MaxValueSize: appConfig.GetMaxValueSize(),
	})

	router.HandlerFunc(http.MethodPost, "/key", storeService.SetKey)
	router.HandlerFunc(http.MethodGet, "/key/:key", storeService.GetKey)
	router.HandlerFunc(http.MethodDelete, "/key/:key", storeService.DeleteKey)

	return router
}
