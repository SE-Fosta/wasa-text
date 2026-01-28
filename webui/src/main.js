import { createApp } from 'vue'
import './style.css' // Importa il CSS globale
import App from './App.vue'
import router from './router' // Importa il router configurato nel passo precedente

// Crea l'applicazione Vue partendo dal componente root App.vue
const app = createApp(App)

// Dice all'app di usare il router per la navigazione
app.use(router)

// Monta l'applicazione nel div con id="app" che si trova in index.html
app.mount('#app')
