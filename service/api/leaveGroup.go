package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// leaveGroup gestisce l'endpoint DELETE /groups/:groupId/members/:userId
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Prendiamo i parametri dall'URL
	groupID := ps.ByName("groupId")
	userIDToRemove := ps.ByName("userId")

	// Chiamiamo il database
	err := rt.db.LeaveGroup(groupID, userIDToRemove)
	if err != nil {
		ctx.Logger.WithError(err).Error("leaveGroup error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Errore durante la rimozione dell'utente"})
		return
	}

	// Successo: 204 No Content è lo standard per le DELETE andate a buon fine
	w.WriteHeader(http.StatusNoContent)
}
