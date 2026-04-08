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
	defer file.Close()

	// 4. Salviamo fisicamente il file su disco
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, os.ModePerm) // Crea la cartella se non esiste

	// Usiamo l'ID utente per il nome del file
	fileName := "profile_" + ctx.UserID + ".jpg"
	filePath := filepath.Join(uploadDir, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error creating file on disk")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}
	defer dst.Close()

	// Copia i dati dell'immagine nel file appena creato
	if _, err := io.Copy(dst, file); err != nil {
		ctx.Logger.WithError(err).Error("Error saving image")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	// 5. Aggiorniamo il database passandogli ESATTAMENTE realPhotoURL
	if err := rt.db.SetMyPhoto(ctx.UserID, "/uploads/"+fileName); err != nil {
		ctx.Logger.WithError(err).Error("Database error updating photo")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	// 6. Successo! 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
