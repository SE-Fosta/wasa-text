package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// sendMessage gestisce POST /conversations/:conversationId/messages
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	contentType := r.Header.Get("Content-Type")

	var messageType, content, photoURL, replyTo string

	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // max 10MB
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error parsing form"})
			return
		}

		content = r.FormValue("content")
		replyTo = r.FormValue("replyTo")

		messageType = r.FormValue("messageType")
		if messageType == "" {
			messageType = "photo"
		}

		file, _, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Photo file is required"})
			return
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			ctx.Logger.WithError(err).Error("Errore nella lettura del file immagine")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error reading photo"})
			return
		}

		mimeType := http.DetectContentType(fileBytes)

		base64Encoding := base64.StdEncoding.EncodeToString(fileBytes)

		photoURL = "data:" + mimeType + ";base64," + base64Encoding

	} else {
		var req struct {
			Content     string `json:"content"`
			MessageType string `json:"messageType"`
			ReplyTo     string `json:"replyTo"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON body"})
			return
		}

		messageType = req.MessageType
		content = req.Content
		replyTo = req.ReplyTo
	}

	msg, err := rt.db.SendMessage(conversationID, ctx.UserID, messageType, content, photoURL, replyTo)
	if err != nil {
		ctx.Logger.WithError(err).Error("sendMessage error")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}
