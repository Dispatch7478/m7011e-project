<template>
  <nav class="navbar">
    <router-link to="/" class="nav-item">Home</router-link>
    <router-link to="/signup" class="nav-item">Sign Up</router-link>
    <router-link to="/login" class="nav-item">Log In</router-link>
  </nav>
</template>

<script>
export default {
  name: 'NavBar'
}
</script>

<style scoped>
.navbar {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  padding: 15px 20px;
  background-color: #333;
  color: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.nav-item {
  color: white;
  text-decoration: none;
  margin-right: 20px;
  font-weight: bold;
  padding: 5px 10px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.nav-item:hover {
  background-color: #555;
}
</style>