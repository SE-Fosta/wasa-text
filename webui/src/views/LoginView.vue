<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import api from '../services/api';

const username = ref("");
const router = useRouter();
const errorMsg = ref("");

async function doLogin() {
    try {
        if (username.value.length < 3) return;
        const response = await api.login(username.value);

        // Salva l'identifier (l'ID utente) nel localStorage
        localStorage.setItem("token", response.data.identifier);
        localStorage.setItem("username", username.value); // Utile per la UI

        router.push("/");
    } catch (e) {
        errorMsg.value = "Errore durante il login: " + e.message;
    }
}
</script>

<template>
    <div class="login-container">
        <h1>WASAText Login</h1>
        <input v-model="username" placeholder="Inserisci il tuo username..." />
        <button @click="doLogin">Entra</button>
        <p v-if="errorMsg" class="error">{{ errorMsg }}</p>
    </div>
</template>

<style scoped>
/* Aggiungi qui il CSS per centrare il login */
.login-container { display: flex; flex-direction: column; align-items: center; margin-top: 100px; }
input { padding: 10px; margin: 10px; font-size: 16px; }
</style>
