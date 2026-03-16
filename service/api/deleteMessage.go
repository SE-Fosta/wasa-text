package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// deleteMessage gestisce l'endpoint DELETE /messages/:messageId
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("messageId")

	// Passiamo ctx.UserID: il database si occuperà di verificare che
	// la riga appartenga effettivamente a lui prima di eliminarla.
	err := rt.db.DeleteMessage(messageID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Warn("Attempt to delete message failed (not found or forbidden)")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Message not found or you are not the sender"})
		return
	}

	// 204 No Content in caso di successo
	w.WriteHeader(http.StatusNoContent)
}
