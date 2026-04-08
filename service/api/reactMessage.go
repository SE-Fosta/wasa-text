package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// commentMessage gestisce l'endpoint POST /messages/:messageId/comments
func (rt *_router) reactMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("messageId")

	var req struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON body"})
		return
	}

	if req.Emoji == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Emoji cannot be empty"})
		return
	}

	// Salva la reazione nel database
	err := rt.db.ReactMessage(messageID, ctx.UserID, req.Emoji)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		ctx.Logger.WithError(err).Error("reactMessage error")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error adding reaction"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Restituiamo un semplice ack
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
