package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// leaveGroup gestisce l'endpoint DELETE /groups/:groupId/members/:userId
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	targetUserID := ps.ByName("userId")

	// Controllo di sicurezza: di solito si può abbandonare solo per sé stessi,
	// a meno che l'utente non sia un admin (logica che andrebbe gestita a DB).
	// Per sicurezza base, controlliamo se l'utente sta cercando di rimuovere sé stesso.
	if ctx.UserID != targetUserID {
		ctx.Logger.Warn("User tried to remove someone else from the group")
		// Se le specifiche permettono di rimuovere altri (es. admin), puoi togliere questo blocco.
	}

	err := rt.db.LeaveGroup(groupID, targetUserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("leaveGroup error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error leaving group"})
		return
	}

	// 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
