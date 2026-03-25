<script setup>
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'

// Usiamo useRoute per sapere in quale pagina ci troviamo
const route = useRoute();
// Usiamo useRouter per fare i reindirizzamenti
const router = useRouter();

// Funzione per il logout: cancella il token e torna al login
const logout = () => {
	localStorage.removeItem('token');
	router.push('/login');
}
</script>

<template>
	<template v-if="route.name !== 'login'">
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

						<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
							<span>Account</span>
						</h6>
						<ul class="nav flex-column">
							<li class="nav-item">
								<a href="#" @click.prevent="logout" class="nav-link text-danger">
									<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#log-out"/></svg>
									Esci
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

	<template v-else>
		<RouterView />
	</template>
</template>

<style>
/* Se hai stili globali lasciali pure qui, altrimenti lascia vuoto */
</style>