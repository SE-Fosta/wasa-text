package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMessages gestisce l'endpoint GET /conversations/:conversationId/messages
func (rt *_router) getMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Estraiamo l'ID della chat dall'URL
	conversationID := ps.ByName("conversationId")

	// 2. Chiediamo i messaggi al database
	messages, err := rt.db.GetMessages(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Errore nel recupero dei messaggi")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. Prepariamo la busta per la risposta (tutto OK, è un JSON)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 4. LA SOLUZIONE FURBA:
	// Se la query non ha trovato messaggi (messages è nil),
	// scriviamo noi a mano l'array vuoto "[]" e ci fermiamo qui!
	if messages == nil {
		_, _ = w.Write([]byte("[]"))
		return
	}

	// 5. Se invece ci sono messaggi, li codifichiamo normalmente
	_ = json.NewEncoder(w).Encode(messages)
}
