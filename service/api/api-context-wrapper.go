package api

import (
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

type httpRouterHandler func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext)

// wrapAuth fa il parsing della richiesta E controlla l'autenticazione Bearer.
// Da usare per TUTTE le rotte tranne il login.
func (rt *_router) wrapAuth(fn httpRouterHandler) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var ctx = reqcontext.RequestContext{
			ReqUUID: reqUUID,
		}

		// === Authorization check ===
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			rt.baseLogger.Warn("Missing Authorization header")
			http.Error(w, `{"message": "Missing Authorization"}`, http.StatusUnauthorized)
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			rt.baseLogger.Warn("Invalid Authorization header format")
			http.Error(w, `{"message": "Invalid Authorization Format"}`, http.StatusUnauthorized)
			return
		}

		// Estraiamo l'ID utente (stringa) e lo mettiamo nel contesto
		userIDStr := strings.TrimPrefix(authHeader, bearerPrefix)
		if userIDStr == "" {
			rt.baseLogger.Warn("Empty Bearer token")
			http.Error(w, `{"message": "Invalid User ID"}`, http.StatusUnauthorized)
			return
		}
		ctx.UserID = userIDStr

		// Logger specifico per la richiesta (includiamo anche l'ID utente)
		ctx.Logger = rt.baseLogger.WithFields(logrus.Fields{
			"reqid":     ctx.ReqUUID.String(),
			"remote-ip": r.RemoteAddr,
			"user-id":   ctx.UserID,
		})

		// Opzionale: se vuoi anche verificare a DB che l'utente esista,
		// puoi chiamare rt.db.CheckUserExists(ctx.UserID) qui.
		// Altrimenti, ci penseranno le foreign keys del database a dare errore!

		fn(w, r, ps, ctx)
	}
}

// wrap fa SOLO il parsing del logger e UUID, NON controlla l'autenticazione.
// Da usare ESCLUSIVAMENTE per la rotta di Login (/session).
func (rt *_router) wrap(fn httpRouterHandler) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var ctx = reqcontext.RequestContext{
			ReqUUID: reqUUID,
		}

		ctx.Logger = rt.baseLogger.WithFields(logrus.Fields{
			"reqid":     ctx.ReqUUID.String(),
			"remote-ip": r.RemoteAddr,
		})

		fn(w, r, ps, ctx)
	}
}
