<script setup>
import { ref, watch, onMounted } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import api from './services/axios.js'

const route = useRoute();
const username = ref('');
const chats = ref([]);

// 1. Variabili per la ricerca
const searchQuery = ref('');
const searchResults = ref([]);

const isSearchFocused = ref(false);

// Funzione per caricare le chat dell'utente
const updateData = async () => {
	username.value = localStorage.getItem('username') || '';
	const userId = localStorage.getItem('token');
	
	if (userId) {
		try {
			let response = await api.get(`/users/${userId}/conversations`);
			chats.value = response.data || [];
		} catch (e) {
			console.error("Errore nel recupero delle chat:", e);
		}
	} else {
		chats.value = [];
	}

	// AGGIUNGI QUESTA RIGA PER CARICARE TUTTI GLI UTENTI:
	searchUsers();
};

// 2. La nuova funzione che chiama il tuo backend Go!
const searchUsers = async () => {
	// ABBIAMO CANCELLATO L'IF CON IL BLOCCO DELLE 2 LETTERE!
	
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



onMounted(() => {
	updateData();
});

watch(() => route.path, () => {
	updateData();
});

// 3. "Ascoltiamo" la barra di ricerca: ogni volta che digiti, parte la funzione searchUsers
watch(searchQuery, () => {
	searchUsers();
});

// Funzione per nascondere la tendina quando clicchi fuori
const hideDropdown = () => {
	setTimeout(() => {
		isSearchFocused.value = false;
	}, 50); // Il ritardo serve per darti il tempo di cliccare il bottone "Chat"!
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
								<button class="btn btn-sm btn-outline-primary py-0 px-2" style="font-size: 0.75rem;">Chat</button>
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
							<a class="nav-link py-3 text-truncate" href="#">
								<svg class="feather me-2"><use href="/feather-sprite-v4.29.0.svg#message-circle"/></svg>
								{{ chat }}
							</a>
						</li>
					</ul>

				</div>
			</nav>

			<main class="col-md-9 ms-sm-auto col-lg-10 px-md-4 mt-3">
				<RouterView />
			</main>
		</div>
	</div>
</template>

<style>
/* Stili globali */
</style>