package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Ci serve l'ID del commento specifico da eliminare
	commentID := ps.ByName("commentId")

	err := rt.db.UncommentMessage(commentID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("uncommentMessage error")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error deleting comment"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
