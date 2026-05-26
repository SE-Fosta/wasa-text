package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
	"github.com/julienschmidt/httprouter"
)

// getMyConversations gestisce GET /users/:userId/conversations
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	targetUserID := ps.ByName("userId")
	if ctx.UserID != targetUserID {
		ctx.Logger.Warn("User tried to read someone else's conversations list")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Forbidden: you can only view your own conversations"})
		return
	}

	conversations, err := rt.db.GetMyConversations(ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error retrieving conversations from database")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	if conversations == nil {
		conversations = []database.ConversationSummary{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(conversations)
}
