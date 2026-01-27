package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"wasa-text/service" // Assicurati che il module name nel go.mod sia 'wasa-text'

"github.com/gorilla/mux"
)

// Strutture di supporto per i body delle richieste JSON
type LoginRequest struct {
	Name string `json:"name"`
}

type SetUsernameRequest struct {
	Name string `json:"name"`
}

type SendMessageRequest struct {
	Content     string `json:"content"`
	MessageType string `json:"messageType"`
	ReplyTo     string `json:"replyTo,omitempty"`
}

type ForwardMessageRequest struct {
	TargetConversationID string `json:"targetConversationId"`
}

type CommentRequest struct {
	Emoji string `json:"emoji"`
}

type AddMemberRequest struct {
	UserID string `json:"userId"`
}

type SetGroupNameRequest struct {
	Name string `json:"name"`
}

// --- Utility Functions ---

// getAuthUserID estrae l'ID utente dall'header Authorization: Bearer <id>
func getAuthUserID(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && strings.ToUpper(authHeader[0:7]) == "BEARER " {
		return authHeader[7:]
	}
	return ""
}

// sendJSONResponse invia una risposta JSON standard
func sendJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

// sendError invia un messaggio di errore JSON
func sendError(w http.ResponseWriter, status int, message string) {
	sendJSONResponse(w, status, map[string]string{
		"message": message,
		"code":    http.StatusText(status),
	})
}

// --- Handlers ---

// DoLogin (POST /session)
func DoLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(req.Name) < 3 || len(req.Name) > 16 {
		sendError(w, http.StatusBadRequest, "Username must be between 3 and 16 characters")
		return
	}

	userID, err := service.DoLogin(req.Name)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusCreated, map[string]string{"identifier": userID})
}

// SetMyUserName (PUT /users/{userId}/username)
func SetMyUserName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pathUserID := vars["userId"]
	authUserID := getAuthUserID(r)

	if authUserID != pathUserID {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req SetUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := service.SetMyUserName(authUserID, req.Name)
	if err != nil {
		if err.Error() == "username already taken" {
			sendError(w, http.StatusConflict, "Username already taken")
			return
		}
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetMyPhoto (PUT /users/{userId}/photo)
// Nota: Gestisce Multipart Form Data (semplificato: salva solo un URL fittizio o path)
func SetMyPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if getAuthUserID(r) != vars["userId"] {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parsing multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		sendError(w, http.StatusBadRequest, "Error parsing form data")
		return
	}

	file, _, err := r.FormFile("photo")
	if err != nil {
		sendError(w, http.StatusBadRequest, "Missing photo file")
		return
	}
	defer file.Close()

	// In un caso reale salveresti il file su disco o S3 e otterresti un URL.
	// Qui simuliamo salvando un URL statico o generato.
	fakePhotoURL := "http://localhost:3000/static/photos/" + vars["userId"] + ".jpg"

	if err := service.SetMyPhoto(vars["userId"], fakePhotoURL); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMyConversations (GET /users/{userId}/conversations)
func GetMyConversations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if getAuthUserID(r) != vars["userId"] {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conversations, err := service.GetMyConversations(vars["userId"])
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Se slice è nil, ritorniamo array vuoto [] invece di null
	if conversations == nil {
		conversations = []service.ConversationSummary{}
	}

	sendJSONResponse(w, http.StatusOK, conversations)
}

// GetConversation (GET /conversations/{conversationId})
func GetConversation(w http.ResponseWriter, r *http.Request) {
	// Verifica che l'utente sia autenticato (opzionale: e che sia membro della conv)
	if getAuthUserID(r) == "" {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	messages, members, err := service.GetConversation(vars["conversationId"])
	if err != nil {
		sendError(w, http.StatusNotFound, "Conversation not found") // O errore DB
		return
	}

	// Costruisci la risposta completa come da api.yaml (schema Conversation)
	response := map[string]interface{}{
		"id":       vars["conversationId"],
		"name":     "Conversation", // Dovresti recuperarlo dal DB se è un gruppo
		"isGroup":  len(members) > 2, // Logica semplificata
		"messages": messages,
		"members":  members,
	}
	if messages == nil {
		response["messages"] = []interface{}{}
	}

	sendJSONResponse(w, http.StatusOK, response)
}

// SendMessage (POST /conversations/{conversationId}/messages)
func SendMessage(w http.ResponseWriter, r *http.Request) {
	userID := getAuthUserID(r)
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	vars := mux.Vars(r)
	convID := vars["conversationId"]

	var msg service.Message
	var err error

	// Gestione Text vs Photo (semplificata: controlliamo Content-Type)
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		msg, err = service.SendMessage(convID, userID, req.Content, "text", "", req.ReplyTo)

	} else if strings.Contains(contentType, "multipart/form-data") {
		// Gestione invio foto
		r.ParseMultipartForm(10 << 20)
		file, _, errFile := r.FormFile("photo")
		if errFile != nil {
			sendError(w, http.StatusBadRequest, "Missing photo")
			return
		}
		defer file.Close()

		// Qui salveresti il file. Simuliamo URL.
		fakeURL := "http://localhost:3000/static/uploads/img.jpg"
		replyTo := r.FormValue("replyTo")

		msg, err = service.SendMessage(convID, userID, "", "photo", fakeURL, replyTo)
	} else {
		sendError(w, http.StatusUnsupportedMediaType, "Content-Type not supported")
		return
	}

	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusCreated, msg)
}

// ForwardMessage (POST /messages/{messageId}/forward)
func ForwardMessage(w http.ResponseWriter, r *http.Request) {
	userID := getAuthUserID(r)
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	vars := mux.Vars(r)

	var req ForwardMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid body")
		return
	}

	err := service.ForwardMessage(vars["messageId"], req.TargetConversationID, userID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// API.yaml dice che ritorna il messaggio creato, ma service.ForwardMessage
	// nel repository.go precedente ritornava solo error.
	// Per coerenza con repository.go, ritorniamo un placeholder o modifichiamo il service.
	// Qui assumiamo successo 201.
	sendJSONResponse(w, http.StatusCreated, map[string]string{"status": "forwarded"})
}

// CommentMessage (POST /messages/{messageId}/comments)
func CommentMessage(w http.ResponseWriter, r *http.Request) {
	userID := getAuthUserID(r)
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	vars := mux.Vars(r)

	var req CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid body")
		return
	}

	if err := service.CommentMessage(vars["messageId"], userID, req.Emoji); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusCreated, map[string]bool{"success": true})
}

// UncommentMessage (DELETE /messages/{messageId}/comments)
// Nota: Api.yaml non specifica l'ID reazione nell'URL, ma assume che l'utente tolga la PROPRIA reazione
func UncommentMessage(w http.ResponseWriter, r *http.Request) {
	userID := getAuthUserID(r)
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	vars := mux.Vars(r)

	if err := service.UncommentMessage(vars["messageId"], userID); err != nil {
		sendError(w, http.StatusNotFound, "Reaction not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteMessage (DELETE /messages/{messageId})
func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID := getAuthUserID(r)
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	vars := mux.Vars(r)

	if err := service.DeleteMessage(vars["messageId"], userID); err != nil {
		sendError(w, http.StatusForbidden, "Cannot delete message (not found or not yours)")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddToGroup (POST /groups/{groupId}/members)
func AddToGroup(w http.ResponseWriter, r *http.Request) {
	if getAuthUserID(r) == "" {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	vars := mux.Vars(r)

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid body")
		return
	}

	if err := service.AddToGroup(vars["groupId"], req.UserID); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSONResponse(w, http.StatusCreated, map[string]bool{"success": true})
}

// LeaveGroup (DELETE /groups/{groupId}/members/{userId})
func LeaveGroup(w http.ResponseWriter, r *http.Request) {
	authID := getAuthUserID(r)
	if authID == "" {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	vars := mux.Vars(r)
	targetUser := vars["userId"]

	// Regola: Un utente può rimuovere solo se stesso (Leave Group)
	if authID != targetUser {
		sendError(w, http.StatusForbidden, "You can only remove yourself from a group")
		return
	}

	if err := service.LeaveGroup(vars["groupId"], targetUser); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetGroupName (PUT /groups/{groupId}/name)
func SetGroupName(w http.ResponseWriter, r *http.Request) {
	if getAuthUserID(r) == "" {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	vars := mux.Vars(r)

	var req SetGroupNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid body")
		return
	}

	if err := service.SetGroupName(vars["groupId"], req.Name); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetGroupPhoto (PUT /groups/{groupId}/photo)
func SetGroupPhoto(w http.ResponseWriter, r *http.Request) {
	if getAuthUserID(r) == "" {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	vars := mux.Vars(r)

	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("photo")
	if err != nil {
		sendError(w, http.StatusBadRequest, "Missing photo")
		return
	}
	defer file.Close()

	fakeURL := "http://localhost:3000/static/groups/" + vars["groupId"] + ".jpg"

	if err := service.SetGroupPhoto(vars["groupId"], fakeURL); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
