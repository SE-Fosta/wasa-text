package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getConversation gestisce l'endpoint GET /conversations/:conversationId
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Estraiamo l'ID della conversazione dall'URL
	conversationID := ps.ByName("conversationId")

	// 2. Chiamata al DB
	// Passiamo anche ctx.UserID al DB così può verificare se l'utente fa parte di quella chat!
	// (Se non ne fa parte, il DB dovrebbe restituire un errore o un not found)
	conv, err := rt.db.GetConversation(conversationID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Warn("Conversation not found or access denied")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // 404 Not Found se la chat non esiste o l'utente non vi appartiene
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Conversation not found"})
		return
	}

	// 3. Successo! Restituiamo l'oggetto Conversation completo (con membri e messaggi)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(conv)
}
