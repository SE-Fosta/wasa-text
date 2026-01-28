import axios from "axios";

// Creazione dell'istanza Axios
const api = axios.create({
    baseURL: "http://localhost:3000", // URL del backend Go
    timeout: 10000, // Timeout di 10 secondi
});

// --- INTERCEPTOR PER L'AUTENTICAZIONE ---
// Aggiunge automaticamente l'header "Authorization: Bearer <id>" a ogni richiesta
// come richiesto dalle specifiche.
api.interceptors.request.use(
    (config) => {
        const userId = localStorage.getItem("token"); // Recupera l'ID salvato al login
        if (userId) {
            config.headers["Authorization"] = `Bearer ${userId}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

export default {
    // --- LOGIN ---
    // Endpoint: POST /session [cite: 61, 91]
    login(username) {
        return api.post("/session", { name: username });
    },

    // --- GESTIONE UTENTE ---

    // Endpoint: PUT /users/{userId}/username [cite: 62]
    setMyUserName(userId, newName) {
        return api.put(`/users/${userId}/username`, { name: newName });
    },

    // Endpoint: PUT /users/{userId}/photo [cite: 73]
    // Nota: Richiede multipart/form-data per l'upload del file
    setMyPhoto(userId, photoFile) {
        const formData = new FormData();
        formData.append("photo", photoFile);
        return api.put(`/users/${userId}/photo`, formData, {
            headers: { "Content-Type": "multipart/form-data" },
        });
    },

    // Endpoint: GET /users/{userId}/conversations [cite: 63]
    getConversations(userId) {
        return api.get(`/users/${userId}/conversations`);
    },

    // --- GESTIONE CONVERSAZIONI ---

    // Endpoint: GET /conversations/{conversationId} [cite: 64]
    // Recupera i messaggi e i partecipanti
    getConversationMessages(conversationId) {
        return api.get(`/conversations/${conversationId}`);
    },

    // Endpoint: POST /conversations/{conversationId}/messages [cite: 65]
    // Gestisce sia messaggi di testo che foto
    sendMessage(conversationId, contentOrFile, isPhoto = false, replyTo = null) {
        if (isPhoto) {
            // Invio Foto (Multipart)
            const formData = new FormData();
            formData.append("photo", contentOrFile); // contentOrFile qui è il File object
            formData.append("messageType", "photo");
            if (replyTo) formData.append("replyTo", replyTo);

            return api.post(`/conversations/${conversationId}/messages`, formData, {
                headers: { "Content-Type": "multipart/form-data" },
            });
        } else {
            // Invio Testo (JSON)
            const payload = {
                content: contentOrFile, // contentOrFile qui è una stringa
                messageType: "text",
            };
            if (replyTo) payload.replyTo = replyTo;

            return api.post(`/conversations/${conversationId}/messages`, payload);
        }
    },

    // --- GESTIONE MESSAGGI ---

    // Endpoint: POST /messages/{messageId}/forward [cite: 66]
    forwardMessage(messageId, targetConversationId) {
        return api.post(`/messages/${messageId}/forward`, {
            targetConversationId: targetConversationId,
        });
    },

    // Endpoint: POST /messages/{messageId}/comments [cite: 67]
    // Aggiunge una reazione (emoji)
    commentMessage(messageId, emoji) {
        return api.post(`/messages/${messageId}/comments`, { emoji: emoji });
    },

    // Endpoint: DELETE /messages/{messageId}/comments [cite: 68]
    // Rimuove la reazione dell'utente corrente
    uncommentMessage(messageId) {
        return api.delete(`/messages/${messageId}/comments`);
    },

    // Endpoint: DELETE /messages/{messageId} [cite: 69]
    deleteMessage(messageId) {
        return api.delete(`/messages/${messageId}`);
    },

    // --- GESTIONE GRUPPI ---

    // Endpoint: PUT /groups/{groupId}/name [cite: 72]
    // Usato per rinominare un gruppo O creare un nuovo gruppo (se l'ID è nuovo)
    setGroupName(groupId, name) {
        return api.put(`/groups/${groupId}/name`, { name: name });
    },

    // Endpoint: PUT /groups/{groupId}/photo [cite: 74]
    setGroupPhoto(groupId, photoFile) {
        const formData = new FormData();
        formData.append("photo", photoFile);
        return api.put(`/groups/${groupId}/photo`, formData, {
            headers: { "Content-Type": "multipart/form-data" },
        });
    },

    // Endpoint: POST /groups/{groupId}/members [cite: 70]
    addToGroup(groupId, userIdToAdd) {
        return api.post(`/groups/${groupId}/members`, { userId: userIdToAdd });
    },

    // Endpoint: DELETE /groups/{groupId}/members/{userId} [cite: 71]
    // Usato per uscire dal gruppo ("leaveGroup")
    leaveGroup(groupId, userId) {
        return api.delete(`/groups/${groupId}/members/${userId}`);
    },
};
