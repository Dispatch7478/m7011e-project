<template>
  <div>
    <div class="signup-container">
      <h2>Create Your Account</h2>
      <form @submit.prevent="handleSignup" class="signup-form">
        <div class="form-group">
          <label for="name">Name:</label>
          <input type="text" id="name" v-model="form.name" required>
        </div>

        <div class="form-group">
          <label for="number">Phone Number:</label>
          <input type="tel" id="number" v-model="form.number">
        </div>

        <div class="form-group">
          <label for="email">Email:</label>
          <input type="email" id="email" v-model="form.email" required>
        </div>

        <div class="form-group">
          <label for="password">Password:</label>
          <input type="password" id="password" v-model="form.password" required>
        </div>

        <div class="form-group">
          <label for="confirmPassword">Confirm Password:</label>
          <input type="password" id="confirmPassword" v-model="form.confirmPassword" required>
        </div>

        <button type="submit" class="signup-button">Sign Up</button>
      </form>
    </div>
  </div>
</template>

<script>
import NavBar from '@/components/NavBar.vue'

export default {
  name: 'SignupPage',
  components: {
    NavBar
  },
  data() {
    return {
      form: {
        name: '',
        number: '',
        email: '',
        password: '',
        confirmPassword: ''
      }
    }
  },
  methods: {
    handleSignup() {
      if (this.form.password !== this.form.confirmPassword) {
        alert('Passwords do not match!')
        return
      }
      // **TODO:** Implement your API call for signup here
      console.log('Signup data:', this.form)
      alert('Signup form submitted! Check console for data.')
    }
  }
}
</script>

<style scoped>
.signup-container {
  max-width: 400px;
  margin: 40px auto;
  padding: 20px;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
}

.signup-form {
  display: flex;
  flex-direction: column;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

.form-group input {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  box-sizing: border-box;
}

.signup-button {
  padding: 10px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 1.1em;
}

.signup-button:hover {
  background-color: #0056b3;
}
</style>