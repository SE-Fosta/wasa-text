import { createRouter, createWebHashHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{ path: '/', name: 'home', component: HomeView },
		{ path: '/login', name: 'login', component: LoginView }
	]
})

// Guardia di navigazione: se non hai il token, ti rimando al login
router.beforeEach((to, from, next) => {
	const isAuthenticated = localStorage.getItem('token');

	if (to.name !== 'login' && !isAuthenticated) {
		next({ name: 'login' });
	} else {
		next();
	}
})

export default router