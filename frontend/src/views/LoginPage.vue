<template>
  <div>
    <div class="login-container">
      <h2>Log In</h2>
      <p>Click the button below to start the Keycloak login process.</p>
      <button @click="handleLogin" class="login-button">
        Log In
      </button>
    </div>
  </div>
</template>

<script>
import NavBar from '@/components/NavBar.vue'

export default {
  name: 'LoginPage',
  components: {
    NavBar
  },
  methods: {
    handleLogin() {
      // **TODO:** Replace this with your actual Keycloak redirection logic
      alert('Simulating Keycloak redirection...')
      console.log('Login button clicked. Initiating Keycloak flow.')
    }
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px;
}

.login-button {
  padding: 10px 20px;
  font-size: 1.1em;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  margin-top: 20px;
}

.login-button:hover {
  background-color: #45a049;
}
</style>