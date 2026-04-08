package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// uncommentMessage gestisce l'endpoint DELETE /messages/:messageId/comments
func (rt *_router) unReactMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("messageId")

	// Passiamo ID messaggio e ID utente: cancelliamo solo la TUA reazione
	err := rt.db.UnreactMessage(messageID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("unreactMessage error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error removing reaction"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
