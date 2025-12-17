<template>
  <nav class="navbar">
    <router-link to="/" class="nav-item">Home</router-link>
    <router-link to="/tournaments" class="nav-item">Tournaments</router-link>
    <router-link v-if="!$keycloak || !$keycloak.authenticated" to="/signup" class="nav-item">Sign Up</router-link>

    <router-link v-if="$keycloak && $keycloak.authenticated" to="/profile" class="nav-item">Profile</router-link>
    <!-- This ONLY routes to /login -->
    <router-link
      v-if="!$keycloak || !$keycloak.authenticated"
      to="/login"
      class="nav-item"
    >
      Log In
    </router-link>

    <!-- Logout stays here -->
    <button
      v-else
      @click="logout"
      class="nav-item"
    >
      Log Out
    </button>
  </nav>
</template>

<script>
export default {
  name: "NavBar",
  methods: {
    logout() {
      if (this.$keycloak?.authenticated) {
        this.$keycloak.logout({
          redirectUri: window.location.origin,
        })
      }
    },
  },
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