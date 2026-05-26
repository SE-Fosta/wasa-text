package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getConversation gestisce GET /conversations/:conversationId
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	conv, err := rt.db.GetConversation(conversationID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Warn("Conversation not found or access denied")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // 404 Not Found se la chat non esiste o l'utente non vi appartiene
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Conversation not found"})
		return
	}

	_ = json.NewEncoder(w).Encode(conv)
}
