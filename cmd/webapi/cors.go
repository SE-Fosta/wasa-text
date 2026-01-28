package main

import "net/http"

// CORSMiddleware gestisce le policy CORS come richiesto dalle specifiche:
// - Allow all origins
// - Max-Age 1 secondo
// - Permette gli header necessari al frontend (Authorization, Content-Type)
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Specifica : "you should allow all origins"
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Specifica : "set the Max-Age attribute to 1 second"
		w.Header().Set("Access-Control-Max-Age", "1")

		// Metodi permessi
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		// Header permessi (Fondamentale: aggiungiamo Authorization per il login)
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Gestione richiesta Pre-flight (OPTIONS)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
