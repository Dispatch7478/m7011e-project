import { createApp } from 'vue'
import App from './App.vue'
import router from './router/index.js'

// Create the Vue application instance
const app = createApp(App)

// Tell the app to use the router
app.use(router)

// Mount the application to the #app element in your public/index.html
app.mount('#app')