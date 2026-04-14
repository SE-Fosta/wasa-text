<script setup>
import { ref, watch, onMounted , nextTick} from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import api from './services/axios.js'

const route = useRoute();
const router = useRouter();

// 1. INIZIALIZZIAMO SUBITO LE VARIABILI LEGGENDO IL BROWSER
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

// Variabili per l'invio delle foto
const fileInput = ref(null);
const selectedPhoto = ref(null);

const doLogout = () => {
    // Cancella i dati salvati nel browser
    localStorage.removeItem('token');
    localStorage.removeItem('username');
	localStorage.removeItem('userId');

    // Resetta le variabili a schermo (USANDO .value!)
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

// Carica le chat dell'utente
const updateData = async () => {
    // AGGIORNIAMO LE VARIABILI USANDO .value
    userId.value = localStorage.getItem('token');
    username.value = localStorage.getItem('username') || '';

    // USIAMO userId.value PER IL CONTROLLO E PER L'URL
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

    // Carica tutti gli utenti
    searchUsers();
};

const formatTime = (timestamp) => {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    // Mostra solo ore e minuti
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};

const truncateText = (text, maxLength) => {
    if (!text) return '';
    if (text.length <= maxLength) return text;
    // Taglia il testo e aggiunge i puntini di sospensione
    return text.substring(0, maxLength) + '...';
};

const scrollToBottom = () => {
    if (chatContainer.value) {
        chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
    }
};

const markAsRead = async (conversationId) => {
    try {
        // ROTTA CORRETTA: punta alla conversazione intera
        await api.put(`/conversations/${conversationId}/read`);
    } catch (e) {
        console.warn("Impossibile segnare i messaggi come letti:", e);
    }
};

const loadMessages = async () => {
    if (!activeChatId.value) return;
    
    try {
        // Recuperiamo i messaggi dal server
        const response = await api.get(`/conversations/${activeChatId.value}/messages`);
        const newMessages = response.data || [];
        
        // --- NUOVA LOGICA: SEGNA COME LETTO ---
        // Avvisiamo il backend che abbiamo ricevuto (e stiamo leggendo) i messaggi
        await markAsRead(activeChatId.value);
        // --------------------------------------

        // 1. Capiamo se l'utente sta guardando vecchi messaggi (ha fatto scroll in su)
        let isAtBottom = true;
        if (chatContainer.value) {
            const { scrollTop, scrollHeight, clientHeight } = chatContainer.value;
            // Se la distanza dal fondo è più di 100 pixel, vuol dire che è salito!
            if (scrollHeight - scrollTop - clientHeight > 100) {
                isAtBottom = false;
            }
        }

        // 2. Capiamo se è la prima volta che apre questa chat
        const isFirstLoad = messages.value.length === 0;

        // Aggiorniamo i messaggi a schermo con quelli nuovi (che ora avranno le spunte aggiornate)
        messages.value = newMessages;
        
        // 3. Scorri in basso SOLO se serve veramente!
        if (isFirstLoad || isAtBottom) {
            nextTick(() => {
                scrollToBottom();
            });
        }

    } catch (e) {
        console.error("Errore nel caricamento dei messaggi:", e);
        // Evitiamo di resettare messages.value = [] in caso di errore di rete momentaneo
        // così l'utente continua a vedere i vecchi messaggi finché non torna il segnale.
    }
};


// Quando l'ID della chat cambia (es. clicchi su un altro utente), scarica i suoi messaggi
watch(activeChatId, (newVal) => {
    // 1. Se c'era già un timer di un'altra chat, fermalo!
    if (pollingTimer) {
        clearInterval(pollingTimer);
        pollingTimer = null;
    }

    if (newVal) {
        // 2. Carica i messaggi subito la prima volta
        loadMessages();
        
        // 3. LA MAGIA: Controlla nuovi messaggi ogni 2 secondi (2000 millisecondi)
        pollingTimer = setInterval(() => {
            loadMessages();
        }, 2000);

    } else {
        messages.value = [];
    }
});

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

const startChat = async (selectedUser) => {
    // Nascondiamo la tendina della ricerca
    isSearchFocused.value = false;
    searchQuery.value = '';
    
    // Controllo di sicurezza
    if (!selectedUser || !selectedUser.id) {
        console.error("Errore: l'utente selezionato non ha un ID valido.", selectedUser);
        alert("Impossibile avviare la chat, utente non valido.");
        return;
    }

    try {
        // IL PAYLOAD PER LA CHAT 1-A-1
        const payload = {
            targetUserId: String(selectedUser.id),
            isGroup: false,
            name: "" // Vuoto, non serve per le chat singole
        };

        // Chiamiamo l'endpoint "universale"
        const response = await api.post(`/users/${userId.value}/conversations`, payload);
        
        // Aggiorniamo la sidebar a sinistra e apriamo la chat appena creata
        await updateData(); 
        activeChatId.value = response.data.conversationId; 
        
    } catch (e) {
        console.error("Errore durante la creazione della chat:", e.response?.data || e);
        alert("Errore durante la creazione della chat.");
    }
};

const sendMessage = async () => {
    // 1. SE C'È UNA FOTO PRONTA, MANDA QUELLA!
    if (selectedPhoto.value) {
        await sendPhoto();
        return; // Ci fermiamo qui per non mandare anche il testo a vuoto
    }

    // 2. COMPORTAMENTO NORMALE PER IL TESTO
    if (!messageText.value.trim() || !activeChatId.value || !userId.value) return;

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

const triggerFileInput = () => {
    fileInput.value.click();
};

// Quando l'utente seleziona una foto dal suo PC
const handlePhotoSelected = (event) => {
    const file = event.target.files[0];
    if (!file) return;

    // Salviamo il file selezionato in memoria, ma NON lo inviamo ancora!
    selectedPhoto.value = file;
};

const sendPhoto = async () => {
    if (!selectedPhoto.value || !activeChatId.value || !userId.value) return;

    try {
        const formData = new FormData();
        formData.append("photo", selectedPhoto.value);
        
        // Aggiungiamo anche il tipo di messaggio, così il server sa che è una foto!
        formData.append("messageType", "photo");

        // IL SEGRETO È QUI: Usiamo lo stesso URL dei messaggi di testo!
        await api.post(`/conversations/${activeChatId.value}/messages`, formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        });

        console.log("Foto inviata con successo!");
        selectedPhoto.value = null; 
        
        // Svuotiamo anche l'input file HTML per poter ricaricare la stessa foto se serve
        if (fileInput.value) fileInput.value.value = '';
        
        await loadMessages(); 
        
    } catch (e) {
        console.error("Errore nell'invio della foto:", e);
    }
};

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

const confirmForward = async (targetChat) => {
    if (!forwardingMessageId.value || !targetChat) return;

    // Recuperiamo l'ID della chat di destinazione
    const targetId = targetChat.id || targetChat.conversationId;

    try {
        // Usa la tua istanza axios "api"
        await api.post(`/messages/${forwardingMessageId.value}/forward`, {
            targetConversationId: targetId
        });
        
        console.log("Messaggio inoltrato con successo!");
        forwardingMessageId.value = null; // Chiudi il popup
        
        // (Opzionale) se inoltri in una chat e ci clicchi sopra, si aggiornerà in automatico
    } catch (e) {
        console.error("Errore durante l'inoltro:", e);
        alert("Impossibile inoltrare il messaggio.");
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

// Gestisce l'aggiunta o la rimozione della reazione
const toggleReaction = async (messageId, emoji) => {
    console.log("1. Hai cliccato l'emoji:", emoji, "sul messaggio:", messageId);

    // Chiudiamo il menu a tendina
    activeReactionMessageId.value = null;

    // Troviamo il messaggio nei nostri dati locali
    const msg = messages.value.find(m => m.id === messageId);
    console.log("2. Messaggio trovato nei dati locali:", msg);

    if (!msg) {
        console.error("ERRORE: Non trovo il messaggio! Forse messageId è undefined?");
        return; 
    }

    // Controlliamo se l'utente loggato ha già una reazione
    const myExistingReaction = msg.reactions?.find(r => r.userId == userId.value);
    console.log("3. La tua reazione precedente era:", myExistingReaction);

    try {
        if (myExistingReaction && myExistingReaction.emoji === emoji) {
            console.log("4. Stai togliendo l'emoji... Faccio la DELETE!");
            await api.delete(`/messages/${messageId}/reactions`);
        } else {
            console.log("4. Stai mettendo una nuova emoji... Faccio la POST!");
            await api.post(`/messages/${messageId}/reactions`, { emoji: emoji });
        }
        
        console.log("5. Chiamata API andata a buon fine! Ricarico i messaggi.");
        await loadMessages();
    } catch (e) {
        console.error("ERRORE DURANTE LA CHIAMATA API:", e);
    }
};

// Apre o chiude la tendina
const toggleDropdown = (messageId) => {
    if (activeDropdownMessageId.value === messageId) {
        activeDropdownMessageId.value = null; // La chiude se era già aperta
    } else {
        activeDropdownMessageId.value = messageId; // La apre
        // Chiude eventuali menu delle emoji aperti per evitare sovrapposizioni
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
    newUsername.value = username.value; // Pre-compila col nome attuale
    profilePhotoPreview.value = userPhotoUrl.value; // Mostra la foto attuale
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
    // Prendi la base da axios e pulisci eventuali residui
    const baseUrl = api.defaults.baseURL?.split('/api')[0] || 'http://localhost:3000';
    return `${baseUrl}${path.startsWith('/') ? path : '/' + path}`;
};

// Quando l'utente sceglie una nuova foto cliccando sull'immagine
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
        // 1. CAMBIO NOME UTENTE
        if (newUsername.value.trim() !== username.value && newUsername.value.trim() !== '') {
            await api.put(`/users/${userId.value}/username`, 
                { username: newUsername.value.trim() },
                { headers: { 'Authorization': `Bearer ${userId.value}` } } // Invia il token/ID
            );
            
            username.value = newUsername.value.trim();
            localStorage.setItem('username', username.value);
            needsUpdate = true;
        }

        // 2. CAMBIO FOTO PROFILO
        if (profilePhotoFile.value) {
            const formData = new FormData();
            formData.append("photo", profilePhotoFile.value);
            
            // Invio con Multipart Form e Authorization Header
            await api.put(`/users/${userId.value}/photo`, formData, {
                headers: { 
                    'Content-Type': 'multipart/form-data',
                    'Authorization': `Bearer ${userId.value}` // Necessario per wrapAuth
                }
            });

            // Aggiorniamo l'URL locale. Aggiungiamo un timestamp (?t=...) 
            // per "ingannare" la cache del browser e mostrare subito la nuova foto
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
        // Carichiamo tutti gli utenti per poterli scegliere
        const response = await api.get('/users');
        // Escludiamo noi stessi dalla lista
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

const createGroup = async () => {
    if (!newGroupName.value.trim() || selectedUsers.value.length === 0) {
        groupError.value = "Inserisci un nome e seleziona almeno un membro.";
        return;
    }

    try {
        // 1. CREA LA STANZA
        const payload = {
            isGroup: true,
            name: newGroupName.value.trim(),
            targetUserId: "" 
        };
        const response = await api.post(`/users/${userId.value}/conversations`, payload);
        const newGroupId = response.data.conversationId;

        // 2. AGGIUNGI I MEMBRI (Aspettiamo che finiscano TUTTI in modo pulito)
        // Usiamo Promise.all per evitare di incartare Axios con un ciclo "for" lento
        const memberPromises = selectedUsers.value.map(uId => {
            return api.post(`/groups/${newGroupId}/members`, { userId: String(uId) });
        });
        await Promise.all(memberPromises); // Aspetta che tutti i membri siano aggiunti!

        // 3. FASE FINALE
        await updateData(); // Ora Axios è libero e il token viaggerà sicuro!
        
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
			<nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
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
								<a class="nav-link py-2 d-flex flex-column" href="#" @click.prevent="activeChatId = chat.id || chat.conversationId">
									
									<div class="d-flex w-100 align-items-center justify-content-between mb-1">
										<strong class="text-dark text-truncate d-flex align-items-center">
											<svg class="feather me-2" style="width: 16px; height: 16px; min-width: 16px;"><use href="/feather-sprite-v4.29.0.svg#message-circle"/></svg>
											{{ chat.name || 'Chat #' + (chat.id || chat.conversationId) }}
										</strong>
										
										<span v-if="chat.unreadCount > 0" class="badge bg-success rounded-pill ms-2">
											{{ chat.unreadCount }}
										</span>
										<small v-else-if="chat.lastActivity" class="text-muted ms-2 text-nowrap" style="font-size: 0.75rem;">
											{{ formatTime(chat.lastActivity) }}
										</small>
									</div>
									
									<div class="text-muted ps-4" style="font-size: 0.85rem; max-width: 100%;">
										<span v-if="chat.lastMessage && chat.lastMessage.content">
											
											<span v-if="chat.isGroup && chat.lastMessage.senderName" class="fw-bold text-dark">
												{{ chat.lastMessage.senderName }}: 
											</span>
											
											{{ truncateText(chat.lastMessage.content, 40) }}
											
										</span>
										<span v-else class="fst-italic">Nessun messaggio</span>
									</div>
									
								</a>
							</li>
						</ul>
					</div>
					
					<div class="mt-auto border-top p-3 w-100 bg-white" style="z-index: 10;">
						<div class="d-flex justify-content-center gap-3">
							
							<button 
								@click="openProfileModal" 
								class="btn rounded-circle p-0 d-flex align-items-center justify-content-center shadow-sm border"
								style="width: 42px; height: 42px; background-color: #f8f9fa;"
								:title="'Modifica profilo di @' + username"
							>
								<img 
									v-if="userPhotoUrl" 
									:src="getImageUrl(userPhotoUrl)"
									class="rounded-circle w-100 h-100" 
									style="object-fit: cover;"
								/>
								<svg 
									v-else
									class="feather text-secondary" 
									style="width: 20px; height: 20px;"
								>
									<use :href="'/feather-sprite-v4.29.0.svg#' + STOCK_PHOTO_ICON"/>
								</svg>
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

			<main class="col-md-9 ms-sm-auto col-lg-10 px-md-4 mt-3" style="height: 100vh; overflow-y: auto;">
				
				<div v-if="activeChatId" class="d-flex flex-column h-100">
					<div class="d-flex justify-content-between align-items-center pt-3 pb-2 mb-3 border-bottom">
						<h1 class="h3 mb-0">
							{{ chats.find(c => (c.id || c.conversationId) === activeChatId)?.name || 'Chat #' + activeChatId }}
						</h1>
						<button class="btn btn-sm btn-outline-danger" @click="activeChatId = null">Chiudi Chat</button>
					</div>
					
					<div ref="chatContainer" class="flex-grow-1 p-3 bg-white border rounded overflow-auto d-flex flex-column" style="max-height: calc(100vh - 200px);">
    
						<div v-if="messages.length === 0" class="text-center text-muted mt-5">
							Nessun messaggio. Scrivi qualcosa per rompere il ghiaccio!
						</div>

						<div 
							v-for="(msg, index) in messages" 
							:key="index" 
							class="mb-3 d-flex flex-column"
							:class="msg.senderId == userId ? 'align-items-end' : 'align-items-start'"
						>
							<div class="d-flex align-items-start">

								<div 
									class="p-2 rounded shadow-sm text-break position-relative pe-4" 
									:class="msg.senderId == userId ? 'bg-primary text-white' : 'bg-light text-dark border'"
									style="max-width: 75%; min-width: 120px; text-align: left;"
								>

									

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
							<span class="small text-muted">
								<svg class="feather me-1" style="width: 14px; height: 14px;"><use href="/feather-sprite-v4.29.0.svg#image"/></svg>
								Foto pronta: <strong>{{ selectedPhoto.name }}</strong>
							</span>
							<button @click="selectedPhoto = null; if(fileInput) fileInput.value = ''" class="btn btn-sm btn-outline-danger py-0 px-1">
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
					@click="confirmForward(chat)"
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
						<input class="form-check-input" type="checkbox" :value="user.id" :id="'u'+user.id" v-model="selectedUsers">
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

</template>

<style>
/* Stili globali */
</style>