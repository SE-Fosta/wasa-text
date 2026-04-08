package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req struct {
		TargetUserID string `json:"targetUserId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TargetUserID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "targetUserId is required"})
		return
	}
	// Leggiamo il TUO id dall'URL
	myUserID := ps.ByName("userId")

	// Sicurezza: sei davvero tu?
	if myUserID != ctx.UserID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Leggiamo l'id dell'ALTRA persona dal JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Creiamo o recuperiamo la chat dal DB
	convID, err := rt.db.CreateConversation(myUserID, req.TargetUserID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Errore nella creazione della conversazione")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Rispondiamo al frontend con l'ID della chat
	response := map[string]string{
		"conversationId": convID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}
