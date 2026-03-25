<script setup>
import { ref, watch, onMounted } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

const route = useRoute();
const username = ref('');

// Questa funzione legge il nome dal browser
const updateUsername = () => {
	username.value = localStorage.getItem('username') || '';
};

// Lo legge appena apri l'app
onMounted(() => {
	updateUsername();
});

// "Ascolta" i cambi di pagina: così quando fai login si aggiorna istantaneamente senza dover premere F5!
watch(() => route.path, () => {
	updateUsername();
});
</script>

<template>
	<header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
		<a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">WASAText</a>
		<button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
			<span class="navbar-toggler-icon"></span>
		</button>
	</header>

	<div class="container-fluid">
		<div class="row">
			<nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
				<div class="position-sticky pt-3 sidebar-sticky">
					
					<div v-if="username" class="text-center mt-3 mb-2">
						<div class="d-inline-block bg-primary text-white rounded-circle p-2 mb-2" style="width: 50px; height: 50px; line-height: 35px; font-size: 1.5rem;">
							{{ username.charAt(0).toUpperCase() }}
						</div>
						<div class="fw-bold fs-5 text-dark">@{{ username }}</div>
					</div>
					<hr v-if="username" class="mx-3">
					<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
						<span>Menu Principale</span>
					</h6>
					<ul class="nav flex-column">
						<li class="nav-item">
							<RouterLink to="/" class="nav-link" active-class="active">
								<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#message-circle"/></svg>
								Le mie Chat
							</RouterLink>
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