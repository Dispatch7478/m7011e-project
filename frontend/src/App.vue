<template>
  <div id="app-container">
    <NavBar /> <router-view />
  </div>
</template>

<script>
// NOTE: Assuming you move the NavBar component to the root of the src/components folder
import NavBar from './components/NavBar.vue'

export default {
  name: 'App',
  components: {
    NavBar
  }
}
</script>

<style>
/* Global Styles can go here (optional) */
body {
  margin: 0;
  font-family: Arial, sans-serif;
  background-color: #f4f4f4;
}

#app-container {
  /* This ensures the entire app takes up the available space */
  min-height: 100vh;
}
</style>