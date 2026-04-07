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
import api from '../services/axios'; 

const username = ref('');
const errorMessage = ref('');
const router = useRouter();

const doLogin = async () => {
  errorMessage.value = '';

  try {
    // Chiama il tuo endpoint di login nel backend Go
    const response = await api.post('/session', {
      name: username.value
    });


    // Il backend ci restituisce l'ID utente (es. "1" o 1)
    const token = response.data.identifier;

    // Salviamo l'ID nel browser
    localStorage.setItem('token', token);
    localStorage.setItem('username', username.value);

    // Andiamo alla Home
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