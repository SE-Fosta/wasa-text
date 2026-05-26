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

// setMyPhoto gestisce PUT /users/:userId/photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	targetUserID := ps.ByName("userId")
	if ctx.UserID != targetUserID {
		ctx.Logger.Warn("User tried to change someone else's photo")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Forbidden: you can only change your own photo"})
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ctx.Logger.WithError(err).Error("Error parsing multipart form")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error parsing form data"})
		return
	}

	file, _, err := r.FormFile("photo")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Photo file is required"})
		return
	}
	defer file.Close()

	uploadDir := "./uploads"
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		ctx.Logger.WithError(err).Error("errore durante la creazione della cartella di upload")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

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

	if _, err := io.Copy(dst, file); err != nil {
		ctx.Logger.WithError(err).Error("Error saving image")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	if err := rt.db.SetMyPhoto(ctx.UserID, "/uploads/"+fileName); err != nil {
		ctx.Logger.WithError(err).Error("Database error updating photo")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
