package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// addToGroup gestisce l'endpoint POST /groups/:groupId/members
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")

	var req struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON body"})
		return
	}

	// 1. Chiamiamo il DB
	err := rt.db.AddToGroup(groupID, req.UserID)

	if err != nil {
		if err == errors.New("user is already in the group") {
			// MAGIA: "Non lo inserisce e basta".
			// Rispondiamo con un 200 OK (invece di 201 Created) e un messaggio generico.
			// NON inviamo i dati dell'utente, così il frontend sa che non deve aggiornare la grafica.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Utente già nel gruppo"})
			return
		}

		// Se è un errore vero
		ctx.Logger.WithError(err).Error("addToGroup error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Errore interno del server"})
		return
	}

	// 2. RECUPERO DEI DATI PER IL FRONTEND (Solo se è stato appena aggiunto)
	addedUser, err := rt.db.GetUser(req.UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	// 3. Successo VERO: Restituiamo 201 Created e l'oggetto Utente completo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(addedUser)
}
