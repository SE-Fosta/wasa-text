package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// forwardMessage gestisce l'endpoint POST /messages/:messageId/forward
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("messageId")

	var req struct {
		TargetConversationID string `json:"targetConversationId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON body"})
		return
	}

	// Chiamata al database passando ctx.UserID (così il DB sa chi risulta come mittente dell'inoltro)
	msg, err := rt.db.ForwardMessage(messageID, req.TargetConversationID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("forwardMessage error")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error forwarding message"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}
