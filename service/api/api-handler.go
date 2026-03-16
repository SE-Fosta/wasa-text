package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// -------------------------------------------------------------------------
	// Configurazione CORS per le richieste Pre-flight (OPTIONS)
	// -------------------------------------------------------------------------
	rt.router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			header := w.Header()
			// Consenti tutte le origini e imposta il Max-Age a 1 come da specifiche
			header.Set("Access-Control-Allow-Origin", "*")
			header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			header.Set("Access-Control-Max-Age", "1")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// -------------------------------------------------------------------------
	// Rotte di base del template (liveness, ecc.)
	// -------------------------------------------------------------------------
	rt.router.GET("/context", rt.wrap(rt.getContextReply))
	rt.router.GET("/liveness", rt.liveness)

	// -------------------------------------------------------------------------
	// Registrazione delle rotte WASAText
	// -------------------------------------------------------------------------

	// -- Login --
	// Il login NON richiede autenticazione: usiamo il wrap() classico
	rt.router.POST("/session", rt.wrap(rt.doLogin))

	// -- Utenti --
	// Da qui in poi, TUTTE le rotte usano wrapAuth() per controllare il Bearer Token
	rt.router.PUT("/users/:userId/username", rt.wrapAuth(rt.setMyUserName))
	rt.router.PUT("/users/:userId/photo", rt.wrapAuth(rt.setMyPhoto))

	// -- Conversazioni --
	rt.router.GET("/users/:userId/conversations", rt.wrapAuth(rt.getMyConversations))
	rt.router.GET("/conversations/:conversationId", rt.wrapAuth(rt.getConversation))

	// -- Messaggi --
	rt.router.POST("/conversations/:conversationId/messages", rt.wrapAuth(rt.sendMessage))
	rt.router.POST("/messages/:messageId/forward", rt.wrapAuth(rt.forwardMessage))
	rt.router.DELETE("/messages/:messageId", rt.wrapAuth(rt.deleteMessage))

	// -- Reazioni (Commenti) --
	rt.router.POST("/messages/:messageId/comments", rt.wrapAuth(rt.commentMessage))
	rt.router.DELETE("/messages/:messageId/comments", rt.wrapAuth(rt.uncommentMessage))

	// -- Gruppi --
	rt.router.POST("/groups/:groupId/members", rt.wrapAuth(rt.addToGroup))
	rt.router.DELETE("/groups/:groupId/members/:userId", rt.wrapAuth(rt.leaveGroup))
	rt.router.PUT("/groups/:groupId/name", rt.wrapAuth(rt.setGroupName))
	rt.router.PUT("/groups/:groupId/photo", rt.wrapAuth(rt.setGroupPhoto))

	// -------------------------------------------------------------------------
	// Middleware CORS per le risposte standard
	// -------------------------------------------------------------------------
	// Restituiamo un handler che aggiunge l'header Allow-Origin a TUTTE le risposte
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		rt.router.ServeHTTP(w, r)
	})
}