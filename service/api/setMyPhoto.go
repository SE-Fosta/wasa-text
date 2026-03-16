package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyPhoto gestisce l'endpoint PUT /users/:userId/photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Controllo Autorizzazione
	targetUserID := ps.ByName("userId")
	if ctx.UserID != targetUserID {
		ctx.Logger.Warn("User tried to change someone else's photo")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Forbidden: you can only change your own photo"})
		return
	}

	// 2. Parsing del form (limite di ~10 MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ctx.Logger.WithError(err).Error("Error parsing multipart form")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error parsing form data"})
		return
	}

	// 3. Estraiamo il file dal campo "photo"
	file, _, err := r.FormFile("photo")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Photo file is required"})
		return
	}
	defer file.Close() // Ricordiamoci sempre di chiudere il file per non sprecare memoria!

	// 4. In un'app reale salveresti il file su disco.
	// Qui simuliamo la creazione di un URL fittizio basato sull'ID utente.
	fakePhotoURL := "/photos/user_" + ctx.UserID + ".jpg"

	// 5. Aggiorniamo il database
	if err := rt.db.SetMyPhoto(ctx.UserID, fakePhotoURL); err != nil {
		ctx.Logger.WithError(err).Error("Database error updating photo")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	// 6. Successo! 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
