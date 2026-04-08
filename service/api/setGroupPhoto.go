package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setGroupPhoto gestisce l'endpoint PUT /groups/:groupId/photo
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")

	// 1. Limite dimensione (es. 5MB)
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)
	if err := r.ParseMultipartForm(5 * 1024 * 1024); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 2. Recuperiamo il file dal form (chiave "photo")
	file, header, err := r.FormFile("photo")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. Creiamo un nome file unico per il gruppo
	ext := filepath.Ext(header.Filename)
	fileName := "group_" + groupID + ext
	filePath := filepath.Join("uploads", fileName)

	// 4. Salviamo il file sul disco
	dst, err := os.Create(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 5. Aggiorniamo il database con l'URL (es: /uploads/group_1.jpg)
	photoURL := "/uploads/" + fileName
	err = rt.db.SetGroupPhoto(groupID, photoURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 6. Successo! Restituiamo il nuovo URL al frontend
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"photoUrl": photoURL})
}
