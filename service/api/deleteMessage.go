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

	err := rt.db.DeleteMessage(messageID, ctx.UserID)
	if err != nil {
		// Se è l'errore specifico che hai creato tu nel database
		if err.Error() == "message not found or forbidden" {
			ctx.Logger.WithError(err).Warn("Attempt to delete message failed (not found or forbidden)")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden) // 403
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Message not found or you are not the sender"})
			return
		}

		// Se è un VERO errore del database (es. connessione persa)
		ctx.Logger.WithError(err).Error("Errore interno del database")
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}
}
