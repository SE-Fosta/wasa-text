<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import api from '../services/api';

// Stato dati
const conversations = ref([]);
const activeConversation = ref(null);
const activeMessages = ref([]);
const newMessage = ref("");

// Stato UI
const showCreateGroupModal = ref(false);
const newGroupName = ref("");
const showAddMemberModal = ref(false);
const newMemberId = ref("");

// Info Utente corrente
const userId = localStorage.getItem("token");
const myUsername = localStorage.getItem("username");

// 1. Carica le conversazioni
async function loadConversations() {
    try {
        const res = await api.getConversations(userId);
        conversations.value = res.data || [];
    } catch (e) {
        console.error("Errore loading conv:", e);
    }
}

// 2. Apri chat
async function openChat(conv) {
    activeConversation.value = conv;
    await refreshMessages();
}

// 3. Carica messaggi
async function refreshMessages() {
    if (!activeConversation.value) return;
    try {
        const res = await api.getConversationMessages(activeConversation.value.id);
        activeMessages.value = res.data.messages || [];
    } catch (e) {
        console.error("Errore loading msgs:", e);
    }
}

// 4. Invia Messaggio
async function send() {
    if (!newMessage.value.trim() || !activeConversation.value) return;
    try {
        await api.sendMessage(activeConversation.value.id, newMessage.value);
        newMessage.value = "";
        await refreshMessages();
        await loadConversations();
    } catch (e) {
        alert("Errore invio: " + e.message);
    }
}

// 5. Crea Nuovo Gruppo
async function createGroup() {
    if (!newGroupName.value) return;
    try {
        const newGroupId = "group-" + Date.now();
        await api.setGroupName(newGroupId, newGroupName.value);
        await api.addToGroup(newGroupId, userId);
        alert("Gruppo creato!");
        showCreateGroupModal.value = false;
        newGroupName.value = "";
        await loadConversations();
    } catch (e) {
        alert("Errore creazione gruppo: " + e.message);
    }
}

// 6. Aggiungi Membro al Gruppo
async function addMember() {
    if (!newMemberId.value || !activeConversation.value) return;
    try {
        await api.addToGroup(activeConversation.value.id, newMemberId.value);
        alert("Utente aggiunto!");
        showAddMemberModal.value = false;
        newMemberId.value = "";
    } catch (e) {
        alert("Errore aggiunta membro: " + e.message);
    }
}

// Polling
let polling;
onMounted(() => {
    loadConversations();
    polling = setInterval(() => {
        loadConversations();
        if (activeConversation.value) refreshMessages();
    }, 3000);
});

onUnmounted(() => clearInterval(polling));
</script>

<template>
    <div class="main-layout">
        <div class="sidebar">
            <div class="header">
                <div class="user-info">
                    <div class="avatar-circle">{{ myUsername ? myUsername.charAt(0).toUpperCase() : 'U' }}</div>
                    <span>{{ myUsername }}</span>
                </div>
                <button class="btn-icon" @click="showCreateGroupModal = true" title="Nuovo Gruppo">+</button>
            </div>
            
            <div class="conv-list">
                <div 
                    v-for="conv in conversations" 
                    :key="conv.id" 
                    class="conv-item"
                    :class="{ active: activeConversation?.id === conv.id }"
                    @click="openChat(conv)"
                >
                    <div class="avatar-circle">
                        {{ conv.name ? conv.name.charAt(0).toUpperCase() : '?' }}
                    </div>
                    <div class="conv-info">
                        <h4>{{ conv.name || 'Chat' }}</h4>
                        <p class="preview">{{ conv.lastMessage }}</p>
                    </div>
                </div>
                <div v-if="conversations.length === 0" class="empty-list">
                    Nessuna conversazione. Creane una!
                </div>
            </div>
        </div>

        <div class="chat-area">
            <div v-if="activeConversation" class="chat-window">
                <div class="chat-header">
                    <h3>{{ activeConversation.name }}</h3>
                    <button 
                        v-if="activeConversation.isGroup" 
                        @click="showAddMemberModal = true"
                        class="btn-small"
                    >
                        + Aggiungi Membro
                    </button>
                </div>
                
                <div class="messages-list">
                    <div 
                        v-for="msg in activeMessages" 
                        :key="msg.id"
                        class="message-bubble"
                        :class="{ 'my-msg': msg.senderId === userId }"
                    >
                        <div class="msg-sender" v-if="msg.senderId !== userId">{{ msg.senderName }}</div>
                        <div class="msg-content">{{ msg.content }}</div>
                        <div class="msg-time">{{ new Date(msg.timestamp).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) }}</div>
                    </div>
                </div>

                <div class="input-area">
                    <input 
                        v-model="newMessage" 
                        @keyup.enter="send" 
                        placeholder="Scrivi un messaggio..." 
                    />
                    <button @click="send">Invia</button>
                </div>
            </div>
            
            <div v-else class="empty-state">
                <p>Seleziona una chat o crea un nuovo gruppo per iniziare.</p>
                <p>Il tuo ID utente è: <strong>{{ userId }}</strong></p>
            </div>
        </div>

        <div v-if="showCreateGroupModal" class="modal-overlay">
            <div class="modal">
                <h3>Crea Nuovo Gruppo</h3>
                <input v-model="newGroupName" placeholder="Nome del gruppo" />
                <div class="modal-actions">
                    <button @click="createGroup">Crea</button>
                    <button class="btn-secondary" @click="showCreateGroupModal = false">Annulla</button>
                </div>
            </div>
        </div>

        <div v-if="showAddMemberModal" class="modal-overlay">
            <div class="modal">
                <h3>Aggiungi Utente al Gruppo</h3>
                <p>Inserisci l'ID dell'utente da aggiungere.</p>
                <input v-model="newMemberId" placeholder="User ID (es. 12345...)" />
                <div class="modal-actions">
                    <button @click="addMember">Aggiungi</button>
                    <button class="btn-secondary" @click="showAddMemberModal = false">Annulla</button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* Layout */
.main-layout { display: flex; height: 100vh; background: #f0f2f5; font-family: 'Segoe UI', sans-serif; }
.sidebar { width: 350px; background: white; border-right: 1px solid #ddd; display: flex; flex-direction: column; }
.chat-area { flex: 1; display: flex; flex-direction: column; background: #efe7dd; position: relative; }

/* Header Sidebar */
.header { padding: 15px; background: #f0f2f5; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #ddd; }
.user-info { display: flex; align-items: center; gap: 10px; font-weight: bold; }
.btn-icon { background: none; border: none; font-size: 24px; color: #54656f; cursor: pointer; padding: 0 10px; }
.btn-icon:hover { color: #008069; }

/* Conversation List */
.conv-list { overflow-y: auto; flex: 1; }
.conv-item { display: flex; align-items: center; padding: 12px 15px; cursor: pointer; border-bottom: 1px solid #f5f5f5; }
.conv-item:hover { background: #f5f5f5; }
.conv-item.active { background: #e9edef; }
.avatar-circle { width: 45px; height: 45px; background: #dfe3e5; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 20px; color: white; margin-right: 15px; flex-shrink: 0; }
.conv-info h4 { margin: 0; font-size: 16px; color: #111b21; }
.preview { margin: 3px 0 0; font-size: 13px; color: #667781; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.empty-list { padding: 20px; text-align: center; color: #888; font-size: 14px; }

/* Chat Window */
.chat-window { display: flex; flex-direction: column; height: 100%; }
.chat-header { padding: 10px 20px; background: #f0f2f5; border-bottom: 1px solid #ddd; display: flex; justify-content: space-between; align-items: center; }
.messages-list { flex: 1; padding: 20px; overflow-y: auto; display: flex; flex-direction: column; gap: 8px; }
.message-bubble { max-width: 60%; padding: 8px 12px; border-radius: 8px; background: white; box-shadow: 0 1px 0.5px rgba(0,0,0,0.13); position: relative; font-size: 14px; }
.my-msg { align-self: flex-end; background: #d9fdd3; }
.msg-sender { font-size: 11px; color: #d65c5c; margin-bottom: 2px; font-weight: bold; }
.msg-time { font-size: 10px; color: #667781; text-align: right; margin-top: 4px; }
.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; color: #667781; text-align: center; }

/* Input Area */
.input-area { padding: 10px 15px; background: #f0f2f5; display: flex; gap: 10px; align-items: center; }
.input-area input { flex: 1; padding: 12px; border: none; border-radius: 8px; outline: none; }
button { background: #008069; color: white; border: none; padding: 8px 16px; border-radius: 20px; cursor: pointer; font-weight: 600; }
button:hover { background: #00a884; }
.btn-secondary { background: #e9edef; color: #111b21; }
.btn-small { font-size: 12px; padding: 5px 10px; }

/* Modals */
.modal-overlay { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal { background: white; padding: 25px; border-radius: 10px; width: 300px; box-shadow: 0 4px 10px rgba(0,0,0,0.2); }
.modal input { width: 100%; padding: 8px; margin: 15px 0; border: 1px solid #ddd; border-radius: 5px; box-sizing: border-box; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; }
</style>