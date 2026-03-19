package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database" // Importa il DB se serve per i tipi
	"github.com/julienschmidt/httprouter"
)

// getMyConversations gestisce l'endpoint GET /users/:userId/conversations
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Controllo Autorizzazione: un utente può vedere solo le PROPRIE conversazioni
	targetUserID := ps.ByName("userId")
	if ctx.UserID != targetUserID {
		ctx.Logger.Warn("User tried to read someone else's conversations list")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Forbidden: you can only view your own conversations"})
		return
	}

	// 2. Chiamata al database per recuperare la lista
	conversations, err := rt.db.GetMyConversations(ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error retrieving conversations from database")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	// 3. Sicurezza JSON: se non ci sono conversazioni, restituiamo un array vuoto `[]` e non `null`
	if conversations == nil {
		conversations = []database.ConversationSummary{} // Assicurati di usare il tipo corretto del tuo DB
	}

	// 4. Successo! Restituiamo l'array JSON (Status 200 OK di default)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(conversations)
}
