import { createApp } from 'vue'
import { createPinia } from 'pinia' // Import
import App from './App.vue'
import './style.css'

const app = createApp(App)
app.use(createPinia()) // Use
app.mount('#app')