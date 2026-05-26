package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// createConversation gestisce POST /users/:userId/conversations
func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req struct {
		TargetUserID string `json:"targetUserId"`
		IsGroup      bool   `json:"isGroup"`
		Name         string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "JSON non valido"})
		return
	}

	if req.IsGroup && req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Il nome del gruppo è obbligatorio"})
		return
	}
	if !req.IsGroup && req.TargetUserID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "targetUserId è obbligatorio per chat singole"})
		return
	}

	myUserID := ps.ByName("userId")
	if myUserID != ctx.UserID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	convID, err := rt.db.CreateConversation(myUserID, req.TargetUserID, req.IsGroup, req.Name)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Errore nella creazione della conversazione")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"conversationId": convID})
}
