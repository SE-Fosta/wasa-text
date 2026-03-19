/*
 * Package api exposes the main API engine. All HTTP APIs are handled here - so-called "business logic" should be here, or
 * in a dedicated package (if that logic is complex enough).
 */
package api

import (
	"errors"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// Config is used to provide dependencies and configuration to the New function.
type Config struct {
	// Logger where log entries are sent
	Logger logrus.FieldLogger

	// Database is the instance of database.AppDatabase where data are saved
	Database database.AppDatabase
}

// Router is the package API interface representing an API handler builder
type Router interface {
	// Handler returns an HTTP handler for APIs provided in this package
	Handler() http.Handler

	// Close terminates any resource used in the package
	Close() error
}

type _router struct {
	router *httprouter.Router

	// baseLogger is a logger for non-requests contexts, like goroutines or background tasks not started by a request.
	baseLogger logrus.FieldLogger

	db database.AppDatabase
}

// New returns a new Router instance
func New(cfg Config) (Router, error) {
	// Check if the configuration is correct
	if cfg.Logger == nil {
		return nil, errors.New("logger is required")
	}
	if cfg.Database == nil {
		return nil, errors.New("database is required")
	}

	// Create a new router where we will register HTTP endpoints.
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	rt := &_router{
		router:     router,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}

	// NOTA: Le rotte e il CORS sono gestiti internamente al metodo Handler() in api-handler.go!

	return rt, nil
}

// Termina ogni risorsa usata nel package
func (rt *_router) Close() error {
	// Se hai risorse da chiudere nel database, puoi chiamarle qui.
	// Esempio: return rt.db.Close() se il tuo AppDatabase espone un metodo Close()
	return nil
}
