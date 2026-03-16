package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// addToGroup gestisce l'endpoint POST /groups/:groupId/members
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")

	// 1. Leggiamo l'ID dell'utente da aggiungere dal JSON
	var req struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON body"})
		return
	}

	// 2. Chiamiamo il DB per aggiungere l'utente
	// (Il DB dovrebbe verificare se chi fa la richiesta, ctx.UserID, è autorizzato a farlo,
	// ad esempio se è già membro del gruppo)
	err := rt.db.AddToGroup(groupID, req.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("addToGroup error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error adding user to group"})
		return
	}

	// 3. Successo (201 Created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
