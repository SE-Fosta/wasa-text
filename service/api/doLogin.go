package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// doLogin gestisce l'endpoint POST /session
// Riceve un JSON con il nome utente e restituisce l'identificatore (Bearer token fittizio)
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Definiamo la struttura attesa nel body della richiesta
	var req struct {
		Name string `json:"name"`
	}

	// 2. Facciamo il parsing del JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("doLogin: error decoding JSON")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON body"})
		return
	}

	// 3. Validazione base dell'username (tra 3 e 16 caratteri come da OpenAPI)
	if len(req.Name) < 3 || len(req.Name) > 16 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Username must be between 3 and 16 characters"})
		return
	}

	// 4. Chiamiamo il Database per effettuare il login/registrazione
	identifier, err := rt.db.DoLogin(req.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("doLogin: database error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	// 5. Restituiamo l'identificatore (con Status 201 Created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"identifier": identifier,
	})
}
