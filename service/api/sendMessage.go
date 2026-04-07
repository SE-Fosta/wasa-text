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

// sendMessage gestisce l'endpoint POST /conversations/:conversationId/messages
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	contentType := r.Header.Get("Content-Type")

	var messageType, content, photoURL, replyTo string

	// Controlliamo se è un invio di foto o di testo
	if strings.Contains(contentType, "multipart/form-data") {
		// --- Caso 1: Invio Foto ---
		if err := r.ParseMultipartForm(10 << 20); err != nil { // max 10MB
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error parsing form"})
			return
		}

		messageType = "photo"
		replyTo = r.FormValue("replyTo")

		file, _, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Photo file is required"})
			return
		}
		defer file.Close()

		// LEGGAMO IL FILE FISICO E LO TRASFORMIAMO IN BASE64
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			ctx.Logger.WithError(err).Error("Errore nella lettura del file immagine")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Error reading photo"})
			return
		}

		// Capiamo se è un jpeg, un png, ecc...
		mimeType := http.DetectContentType(fileBytes)

		// Convertiamo in stringa Base64
		base64Encoding := base64.StdEncoding.EncodeToString(fileBytes)

		// Creiamo l'URL speciale "data:image/..." che Vue sa leggere nativamente
		photoURL = "data:" + mimeType + ";base64," + base64Encoding

	} else {
		// --- Caso 2: Invio Testo JSON ---
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

	// 3. Inserimento nel Database
	msg, err := rt.db.SendMessage(conversationID, ctx.UserID, messageType, content, photoURL, replyTo)
	if err != nil {
		ctx.Logger.WithError(err).Error("sendMessage error")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	// 4. Successo! Restituiamo il messaggio creato
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}
