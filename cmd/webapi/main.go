package main

import (
	"log"
	"net/http"
	"time"
	"wasa-text/service/api"     // Assicurati che questo path corrisponda al tuo go.mod
	"wasa-text/service/database"

	"github.com/gorilla/mux"
)

func main() {
	// 1. Configurazione Database
	// Modifica la stringa di connessione secondo il tuo setup MySQL locale o Docker
	dsn := "root:password@tcp(127.0.0.1:3306)/wasatext?parseTime=true"

	db, err := database.InitDatabase(dsn)
	if err != nil {
		log.Fatalf("Errore connessione DB: %v", err)
	}
	defer db.Close()

	// 2. Setup Router
	r := mux.NewRouter()

	// --- Login ---
	r.HandleFunc("/session", api.DoLogin).Methods("POST", "OPTIONS")

	// --- Utenti ---
	r.HandleFunc("/users/{userId}/username", api.SetMyUserName).Methods("PUT", "OPTIONS")
	r.HandleFunc("/users/{userId}/photo", api.SetMyPhoto).Methods("PUT", "OPTIONS")
	r.HandleFunc("/users/{userId}/conversations", api.GetMyConversations).Methods("GET", "OPTIONS")

	// --- Conversazioni ---
	r.HandleFunc("/conversations/{conversationId}", api.GetConversation).Methods("GET", "OPTIONS")
	r.HandleFunc("/conversations/{conversationId}/messages", api.SendMessage).Methods("POST", "OPTIONS")

	// --- Messaggi ---
	r.HandleFunc("/messages/{messageId}/forward", api.ForwardMessage).Methods("POST", "OPTIONS")
	r.HandleFunc("/messages/{messageId}/comments", api.CommentMessage).Methods("POST", "OPTIONS")
	r.HandleFunc("/messages/{messageId}/comments", api.UncommentMessage).Methods("DELETE", "OPTIONS") // Nota: endpoint semplificato
	r.HandleFunc("/messages/{messageId}", api.DeleteMessage).Methods("DELETE", "OPTIONS")

	// --- Gruppi ---
	r.HandleFunc("/groups/{groupId}/members", api.AddToGroup).Methods("POST", "OPTIONS")
	r.HandleFunc("/groups/{groupId}/members/{userId}", api.LeaveGroup).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/groups/{groupId}/name", api.SetGroupName).Methods("PUT", "OPTIONS")
	r.HandleFunc("/groups/{groupId}/photo", api.SetGroupPhoto).Methods("PUT", "OPTIONS")

	// 3. Configurazione Server
	srv := &http.Server{
		Handler:      CORSMiddleware(r), // Wrappa il router col middleware CORS creato nel passo precedente
		Addr:         ":3000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("WASAText Backend listening on port 3000...")
	log.Fatal(srv.ListenAndServe())
}
