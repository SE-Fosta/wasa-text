<script setup>
import { ref, watch, onMounted , nextTick} from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import api from './services/axios.js'

const route = useRoute();
const router = useRouter();

const userId = ref(localStorage.getItem('token'));
const username = ref(localStorage.getItem('username') || '');

const chats = ref([]);
const chatContainer = ref(null);

let pollingTimer = null;

// Variabili per il profilo
const userPhotoUrl = ref(localStorage.getItem('photoUrl') || ''); 
const isProfileModalOpen = ref(false);
const newUsername = ref('');
const profilePhotoFile = ref(null);
const profilePhotoPreview = ref(null);
const profileError = ref('');

// Variabili per la ricerca
const searchQuery = ref('');
const searchResults = ref([]);
const isSearchFocused = ref(false);

// Variabili per la chat
const activeChatId = ref(null);
const messageText = ref('');
const messages = ref([]);
const forwardingMessageId = ref(null);
const emojis = ['👍', '❤️', '😂', '😮', '😢', '🙏'];
const activeReactionMessageId = ref(null);
const activeDropdownMessageId = ref(null);
const replyingToMessage = ref(null);

// Variabili per il gruppo
const isGroupModalOpen = ref(false);
const newGroupName = ref('');
const selectedUsers = ref([]);
const groupPhotoFile = ref(null);
const groupPhotoPreview = ref(null);
const groupError = ref('');
const allUsers = ref([]);
const isGroupInfoModalOpen = ref(false);
const editingGroupName = ref('');
const editingGroupPhotoPreview = ref(null);
const groupMembersList = ref([]); 
const userToAdd = ref('');
const fileInput = ref(null);
const selectedPhoto = ref(null);
const groupInfoError = ref('');
const groupAddSearchQuery = ref('');
const groupAddSearchResults = ref([]);
const selectedUsersToAdd = ref([]);
const isGroupAddSearchFocused = ref(false);
const editingGroupPhotoFile = ref(null);

const doLogout = () => {
    // Cancella i dati salvati dell'utenye
    localStorage.removeItem('token');
    localStorage.removeItem('username');
	localStorage.removeItem('userId');
    localStorage.removeItem('photoUrl');

    // Resetta le variabili a schermo 
    username.value = '';
    userId.value = null;
    chats.value = [];
    activeChatId.value = null;

    if (pollingTimer) {
        clearInterval(pollingTimer);
        pollingTimer = null;
    }

    router.push('/login');
};

const updateData = async () => {
    // Carica i dati dell'utente
    userId.value = localStorage.getItem('token');
    username.value = localStorage.getItem('username') || '';
    userPhotoUrl.value = localStorage.getItem('photoUrl') || '';

    // Carica le chat dell'utente
    if (userId.value) {
        try {
            let response = await api.get(`/users/${userId.value}/conversations`);
            chats.value = response.data || [];
        } catch (e) {
            console.error("Errore nel recupero delle chat:", e);
        }
    } else {
        chats.value = [];
    }

    searchUsers();
};

const formatTime = (timestamp) => {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};

const truncateText = (text, maxLength) => {
    if (!text) return '';
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
};

const scrollToBottom = () => {
    if (chatContainer.value) {
        chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
    }
};

const scrollToMessage = (msgId) => {
    const element = document.getElementById(`msg-${msgId}`);
    if (element && chatContainer.value) {
        // block: 'center' mette il messaggio al centro dello schermo così è ben visibile
        element.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
};

const markAsRead = async (conversationId) => {
    try {
        await api.put(`/conversations/${conversationId}/read`);
    } catch (e) {
        console.warn("Impossibile segnare i messaggi come letti:", e);
    }
};

const loadMessages = async () => {
    if (!activeChatId.value) return;
    
    try {
        // Recupera i messaggi dal server
        const response = await api.get(`/conversations/${activeChatId.value}/messages`);
        const newMessages = response.data || [];
        
        // 1. Troviamo il primo messaggio non letto (inviato da altri)
        // Dobbiamo farlo ORA, prima che markAsRead() aggiorni lo stato!
        const firstUnreadMsg = newMessages.find(m => !m.read && m.senderId != userId.value);
        
        // Aggiorna la spunta sul database
        await markAsRead(activeChatId.value);

        // Evita di aggiornare la chat mentre si legge un messaggio
        let isAtBottom = true;
        if (chatContainer.value) {
            const { scrollTop, scrollHeight, clientHeight } = chatContainer.value;
            if (scrollHeight - scrollTop - clientHeight > 100) {
                isAtBottom = false;
            }
        }

        const isFirstLoad = messages.value.length === 0;

        // Aggiorna i messaggi a schermo
        messages.value = newMessages;
        
        // Carica i messaggi dal non letto o dal basso se non si hanno messaggi non letti
        nextTick(() => {
            if (isFirstLoad) {
                if (firstUnreadMsg) {
                    scrollToMessage(firstUnreadMsg.id);
                } else {
                    scrollToBottom();
                }
            } else if (isAtBottom) {
                scrollToBottom();
            }
        });

    } catch (e) {
        console.error("Errore nel caricamento dei messaggi:", e);
    }
};


// Carica messaggi della chat selezionata
watch(activeChatId, (newVal) => {
    if (pollingTimer) {
        clearInterval(pollingTimer);
        pollingTimer = null;
    }

    if (newVal) {
        loadMessages();
        
        pollingTimer = setInterval(() => {
            loadMessages();
        }, 2000);

    } else {
        messages.value = [];
    }
});

//Carica tutti gli utenti o filtrati tramite l'iniziale 
const searchUsers = async () => {
    try {
        const response = await api.get('/users', { params: { username: searchQuery.value } });
        
        if (response.data) {
            searchResults.value = response.data.filter(u => u.username !== username.value);
        } else {
            searchResults.value = [];
        }
    } catch (e) {
        console.error("Errore nella ricerca utenti:", e);
    }
};

//
const startChat = async (selectedUser) => {
    isSearchFocused.value = false;
    searchQuery.value = '';
    
    // Controllo di sicurezza
    if (!selectedUser || !selectedUser.id) {
        console.error("Errore: l'utente selezionato non ha un ID valido.", selectedUser);
        alert("Impossibile avviare la chat, utente non valido.");
        return;
    }

    try {
        const payload = {
            targetUserId: String(selectedUser.id),
            isGroup: false,
            name: "" 
        };

        const response = await api.post(`/users/${userId.value}/conversations`, payload);
        
        // Aggiorna la sidebar a sinistra
        await updateData(); 
        
        // Prendiamo l'ID della conversazione 
        const convId = response.data.conversationId;

        // Cerca la chat appena caricata in chats.value
        const chatIndex = chats.value.findIndex(c => (c.id || c.conversationId) === convId);
        
        // Crea la chat
        chats.value.unshift({
            id: convId,
            name: selectedUser.username,
            photoUrl: selectedUser.photoUrl || null,
            isGroup: false,
            unreadCount: 0,
            lastActivity: new Date().toISOString()
        });
    

        activeChatId.value = convId; 
        
    } catch (e) {
        console.error("Errore durante la creazione della chat:", e.response?.data || e);
    }
};

const triggerFileInput = () => {
    fileInput.value.click();
};

const handlePhotoSelected = (event) => {
    const file = event.target.files[0];
    if (!file) return;

    // Salviamo il file selezionato in memoria, ma NON lo inviamo ancora!
    selectedPhoto.value = file;
};

//Gestisce la foto per esser inviata come messaggio
const sendPhoto = async (caption = '', replyToId = null) => {
    if (!selectedPhoto.value || !activeChatId.value || !userId.value) return;

    try {
        const formData = new FormData();
        formData.append("photo", selectedPhoto.value);
        
        formData.append("messageType", "photo");

        if (caption && caption.trim() !== '') {
            formData.append("content", caption.trim());
        }
        if (replyToId) {
            formData.append("replyTo", replyToId);
        }
        
        await api.post(`/conversations/${activeChatId.value}/messages`, formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        });

        console.log("Foto inviata con successo!");
        selectedPhoto.value = null; 
        
        if (fileInput.value) fileInput.value.value = '';
        
        await loadMessages(); 
        
    } catch (e) {
        console.error("Errore nell'invio della foto:", e);
    }
};

// Crea il messaggio da inviare
const sendMessage = async () => {
    // Invia foto e se presente un testo
    if (selectedPhoto.value) {
        await sendPhoto(
            messageText.value, 
            replyingToMessage.value ? replyingToMessage.value.id : undefined
        );
        
        // Svuota i campi dopo aver inviato il messaggio
        messageText.value = ''; 
        replyingToMessage.value = null;
        return;
    }

    // Invia testo
    if (!messageText.value.trim() || !activeChatId.value || !userId.value) return;

    //Crea il messaggio nel backend
    try {
        await api.post(`/conversations/${activeChatId.value}/messages`, {
            content: messageText.value,
            messageType: "text",
            replyTo: replyingToMessage.value ? replyingToMessage.value.id : undefined
        });

        messageText.value = ''; 
        replyingToMessage.value = null;

        await loadMessages(); 
        await updateData(); 
    } catch (e) {
        console.error("Errore nell'invio del messaggio di testo:", e);
    }
};

// Elimina un messaggio
const deleteMessage = async (messageId) => {
    // Chiediamo conferma per sicurezza!
    if (!confirm("Vuoi davvero eliminare questo messaggio?")) return;
    
    try {
        await api.delete(`/messages/${messageId}`);
        // Ricarichiamo la chat per far sparire il fumetto
        await loadMessages();
    } catch (e) {
        console.error("Errore durante l'eliminazione del messaggio:", e);
    }
};

const openForwardModal = (messageId) => {
    forwardingMessageId.value = messageId;
};

const cancelForward = () => {
    forwardingMessageId.value = null;
};

// Inoltra il messaggio
const forwardMessage = async (targetChat) => {
    if (!forwardingMessageId.value || !targetChat) return;

    const targetId = targetChat.id || targetChat.conversationId;

    try {
        await api.post(`/messages/${forwardingMessageId.value}/forward`, {
            targetConversationId: targetId
        });
        
        console.log("Messaggio inoltrato con successo!");
        forwardingMessageId.value = null; // Chiudi il popup
        
    } catch (e) {
        console.error("Errore durante l'inoltro:", e);
    }
};

// Mostra o nasconde il menu delle reaction
const toggleReactionMenu = (messageId) => {
    if (activeReactionMessageId.value === messageId) {
        activeReactionMessageId.value = null;
    } else {
        activeReactionMessageId.value = messageId;
    }
};

// Gestisce l'aggiunta, rimozione o modifica della reaction
const toggleReaction = async (messageId, emoji) => {
    console.log("1. Hai cliccato l'emoji:", emoji, "sul messaggio:", messageId);

    // Chiude la tendina
    activeReactionMessageId.value = null;

    const msg = messages.value.find(m => m.id === messageId);

    if (!msg) {
        console.error("ERRORE: Non trovo il messaggio! Forse messageId è undefined?");
        return; 
    }

    // Controlla se l'utente loggato ha già una reazione
    const myExistingReaction = msg.reactions?.find(r => r.userId == userId.value);

    // Aggiunge o modifica una reaction
    try {
        if (myExistingReaction && myExistingReaction.emoji === emoji) {
            await api.delete(`/messages/${messageId}/reactions`);
        } else {
            await api.post(`/messages/${messageId}/reactions`, { emoji: emoji });
        }
        
        await loadMessages();
    } catch (e) {
        console.error("ERRORE DURANTE LA CHIAMATA API:", e);
    }
};

// Apre o chiude la tendina delle reaction
const toggleDropdown = (messageId) => {
    if (activeDropdownMessageId.value === messageId) {
        activeDropdownMessageId.value = null;
    } else {
        activeDropdownMessageId.value = messageId; // La apre
        activeReactionMessageId.value = null; 
    }
};

const cancelReply = () => {
    replyingToMessage.value = null;
};

const getRepliedMessage = (replyToId) => {
    if (!replyToId) return null;
    return messages.value.find(m => m.id === replyToId) || null;
};

const handleAction = (action, msgId) => {
    activeDropdownMessageId.value = null; // Chiude la tendina
    
    if (action === 'react') {
        activeReactionMessageId.value = msgId;
    } else if (action === 'forward') {
        openForwardModal(msgId);
    } else if (action === 'delete') {
        deleteMessage(msgId);
    } else if (action === 'reply') {
        // Troviamo il messaggio intero e lo salviamo nella variabile
        const msg = messages.value.find(m => m.id === msgId);
        replyingToMessage.value = msg;
    }
};

onMounted(() => {
    updateData();
});

watch(() => route.path, () => {
    updateData();
});

// "Ascoltiamo" la barra di ricerca
watch(searchQuery, () => {
    searchUsers();
});

// Funzione per nascondere la tendina quando clicchi fuori
const hideDropdown = () => {
    setTimeout(() => {
        isSearchFocused.value = false;
    }, 200); 
};

const openProfileModal = () => {
    newUsername.value = username.value; 
    profilePhotoPreview.value = userPhotoUrl.value;
    profileError.value = '';
    isProfileModalOpen.value = true;
};

const closeProfileModal = () => {
    isProfileModalOpen.value = false;
    profilePhotoFile.value = null;
    profilePhotoPreview.value = null;
};

const getImageUrl = (path) => {
    if (!path) return '';

    if (path.startsWith('blob:') || path.startsWith('data:') || path.startsWith('http')) {
        return path;
    }

    let baseUrl = api.defaults.baseURL || '';
    // Rimuove /api alla fine
    baseUrl = baseUrl.replace(/\/api\/?$/, '');

    if (!baseUrl) {
        baseUrl = window.location.origin;
    }

    const cleanPath = path.startsWith('/') ? path : '/' + path;
    return `${baseUrl}${cleanPath}`;
};

// Gestisce la nuova foto profilo selezionata
const handleProfilePhotoSelected = (event) => {
    const file = event.target.files[0];
    if (file) {
        profilePhotoFile.value = file;
        // Crea un'anteprima temporanea per farla vedere subito all'utente
        profilePhotoPreview.value = URL.createObjectURL(file); 
    }
};

const saveProfile = async () => {
    profileError.value = '';
    let needsUpdate = false;

    try {
        // Cambia il nome utente
        if (newUsername.value.trim() !== username.value && newUsername.value.trim() !== '') {
            await api.put(`/users/${userId.value}/username`, 
                { username: newUsername.value.trim() },
                { headers: { 'Authorization': `Bearer ${userId.value}` } } 
            );
            
            username.value = newUsername.value.trim();
            localStorage.setItem('username', username.value);
            needsUpdate = true;
        }

        // Cambia la foto profilo
        if (profilePhotoFile.value) {
            const formData = new FormData();
            formData.append("photo", profilePhotoFile.value);
            
            await api.put(`/users/${userId.value}/photo`, formData, {
                headers: { 
                    'Content-Type': 'multipart/form-data',
                    'Authorization': `Bearer ${userId.value}` 
                }
            });

            const newPath = `/uploads/profile_${userId.value}.jpg`;
            userPhotoUrl.value = `${newPath}?t=${new Date().getTime()}`;
            localStorage.setItem('photoUrl', userPhotoUrl.value);
            needsUpdate = true;
        }

        if (needsUpdate) {
            await updateData(); 
        }

        closeProfileModal();

    } catch (e) {
        if (e.response && e.response.status === 409) {
            profileError.value = "Questo nome utente è già in uso.";
        } else if (e.response && e.response.status === 401) {
            profileError.value = "Sessione scaduta. Effettua di nuovo il login.";
        } else {
            profileError.value = "Errore durante il salvataggio.";
        }
        console.error("Errore profilo:", e);
    }
};

const openGroupModal = async () => {
    newGroupName.value = '';
    selectedUsers.value = [];
    groupPhotoFile.value = null;
    groupPhotoPreview.value = null;
    groupError.value = '';
    
    try {
        const response = await api.get('/users');
        allUsers.value = response.data.filter(u => u.username !== username.value);
        isGroupModalOpen.value = true;
    } catch (e) {
        console.error("Errore caricamento utenti per gruppo:", e);
    }
};

const closeGroupModal = () => { isGroupModalOpen.value = false; };

const handleGroupPhotoSelected = (event) => {
    const file = event.target.files[0];
    if (file) {
        groupPhotoFile.value = file;
        groupPhotoPreview.value = URL.createObjectURL(file);
    }
};

// Crea il gruppo
const createGroup = async () => {
    if (!newGroupName.value.trim() || selectedUsers.value.length === 0) {
        groupError.value = "Inserisci un nome e seleziona almeno un membro.";
        return;
    }

    try {
        // Crea la istanza
        const payload = {
            isGroup: true,
            name: newGroupName.value.trim(),
            targetUserId: "" 
        };
        const response = await api.post(`/users/${userId.value}/conversations`, payload);
        const newGroupId = response.data.conversationId;

        // Aggiunge i membri al gruppo
        const memberPromises = selectedUsers.value.map(uId => {
            return api.post(`/groups/${newGroupId}/members`, { userId: String(uId) });
        });

        // Aspetta che tutti i membri siano aggiunti
        await Promise.all(memberPromises); 
        
        await updateData(); 
        
        // Crea il gruppo
        chats.value.unshift({
            id: newGroupId,
            name: newGroupName.value.trim(), 
            photoUrl: null, 
            isGroup: true,
            unreadCount: 0,
            lastActivity: new Date().toISOString()
        });

        activeChatId.value = newGroupId; 
        closeGroupModal();

    } catch (e) {
        console.error("Errore creazione gruppo:", e.response?.data || e);
        groupError.value = "Errore durante la creazione del gruppo.";
    }
};

const addTenUsersModal = async () => {
    const usersToAdd = [
        "Giuditta", "Franco", "Noyz", "Gianfranco", "Forte", 
        "Giuno", "Henry", "Saverio", "Eugenio", "Rita"
    ];
    
    let successCount = 0;
    
    console.log("Inizio l'aggiunta automatica di 10 utenti...");

    for (const name of usersToAdd) {
        try {
            await api.post('session', { name: name });
            successCount++;
            console.log(`Utente creato: ${name}`);
        } catch (e) {
            console.error(`Errore nella creazione di ${name}:`, e);
        }
    }
};

const getActiveChat = () => {
    return chats.value.find(c => (c.id || c.conversationId) === activeChatId.value) || {};
};



// Funzioni per aprire/chiudere la modale
const openGroupInfo = async () => {
    const chat = getActiveChat();
    const cId = chat.id || chat.conversationId;

    if (chat && chat.isGroup) {
        try {
            // Carica i dati del gruppo
            const response = await api.get(`/conversations/${cId}`);
            const groupData = response.data;

            groupMembersList.value = groupData.members || [];
            
            editingGroupName.value = groupData.name || '';
            editingGroupPhotoPreview.value = groupData.photoUrl || null;
            
            groupAddSearchQuery.value = ''; 
            selectedUsersToAdd.value = [];
            
            // Carica la lista filtrata
            await searchUsersForGroup();
            
            isGroupInfoModalOpen.value = true;
        } catch (e) {
            console.error("Errore nel caricamento dettagli gruppo:", e);
            alert("Impossibile caricare le informazioni del gruppo.");
        }
    }
};

const closeGroupInfo = () => {
    isGroupInfoModalOpen.value = false;
};

const saveGroupName = async () => {
    groupInfoError.value = '';
    const newName = editingGroupName.value.trim();

    if (!newName) {
        groupInfoError.value = "Il nome del gruppo non può essere vuoto.";
        return;
    }

    try {
        // Chiamata al backend per aggiornare il nome
        await api.put(`/groups/${activeChatId.value}/name`, { 
            name: newName 
        });

        // Aggiorniamo la lista delle chat locale per riflettere il cambio
        await updateData();
        
    } catch (e) {
        console.error("Errore nel salvataggio del nome del gruppo:", e);
        groupInfoError.value = "Impossibile aggiornare il nome. Riprova.";
    }
};

// Funzione che cerca gli utenti per il gruppo
const searchUsersForGroup = async () => {
    try {
        const params = groupAddSearchQuery.value.trim() ? { username: groupAddSearchQuery.value } : {};
        const response = await api.get('/users', { params });
        
        if (response.data) {
            const currentMemberIds = new Set(groupMembersList.value.map(m => String(m.id)));
            
            // Filtra la lista globale degli utenti
            groupAddSearchResults.value = response.data.filter(u => {
                const isMe = String(u.id) === String(userId.value);
                const isAlreadyInGroup = currentMemberIds.has(String(u.id));
                
                return !isMe && !isAlreadyInGroup;
            });
        }
    } catch (e) {
        console.error("Errore recupero utenti per gruppo:", e);
    }
};

// Guarda la variabile di ricerca: appena l'utente digita, fa partire la ricerca
watch(groupAddSearchQuery, () => {
    searchUsersForGroup();
});

// Aggiunge i membri al gruppo
const addMembersToGroup = async () => {
    if (selectedUsersToAdd.value.length === 0) return;
    
    groupInfoError.value = '';

    try {
        const memberPromises = selectedUsersToAdd.value.map(uId => {
            return api.post(`/groups/${activeChatId.value}/members`, { userId: String(uId) });
        });
        
        await Promise.all(memberPromises); 
        
        // Pulisce l'interfaccia dopo il successo
        selectedUsersToAdd.value = [];
        groupAddSearchQuery.value = '';
        groupAddSearchResults.value = [];
        
        // Aggiorna le chat a sinistra
        await updateData(); 
        
    } catch (e) {
        groupInfoError.value = "Impossibile aggiungere i partecipanti selezionati.";
    }
};
const hideGroupAddDropdown = () => {
    setTimeout(() => {
        isGroupAddSearchFocused.value = false;
    }, 200);
};

// Rimuove un utente dal gruppo
const removeMemberFromGroup = async (memberId) => {
    const isMe = String(memberId) === String(userId.value);

    // Messaggio di conferma
    const confirmMessage = isMe 
        ? "Sei sicuro di voler abbandonare questo gruppo?" 
        : "Sei sicuro di voler rimuovere questo utente dal gruppo?";

    if (!confirm(confirmMessage)) {
        return;
    }

    try {
        await api.delete(`/groups/${activeChatId.value}/members/${memberId}`);

        if (isMe) {
            closeGroupInfo();
            activeChatId.value = null;
        } else {
            groupMembersList.value = groupMembersList.value.filter(m => m.id !== memberId);
            await searchUsersForGroup(); 
        }

        await updateData(); 

    } catch (e) {
        console.error("Errore durante la rimozione/abbandono:", e);
    }
};

const handleEditingGroupPhotoSelected = (event) => {
    const file = event.target.files[0];
    if (file) {
        editingGroupPhotoFile.value = file;
        editingGroupPhotoPreview.value = URL.createObjectURL(file);
    }
};

const saveGroupPhoto = async () => {
    if (!editingGroupPhotoFile.value) return;

    groupInfoError.value = '';
    try {
        const formData = new FormData();
        formData.append("photo", editingGroupPhotoFile.value);
        
        await api.put(`/groups/${activeChatId.value}/photo`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        });

        editingGroupPhotoFile.value = null; 
        
        await updateData(); 
    } catch (e) {
        console.error("Errore salvataggio foto gruppo:", e);
        groupInfoError.value = "Impossibile aggiornare la foto del gruppo.";
    }
};
</script>

<template>
	<header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
		<a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6 d-flex align-items-center" href="#/">
			WASAText
			<span v-if="username" class="ms-2 fw-normal text-white-50">
				| @{{ username }}
			</span>
		</a>
		
		<div class="navbar-nav ms-auto me-3" v-if="username">
			<button class="btn btn-sm btn-outline-light d-flex align-items-center" @click="doLogout">
				<svg class="feather me-1" style="width: 14px; height: 14px;"><use href="/feather-sprite-v4.29.0.svg#log-out"/></svg>
				Logout
			</button>
		</div>

		<button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
			<span class="navbar-toggler-icon"></span>
		</button>
	</header>

	<div class="container-fluid">
		<div class="row">
			<nav v-if="userId" id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
				<div class="position-sticky pt-3 sidebar-sticky d-flex flex-column" style="height: calc(100vh - 48px);">
					
					<div class="flex-grow-1 overflow-auto pb-3">
						<div class="px-3 mb-4 mt-2 position-relative">
							<label class="form-label small text-muted text-uppercase fw-bold">Cerca Utenti</label>
							
							<input 
								type="text" 
								v-model="searchQuery" 
								@focus="isSearchFocused = true"
								@blur="isSearchFocused = false"
								class="form-control form-control-sm" 
								placeholder="Scrivi un nome..."
							/>
							
							<ul 
								class="list-group position-absolute w-100 shadow mt-1" 
								style="z-index: 1050; left: 0; padding-left: 1rem; padding-right: 1rem;" 
								v-if="isSearchFocused && searchResults.length > 0"
							>
								<li v-for="user in searchResults" :key="user.id" class="list-group-item list-group-item-action d-flex justify-content-between align-items-center small px-2 py-2">
									<span class="text-truncate fw-medium">@{{ user.username }}</span>
									<button 
										@mousedown.prevent="startChat(user)" 
										class="btn btn-sm btn-outline-primary py-0 px-2" 
										style="font-size: 0.75rem;"
									>
										Chat
									</button>
								</li>
							</ul>
							
							<div 
								v-else-if="isSearchFocused && searchQuery.length > 0 && searchResults.length === 0" 
								class="position-absolute w-100 bg-white border rounded shadow-sm p-2 text-muted small mt-1"
								style="z-index: 1050; left: 0; margin-left: 1rem; width: calc(100% - 2rem) !important;"
							>
								Nessun utente trovato.
							</div>
						</div>

						<hr class="mx-3">

						<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-3 mb-1 text-muted text-uppercase">
							<span>Le Mie Chat</span>
						</h6>
						
						<ul class="nav flex-column mb-2 mt-2">
							<li class="nav-item px-3" v-if="chats.length === 0">
								<small class="text-muted">Nessuna chat presente.</small>
							</li>
							
							<li class="nav-item border-bottom" v-for="(chat, index) in chats" :key="index">
                                <a class="nav-link py-2 d-flex align-items-center w-100" href="#" @click.prevent="activeChatId = chat.id || chat.conversationId">
                                    
                                    <div class="me-3 flex-shrink-0">
                                        <img v-if="chat.photoUrl" 
                                            :src="getImageUrl(chat.photoUrl)" 
                                            class="rounded-circle border shadow-sm" 
                                            style="width: 38px; height: 38px; object-fit: cover;"
                                            alt="Avatar">
                                        
                                        <div v-else 
                                            class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center fw-bold shadow-sm" 
                                            style="width: 38px; height: 38px; font-size: 0.9rem; text-transform: uppercase;">
                                            {{ chat.name ? chat.name.charAt(0) : '?' }}
                                        </div>
                                    </div>

                                    <div class="flex-grow-1 min-width-0 overflow-hidden">
                                        
                                        <div class="d-flex justify-content-between align-items-center">
                                            <strong class="text-dark text-truncate d-inline-block" style="font-size: 0.95rem; max-width: 70%;">
                                                {{ chat.name || 'Chat #' + (chat.id || chat.conversationId) }}
                                            </strong>
                                            
                                            <small v-if="chat.lastActivity" class="text-muted text-nowrap ms-2 flex-shrink-0" style="font-size: 0.7rem;">
                                                {{ formatTime(chat.lastActivity) }}
                                            </small>
                                        </div>
                                        
                                        <div class="d-flex justify-content-between align-items-center">
                                            <div class="text-muted text-truncate flex-grow-1" style="font-size: 0.8rem;">
                                                <span v-if="chat.lastMessage && chat.lastMessage.content">
                                                    <span v-if="chat.isGroup && chat.lastMessage.senderName" class="fw-bold text-dark">
                                                        {{ chat.lastMessage.senderName }}: 
                                                    </span>
                                                    {{ chat.lastMessage.content }}
                                                </span>
                                                <span v-else class="fst-italic">Nessun messaggio</span>
                                            </div>
                                            
                                            <span v-if="chat.unreadCount > 0" class="badge bg-primary rounded-pill ms-2 flex-shrink-0" style="font-size: 0.65rem;">
                                                {{ chat.unreadCount }}
                                            </span>
                                        </div>
                                    </div>
                                    
                                </a>
                            </li>
						</ul>
					</div>
					
					<div class="mt-auto border-top p-3 w-100 bg-white" style="z-index: 10;">
						<div class="d-flex justify-content-center gap-3">
							
							<button 
                                @click="openProfileModal" 
                                class="btn rounded-circle p-0 d-flex align-items-center justify-content-center shadow-sm border overflow-hidden"
                                style="width: 42px; height: 42px; background-color: #f8f9fa;"
                                :title="'Modifica profilo di @' + username"
                            >
                                <img 
                                    v-if="userPhotoUrl" 
                                    :src="getImageUrl(userPhotoUrl)"
                                    class="w-100 h-100" 
                                    style="object-fit: cover;"
                                />
                                
                                <div 
                                    v-else
                                    class="w-100 h-100 bg-primary text-white d-flex align-items-center justify-content-center fw-bold" 
                                    style="font-size: 1.2rem; text-transform: uppercase;"
                                >
                                    {{ username ? username.charAt(0) : '?' }}
                                </div>
                            </button>

							<button 
								@click="openGroupModal" 
								class="btn btn-outline-primary rounded-circle d-flex align-items-center justify-content-center p-0"
								style="width: 42px; height: 42px;"
								title="Crea un nuovo gruppo"
							>
								<svg class="feather" style="width: 18px; height: 18px;"><use href="/feather-sprite-v4.29.0.svg#users"/></svg>
							</button>

							<button 
								@click="addTenUsersModal" 
								class="btn btn-outline-dark rounded-circle d-flex align-items-center justify-content-center p-0"
								style="width: 42px; height: 42px;"
								title="Genera utenti di test"
							>
								<svg class="feather" style="width: 18px; height: 18px;"><use href="/feather-sprite-v4.29.0.svg#user-plus"/></svg>
							</button>

						</div>
					</div>

				</div>
			</nav>

			<main :class="userId ? 'col-md-9 ms-sm-auto col-lg-10' : 'col-12'" class="px-md-4 mt-3" style="height: 100vh; overflow-y: auto;">
				
				<div v-if="activeChatId" class="d-flex flex-column h-100">
                    <div class="d-flex justify-content-between align-items-center pt-3 pb-2 mb-3 border-bottom">
                        <div class="d-flex align-items-center">
                            
                            <div class="me-3">
                                <img v-if="getActiveChat().photoUrl" 
                                    :src="getImageUrl(getActiveChat().photoUrl)" 
                                    class="rounded-circle border shadow-sm" 
                                    style="width: 45px; height: 45px; object-fit: cover;">
                                
                                <div v-else 
                                    class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center fw-bold shadow-sm" 
                                    style="width: 45px; height: 45px; font-size: 1.2rem;">
                                    {{ getActiveChat().name ? getActiveChat().name.charAt(0).toUpperCase() : '?' }}
                                </div>
                            </div>

                            <h1 
                                class="h3 mb-0 d-flex align-items-center text-dark" 
                                :style="getActiveChat().isGroup ? 'cursor: pointer; transition: 0.2s;' : ''"
                                @click="getActiveChat().isGroup ? openGroupInfo() : null"
                                :title="getActiveChat().isGroup ? 'Clicca per info gruppo' : ''"
                                onmouseover="this.style.opacity='0.7'"
                                onmouseout="this.style.opacity='1'"
                            >
                                {{ getActiveChat().name || 'Chat #' + activeChatId }}
                            </h1>
                        </div>

                        <button class="btn btn-sm btn-outline-danger" @click="activeChatId = null">Chiudi Chat</button>
                    </div>
                    
                    <div ref="chatContainer" class="flex-grow-1 p-3 bg-white border rounded overflow-auto d-flex flex-column" style="max-height: calc(100vh - 200px);">

                        <div v-if="messages.length === 0" class="text-center text-muted mt-5">
                            Nessun messaggio. Scrivi qualcosa per rompere il ghiaccio!
                        </div>

                        <div 
                            v-for="(msg, index) in messages" 
                            :key="index" 
                            :id="'msg-' + msg.id" 
                            class="mb-3 d-flex flex-column w-100 position-relative"
                            :class="msg.senderId == userId ? 'align-items-end' : 'align-items-start'"
                            :style="{ zIndex: (activeDropdownMessageId === msg.id || activeReactionMessageId === msg.id) ? 1050 : 1 }"
                        >
                            <div class="d-flex align-items-end" style="max-width: 85%;">

                                <div v-if="msg.senderId != userId && getActiveChat().isGroup" class="me-2 mb-1 flex-shrink-0 align-self-end">
                                    <img v-if="msg.senderPhotoUrl" :src="getImageUrl(msg.senderPhotoUrl)" class="rounded-circle border shadow-sm" style="width: 32px; height: 32px; object-fit: cover;">
                                    <div v-else class="rounded-circle bg-secondary bg-opacity-25 border shadow-sm d-flex align-items-center justify-content-center text-dark fw-bold" style="width: 32px; height: 32px; font-size: 0.85rem;">
                                        {{ msg.senderName ? msg.senderName.charAt(0).toUpperCase() : 'U' }}
                                    </div>
                                </div>

                                <div 
                                    class="p-2 rounded shadow-sm text-break position-relative pe-4" 
                                    :class="msg.senderId == userId ? 'bg-primary text-white ms-auto' : 'bg-light text-dark border'"
                                    style="min-width: 120px; width: fit-content; text-align: left;"
                                >

                                    <div v-if="msg.senderId != userId && getActiveChat().isGroup" class="fw-bold mb-1" style="font-size: 0.75rem; color: #0d6efd;">
                                        ~ {{ msg.senderName }}
                                    </div>

                                    <div v-if="msg.messageType === 'photo' || msg.photoUrl" class="mb-1">
                                        <img :src="msg.photoUrl" class="img-fluid rounded" alt="Foto" style="max-height: 200px; object-fit: cover;">
                                    </div>

                                    <div 
                                        v-if="msg.replyTo && getRepliedMessage(msg.replyTo)" 
                                        class="mb-2 p-2 rounded border-start border-primary border-3 text-start"
                                        :class="msg.senderId == userId ? 'bg-light text-dark bg-opacity-75' : 'bg-secondary bg-opacity-10'"
                                        style="font-size: 0.85rem;"
                                    >
                                        <strong class="text-primary d-block" style="font-size: 0.75rem;">
                                            {{ getRepliedMessage(msg.replyTo)?.senderName || 'Utente' }}
                                        </strong>
                                        <span class="text-truncate d-inline-block w-100">
                                            <span v-if="getRepliedMessage(msg.replyTo)?.messageType === 'photo' && !getRepliedMessage(msg.replyTo)?.content">
                                                📷 Foto
                                            </span>
                                            <span v-else>
                                                {{ getRepliedMessage(msg.replyTo)?.content || 'Contenuto non disponibile' }}
                                            </span>
                                        </span>
                                    </div>

                                    <div v-if="msg.content" style="font-size: 0.95rem;">
                                        {{ msg.content }}
                                    </div>

                                    <div v-if="msg.reactions && msg.reactions.length > 0" class="d-flex flex-wrap gap-1 mt-1">
                                        <span 
                                            v-for="(reaction, rIndex) in msg.reactions" 
                                            :key="rIndex"
                                            @click.stop="reaction.userId === userId ? toggleReaction(msg.id, reaction.emoji) : null"
                                            class="badge bg-white text-dark shadow-sm d-flex align-items-center justify-content-center"
                                            :class="reaction.userId === userId ? 'border border-primary' : 'border border-light'"
                                            style="font-size: 0.85rem; padding: 3px 6px; border-radius: 12px; cursor: pointer;"
                                            :title="reaction.userId === userId ? 'Clicca per rimuovere' : ''"
                                        >
                                            {{ reaction.emoji }}
                                        </span>
                                    </div>

                                    <button 
                                        @click.stop="toggleDropdown(msg.id)" 
                                        class="btn btn-sm position-absolute top-0 end-0 mt-1 me-1 border-0" 
                                        :class="msg.senderId == userId ? 'text-white' : 'text-dark'"
                                        style="padding: 2px; background: transparent;"
                                        title="Opzioni"
                                    >
                                        <svg class="feather" style="width: 18px; height: 18px;"><use href="/feather-sprite-v4.29.0.svg#chevron-down"/></svg>
                                    </button>

                                    <div 
                                        v-if="activeDropdownMessageId === msg.id"
                                        class="position-absolute bg-white border rounded shadow-sm py-1 z-3"
                                        style="top: 25px; right: 10px; min-width: 130px;"
                                    >
                                        <button @click.stop="handleAction('react', msg.id)" class="dropdown-item d-flex align-items-center text-dark text-start w-100 px-3 py-2 border-0 bg-transparent">
                                            <svg class="feather me-2" style="width: 14px; height: 14px;"><use href="/feather-sprite-v4.29.0.svg#smile"/></svg>
                                            Reagisci
                                        </button>
                                        <button @click.stop="handleAction('forward', msg.id)" class="dropdown-item d-flex align-items-center text-dark text-start w-100 px-3 py-2 border-0 bg-transparent">
                                            <svg class="feather me-2" style="width: 14px; height: 14px;"><use href="/feather-sprite-v4.29.0.svg#corner-up-right"/></svg>
                                            Inoltra
                                        </button>
                                        <button v-if="msg.senderId == userId" @click.stop="handleAction('delete', msg.id)" class="dropdown-item d-flex align-items-center text-danger text-start w-100 px-3 py-2 border-0 bg-transparent">
                                            <svg class="feather me-2" style="width: 14px; height: 14px;"><use href="/feather-sprite-v4.29.0.svg#trash-2"/></svg>
                                            Elimina
                                        </button>
                                        <button 
                                            @click.stop="handleAction('reply', msg.id)"
                                            class="dropdown-item d-flex align-items-center text-dark text-start w-100 px-3 py-2 border-0 bg-transparent"
                                            onmouseover="this.style.backgroundColor='#f8f9fa'"
                                            onmouseout="this.style.backgroundColor='transparent'"
                                        >
                                            <svg class="feather me-2" style="width: 14px; height: 14px;"><use href="/feather-sprite-v4.29.0.svg#message-square"/></svg>
                                            Rispondi
                                        </button>
                                    </div>

                                    <div 
                                        v-if="activeReactionMessageId === msg.id"
                                        class="position-absolute bg-white border rounded shadow-lg p-1 d-flex gap-1 z-3"
                                        style="top: 25px; right: 10px; min-width: max-content;"
                                    >
                                        <button 
                                            v-for="emoji in emojis" 
                                            :key="emoji"
                                            @click.stop="toggleReaction(msg.id, emoji)"
                                            class="btn btn-sm btn-light rounded-circle border-0 d-flex align-items-center justify-content-center"
                                            style="width: 32px; height: 32px; font-size: 1.2rem; padding: 0; transition: transform 0.1s ease;"
                                            onmouseover="this.style.transform='scale(1.2)'"
                                            onmouseout="this.style.transform='scale(1)'"
                                        >
                                            {{ emoji }}
                                        </button>
                                    </div>
                                    
                                        
                                    <div 
                                        class="d-flex align-items-center justify-content-end mt-1" 
                                        style="font-size: 0.65rem;"
                                        :class="msg.senderId == userId ? 'text-white-50' : 'text-muted'"
                                    >
                                        <span class="me-1">{{ formatTime(msg.timestamp) }}</span>
                                        
                                        <span v-if="msg.senderId == userId" class="ms-1" style="font-size: 0.8rem;">
                                            <span v-if="msg.read" class="text-info" title="Letto">✓✓</span>
                                            <span v-else-if="msg.delivered" class="text-white-50" title="Consegnato">✓✓</span>
                                            <span v-else class="text-white-50" title="Inviato">✓</span>
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>                  
                        
                    </div>

					<div class="mt-3 mb-5">

						<div v-if="selectedPhoto" class="mb-2 p-2 bg-light border rounded d-flex justify-content-between align-items-center">
                            <span class="small text-muted text-truncate me-2">
                                <svg class="feather me-1" style="width: 14px; height: 14px;"><use href="/feather-sprite-v4.29.0.svg#image"/></svg>
                                Foto pronta: <strong>{{ selectedPhoto.name }}</strong>
                            </span>
                            
                            <button @click="selectedPhoto = null; if(fileInput) fileInput.value = ''" class="btn btn-sm btn-outline-danger py-0 px-1 flex-shrink-0">
                                <svg class="feather" style="width: 14px; height: 14px;"><use href="/feather-sprite-v4.29.0.svg#x"/></svg>
                            </button>
                        </div>

						<div 
							v-if="replyingToMessage" 
							class="bg-light border-start border-primary border-4 p-2 mb-1 mx-2 rounded position-relative d-flex justify-content-between align-items-center shadow-sm"
						>
							<div class="text-truncate pe-3">
								<strong class="text-primary d-block" style="font-size: 0.8rem;">
									Rispondi a {{ replyingToMessage.senderName }}
								</strong>
								<span class="text-muted text-truncate d-inline-block" style="max-width: 250px; font-size: 0.85rem;">
									<span v-if="replyingToMessage.messageType === 'photo' && !replyingToMessage.content">
										📷 Foto
									</span>
									<span v-else>{{ replyingToMessage.content }}</span>
								</span>
							</div>
							
							<button 
								@click="cancelReply" 
								class="btn btn-sm text-danger p-1 border-0" 
								title="Annulla risposta"
							>
								<svg class="feather" style="width: 20px; height: 20px;"><use href="/feather-sprite-v4.29.0.svg#x"/></svg>
							</button>
						</div>

						<div class="input-group shadow-sm">
							
							<input 
								type="file" 
								ref="fileInput" 
								@change="handlePhotoSelected" 
								accept="image/*" 
								class="d-none"
							>

							<button 
								class="btn btn-outline-secondary" 
								type="button" 
								@click="triggerFileInput"
								title="Invia una foto"
							>
								<svg class="feather" style="width: 18px; height: 18px;"><use href="/feather-sprite-v4.29.0.svg#camera"/></svg>
							</button>

							<input 
								type="text" 
								class="form-control" 
								placeholder="Scrivi un messaggio..."
								v-model="messageText"
								@keyup.enter="sendMessage"
							>
							<button 
								class="btn btn-primary" 
								type="button" 
								@click="sendMessage"
							>
								Invia
							</button>
						</div>
					</div>
				</div>

				<div v-else>
					<RouterView />
				</div>

			</main>
		</div>
	</div>

	<div v-if="forwardingMessageId" class="position-fixed top-0 start-0 w-100 h-100 d-flex justify-content-center align-items-center" style="background: rgba(0,0,0,0.6); z-index: 1060;">
		<div class="bg-white p-4 rounded shadow-lg" style="width: 350px; max-width: 90%;">
			<h5 class="mb-3 text-dark">Inoltra messaggio a...</h5>
			
			<div v-if="chats.length === 0" class="text-muted small">
				Nessuna chat disponibile.
			</div>

			<div class="list-group mb-3" style="max-height: 250px; overflow-y: auto;">
				<button 
					v-for="(chat, index) in chats" 
					:key="index"
					@click="forwardMessage(chat)"
					class="list-group-item list-group-item-action d-flex align-items-center"
				>
					<svg class="feather me-2 text-muted" style="width: 16px; height: 16px;"><use href="/feather-sprite-v4.29.0.svg#message-circle"/></svg>
					{{ chat.name || 'Chat #' + (chat.id || chat.conversationId) }}
				</button>
			</div>
			
			<div class="text-end mt-2">
				<button class="btn btn-sm btn-secondary" @click="cancelForward">Annulla</button>
			</div>
		</div>
	</div>

	<div v-if="isProfileModalOpen" class="position-fixed top-0 start-0 w-100 h-100 d-flex justify-content-center align-items-center" style="background: rgba(0,0,0,0.6); z-index: 1060;">
		<div class="bg-white p-4 rounded shadow-lg" style="width: 400px; max-width: 90%;">
			<div class="d-flex justify-content-between align-items-center mb-4">
				<h5 class="mb-0 text-dark">Modifica Profilo</h5>
				<button @click="closeProfileModal" class="btn-close"></button>
			</div>
			
			<div v-if="profileError" class="alert alert-danger py-2 small mb-3">
				{{ profileError }}
			</div>

			<div class="text-center mb-4">
				<div class="position-relative d-inline-block">
					<div 
						class="rounded-circle bg-light border d-flex align-items-center justify-content-center overflow-hidden"
						style="width: 100px; height: 100px;"
					>	
						<img 
							v-if="profilePhotoPreview" 
							:src="getImageUrl(profilePhotoPreview)" 
							class="w-100 h-100" 
							style="object-fit: cover;" 
						/>
						<svg v-else class="feather text-secondary" style="width: 50px; height: 50px;"><use :href="'/feather-sprite-v4.29.0.svg#' + STOCK_PHOTO_ICON"/></svg>
					</div>
					
					<label class="position-absolute bottom-0 end-0 bg-primary text-white rounded-circle p-2 shadow" style="cursor: pointer;">
						<svg class="feather" style="width: 16px; height: 16px;"><use href="/feather-sprite-v4.29.0.svg#camera"/></svg>
						<input type="file" class="d-none" accept="image/*" @change="handleProfilePhotoSelected">
					</label>
				</div>
			</div>

			<div class="mb-4">
				<label class="form-label small fw-bold text-muted">Nome Utente</label>
				<div class="input-group">
					<span class="input-group-text bg-light">@</span>
					<input type="text" class="form-control" v-model="newUsername" placeholder="Il tuo nome utente">
				</div>
			</div>
			
			<div class="d-flex justify-content-end gap-2">
				<button class="btn btn-secondary" @click="closeProfileModal">Annulla</button>
				<button class="btn btn-primary" @click="saveProfile">Salva Modifiche</button>
			</div>
		</div>
	</div>
	<div v-if="isGroupModalOpen" class="position-fixed top-0 start-0 w-100 h-100 d-flex justify-content-center align-items-center" style="background: rgba(0,0,0,0.6); z-index: 1060;">
		<div class="bg-white p-4 rounded shadow-lg" style="width: 450px; max-width: 90%;">
			<div class="d-flex justify-content-between align-items-center mb-4">
				<h5 class="mb-0 text-dark">Nuovo Gruppo</h5>
				<button @click="closeGroupModal" class="btn-close"></button>
			</div>

			<div v-if="groupError" class="alert alert-danger py-2 small mb-3">{{ groupError }}</div>

			<div class="text-center mb-4">
				<div class="position-relative d-inline-block">
					<div class="rounded-circle bg-light border d-flex align-items-center justify-content-center overflow-hidden" style="width: 80px; height: 80px;">
						<img v-if="groupPhotoPreview" :src="getImageUrl(groupPhotoPreview)" class="w-100 h-100" style="object-fit: cover;" />
						<svg v-else class="feather text-secondary" style="width: 30px; height: 30px;"><use href="/feather-sprite-v4.29.0.svg#users"/></svg>
					</div>
					<label class="position-absolute bottom-0 end-0 bg-primary text-white rounded-circle p-1 shadow" style="cursor: pointer;">
						<svg class="feather" style="width: 14px; height: 14px;"><use href="/feather-sprite-v4.29.0.svg#camera"/></svg>
						<input type="file" class="d-none" @change="handleGroupPhotoSelected">
					</label>
				</div>
			</div>

			<div class="mb-3">
				<label class="form-label small fw-bold">Nome Gruppo</label>
				<input type="text" class="form-control" v-model="newGroupName" placeholder="Esempio: Calcetto">
			</div>

			<div class="mb-4">
				<label class="form-label small fw-bold">Seleziona Partecipanti</label>
				<div class="border rounded p-2" style="max-height: 150px; overflow-y: auto;">
					<div v-for="user in allUsers" :key="user.id" class="form-check">
						<input class="form-check-input" type="checkbox" :value="user.id" :id="'u'+user.id"  v-model="selectedUsers">
						<label class="form-check-label small" :for="'u'+user.id">@{{ user.username }}</label>
					</div>
				</div>
			</div>

			<div class="d-flex justify-content-end gap-2">
				<button class="btn btn-secondary" @click="closeGroupModal">Annulla</button>
				<button class="btn btn-primary" @click="createGroup">Crea</button>
			</div>
		</div>
	</div>

	<div v-if="isGroupInfoModalOpen" class="position-fixed top-0 start-0 w-100 h-100 d-flex justify-content-center align-items-center" style="background: rgba(0,0,0,0.6); z-index: 1060;">
    <div class="bg-white p-4 rounded shadow-lg d-flex flex-column" style="width: 450px; max-width: 90%; max-height: 90vh;">
        
        <div class="d-flex justify-content-between align-items-center mb-4">
            <h5 class="mb-0 text-dark">Impostazioni Gruppo</h5>
            <button @click="closeGroupInfo" class="btn-close"></button>
        </div>

        <div class="overflow-auto pe-2" style="flex-grow: 1;">
            <div class="text-center mb-4">
                <div class="position-relative d-inline-block">
                    <div class="rounded-circle bg-light border d-flex align-items-center justify-content-center overflow-hidden shadow-sm" style="width: 120px; height: 120px;">
                        <img 
                            v-if="editingGroupPhotoPreview" 
                            :src="getImageUrl(editingGroupPhotoPreview)" 
                            class="w-100 h-100" 
                            style="object-fit: cover;" 
                        />
                        <svg v-else class="feather text-secondary" style="width: 50px; height: 50px;"><use href="/feather-sprite-v4.29.0.svg#users"/></svg>
                    </div>
                    
                    <label class="position-absolute bottom-0 end-0 bg-primary text-white rounded-circle p-2 shadow" style="cursor: pointer; transition: 0.2s;" title="Cambia foto">
                        <svg class="feather" style="width: 20px; height: 20px;"><use href="/feather-sprite-v4.29.0.svg#camera"/></svg>
                        <input 
                            type="file" 
                            class="d-none" 
                            accept="image/*" 
                            @change="handleEditingGroupPhotoSelected"
                        >
                    </label>
                </div>

                <div v-if="editingGroupPhotoFile" class="mt-3">
                    <button @click="saveGroupPhoto" class="btn btn-sm btn-primary px-3">
                        Conferma Nuova Foto
                    </button>
                    <button @click="editingGroupPhotoFile = null; editingGroupPhotoPreview = getActiveChat().photoUrl" class="btn btn-sm btn-link text-muted">
                        Annulla
                    </button>
                </div>
            </div>

            <div class="mb-4">
				<label class="form-label small fw-bold text-muted">Nome Gruppo</label>
				<div class="input-group">
					<input 
						type="text" 
						class="form-control" 
						v-model="editingGroupName"
						@keyup.enter="saveGroupName" 
					>
					<button 
						class="btn btn-primary" 
						type="button" 
						@click="saveGroupName"
					>
						Salva
					</button>
				</div>
				<div v-if="groupInfoError" class="text-danger small mt-1">
					{{ groupInfoError }}
				</div>
			</div>

            <div class="mb-4">
				<label class="form-label small fw-bold text-muted">Aggiungi Partecipanti</label>
				
				<div class="position-relative">
                    <div class="input-group">
						<input 
							type="text" 
							class="form-control" 
							placeholder="Cerca utente..." 
							v-model="groupAddSearchQuery"
							@focus="isGroupAddSearchFocused = true; searchUsersForGroup()"
							@blur="hideGroupAddDropdown"
						>
						<button 
							class="btn btn-primary d-flex align-items-center" 
							type="button" 
							@click="addMembersToGroup"
							:disabled="selectedUsersToAdd.length === 0"
						>
							<svg class="feather me-1" style="width: 16px; height: 16px;"><use href="/feather-sprite-v4.29.0.svg#user-plus"/></svg>
							Aggiungi
						</button>
					</div>

                    <div 
						v-if="isGroupAddSearchFocused" 
						class="dropdown-menu show w-100 shadow-lg border-0 mt-1" 
						style="position: absolute; top: 100%; left: 0; z-index: 1070; max-height: 200px; overflow-y: auto;"
						@mousedown.prevent
					>
                        <div v-if="groupAddSearchResults.length === 0" class="dropdown-item text-muted small fst-italic text-center py-2">
                            Nessun utente disponibile.
                        </div>

                        <label 
							v-for="user in groupAddSearchResults" 
							:key="user.id" 
							class="dropdown-item d-flex align-items-center py-2 mb-0"
                            style="cursor: pointer;"
                            onmouseover="this.style.backgroundColor='#f8f9fa'"
                            onmouseout="this.style.backgroundColor='transparent'"
						>
							<div class="form-check mb-0 w-100 d-flex align-items-center">
								<input 
									class="form-check-input border-secondary me-2 mt-0" 
									type="checkbox" 
									:value="user.id" 
									v-model="selectedUsersToAdd"
									style="cursor: pointer;"
								>
								<span class="small">@{{ user.username }}</span>
							</div>
						</label>
					</div>
				</div>

                <div v-if="selectedUsersToAdd.length > 0" class="mt-2 p-2 bg-light border rounded shadow-sm d-flex flex-wrap gap-1">
					<span class="small text-muted w-100 mb-1" style="font-size: 0.7rem; text-transform: uppercase;">Pronti per l'aggiunta:</span>
					<span 
						v-for="id in selectedUsersToAdd" 
						:key="id" 
						class="badge bg-primary d-flex align-items-center py-1 pe-2 shadow-sm"
					>
						@{{ groupAddSearchResults.find(u => u.id === id)?.username || 'Utente' }}
                        
                        <button 
                            type="button" 
                            class="btn-close btn-close-white ms-2" 
                            style="font-size: 0.45rem;" 
                            @click="selectedUsersToAdd = selectedUsersToAdd.filter(uId => uId !== id)"
                            title="Rimuovi selezione"
                        ></button>
					</span>
				</div>
			</div>

            <div class="mb-2">
				<label class="form-label small fw-bold text-muted">Membri del Gruppo ({{ groupMembersList.length }})</label>
				<ul class="list-group">
					<li v-for="member in groupMembersList" :key="member.id" class="list-group-item d-flex justify-content-between align-items-center py-2">
						<span class="fw-medium">
							@{{ member.username }}
							<span v-if="member.id == userId" class="badge bg-secondary ms-1" style="font-size: 0.65rem;">Tu</span>
						</span>
						
						<button 
							class="btn btn-sm btn-outline-danger py-0 px-2" 
							title="Rimuovi dal gruppo"
							@click="removeMemberFromGroup(member.id)"
						>
							<svg class="feather" style="width: 14px; height: 14px;"><use href="/feather-sprite-v4.29.0.svg#user-minus"/></svg>
						</button>
					</li>
				</ul>
			</div>

			<hr class="my-4">

            <div class="d-grid mb-3">
                <button 
                    @click="removeMemberFromGroup(userId)" 
                    class="btn btn-outline-danger d-flex align-items-center justify-content-center py-2"
                >
                    <svg class="feather me-2" style="width: 18px; height: 18px;"><use href="/feather-sprite-v4.29.0.svg#log-out"/></svg>
                    Abbandona Gruppo
                </button>
            </div>
        </div>

    </div>
</div>

</template>

<style>
/* Stili globali */
</style>