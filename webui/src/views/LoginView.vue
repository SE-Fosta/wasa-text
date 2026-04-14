<template>
  <div class="login-container">
    <h1>WASAText</h1>
    <p>Inserisci il tuo username per iniziare a chattare</p>
    
    <form @submit.prevent="doLogin">
      <input 
        type="text" 
        v-model="username" 
        placeholder="Es: Mario" 
        required 
        minlength="3" 
        maxlength="16"
      />
      <button type="submit">Entra</button>
    </form>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
// Importiamo l'istanza di axios configurata
import api from '../services/axios.js'; 

const username = ref('');
const errorMessage = ref('');
const router = useRouter();

const doLogin = async () => {
  errorMessage.value = '';

  try {
    const response = await api.post('/session', {
      name: username.value
    });

    // 1. STAMPIAMO LA VERITÀ: Vediamo esattamente cosa dice Go
    console.log("Risposta grezza dal server:", response.data);

    // 2. LA PAROLA MAGICA: usiamo "identifier" perché il tuo Go usa quella!
    const token = response.data.identifier;

    // 3. IL BLOCCO DI SICUREZZA
    if (!token || token === 'undefined') {
        alert("ERRORE: Non riesco a leggere l'ID. Guarda la console (F12)!");
        console.error("Il backend non ha inviato 'identifier'. Ha inviato:", response.data);
        return; // Ci fermiamo qui per non infettare il browser!
    }

    // 4. Salvataggio sicuro e reindirizzamento
    console.log("Login perfetto! Salvo il token:", token);
    localStorage.setItem('token', token);
    localStorage.setItem('username', username.value);
    
    router.push('/');
    
  } catch (error) {
    if (error.response && error.response.data && error.response.data.message) {
      errorMessage.value = error.response.data.message;
    } else {
      errorMessage.value = "Errore di connessione al server.";
    }
  }
};
</script>

<style scoped>
.login-container {
  max-width: 400px;
  margin: 100px auto;
  text-align: center;
  font-family: sans-serif;
}
input {
  padding: 10px;
  width: 70%;
  margin-right: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
}
button {
  padding: 10px 20px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
button:hover {
  background-color: #45a049;
}
.error {
  color: red;
  margin-top: 15px;
}
</style>