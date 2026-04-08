package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyUserName gestisce l'endpoint PUT /users/:userId/username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Leggiamo l'ID bersaglio
	targetUserID := ps.ByName("userId")

	// 2. Controllo Autorizzazione
	if ctx.UserID != targetUserID {
		ctx.Logger.Warn("User tried to change someone else's username")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Forbidden: you can only change your own username"})
		return
	}

	// 3. Parsing del JSON: ORA CERCHIAMO "username"
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON body"})
		return
	}

	// 4. Validazione lunghezza nome
	if len(req.Username) < 3 || len(req.Username) > 16 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Username must be between 3 and 16 characters"})
		return
	}

	// 5. Aggiorniamo il database
	err := rt.db.SetMyUserName(ctx.UserID, req.Username)
	if err != nil {
		ctx.Logger.WithError(err).Error("setMyUserName error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Username already taken or invalid"})
		return
	}

	// 6. Successo! 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
