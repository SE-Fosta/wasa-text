package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setGroupName gestisce l'endpoint PUT /groups/:groupId/name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")

	// 1. Parsing del JSON
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON body"})
		return
	}

	// 2. Validazione minima (es. nome non vuoto e non troppo lungo)
	if len(req.Name) < 3 || len(req.Name) > 30 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Il nome deve essere tra 3 e 30 caratteri"})
		return
	}

	// 3. Chiamata al Database
	err := rt.db.SetGroupName(groupID, req.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("setGroupName error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Errore durante l'aggiornamento del nome"})
		return
	}

	// 4. Successo: 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
