// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
                            routes: [
                                { path: '/login', name: 'login', component: LoginView },
                                { path: '/', name: 'home', component: HomeView }
                            ]
})

// Guardia di navigazione: se non ho l'ID, vado al login
router.beforeEach((to, from, next) => {
    const publicPages = ['/login'];
    const authRequired = !publicPages.includes(to.path);
    const loggedIn = localStorage.getItem('token');

    if (authRequired && !loggedIn) {
        next('/login');
    } else {
        next();
    }
});

export default router
