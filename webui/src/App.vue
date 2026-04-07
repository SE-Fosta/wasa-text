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

// Variabili per la ricerca
const searchQuery = ref('');
const searchResults = ref([]);
const isSearchFocused = ref(false);

// Variabili per la chat
const activeChatId = ref(null);
const messageText = ref('');
const messages = ref([]);

// Variabili per l'invio delle foto
const fileInput = ref(null);
const selectedPhoto = ref(null);

const doLogout = () => {
    // Cancella i dati salvati nel browser
    localStorage.removeItem('token');
    localStorage.removeItem('username');

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
    // Chiude la tendina e pulisce la ricerca
    isSearchFocused.value = false;
    searchQuery.value = '';
    
    // Controlla se sei loggato usando .value
    if (!userId.value) {
        console.error("Errore: Utente non loggato");
        return;
    }

    try {
        // CORRETTO myUserId in userId.value
        const response = await api.post(`/users/${userId.value}/conversations`, {
            targetUserId: selectedUser.id 
        });
        
        await updateData();
        activeChatId.value = response.data.conversationId;
        
    } catch (e) {
        console.error("Errore durante la creazione della chat:", e);
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
            messageType: "text"
        });

        messageText.value = ''; 
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
				<div class="position-sticky pt-3 sidebar-sticky">
					
					<div class="px-3 mb-4 mt-2 position-relative">
						<label class="form-label small text-muted text-uppercase fw-bold">Cerca Utenti</label>
						
						<input 
							type="text" 
							v-model="searchQuery" 
							@focus="isSearchFocused = true"
							@blur="hideDropdown"
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
									@click.stop="startChat(user)" 
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
							<a class="nav-link py-3 text-truncate" href="#" @click.prevent="activeChatId = chat.id || chat.conversationId">
								<svg class="feather me-2"><use href="/feather-sprite-v4.29.0.svg#message-circle"/></svg>
								{{ chat.name || 'Chat #' + (chat.id || chat.conversationId) }}
							</a>
						</li>
					</ul>

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
							<div 
								class="p-2 rounded shadow-sm text-break position-relative" 
								:class="msg.senderId == userId ? 'bg-primary text-white' : 'bg-secondary text-white'"
								style="max-width: 75%; min-width: 120px; text-align: left;"
							>
								

								<div v-if="msg.messageType === 'photo' || msg.photoUrl" class="mb-1">
									<img :src="msg.photoUrl" class="img-fluid rounded" alt="Foto" style="max-height: 200px; object-fit: cover;">
								</div>

								<div v-if="msg.content" style="font-size: 0.95rem;">
									{{ msg.content }}
								</div>

								<div 
									class="p-2 rounded shadow-sm text-break position-relative pe-4" 
									:class="msg.senderId == userId ? 'bg-primary text-white' : 'bg-light text-dark border'"
									style="max-width: 75%; min-width: 120px; text-align: left;"
								>

									<button 
										v-if="msg.senderId == userId" 
										@click.stop="deleteMessage(msg.id)" 
										class="btn btn-sm position-absolute top-0 end-0 text-white border-0 mt-1 me-1" 
										style="padding: 2px;"
										title="Elimina"
									>
										<svg class="feather" style="width: 14px; height: 14px;"><use href="/feather-sprite-v4.29.0.svg#trash-2"/></svg>
									</button>
										

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
</template>

<style>
/* Stili globali */
</style>