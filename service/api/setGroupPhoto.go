package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setGroupPhoto gestisce l'endpoint PUT /groups/:groupId/photo
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")

	// 1. Parsing del form (limite di ~10 MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ctx.Logger.WithError(err).Error("Error parsing multipart form")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error parsing form data"})
		return
	}

	// 2. Estraiamo il file
	file, _, err := r.FormFile("photo")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Photo file is required"})
		return
	}
	defer file.Close()

	// 3. Generiamo un URL fittizio per la foto del gruppo
	fakePhotoURL := "/photos/group_" + groupID + ".jpg"

	// 4. Aggiorniamo il database
	if err := rt.db.SetGroupPhoto(groupID, fakePhotoURL); err != nil {
		ctx.Logger.WithError(err).Error("setGroupPhoto error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	// 5. Successo!
	w.WriteHeader(http.StatusNoContent)
}
