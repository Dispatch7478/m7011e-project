<template>
  <div class="bracket-page">
    <h1>Bracket for Tournament {{ $route.params.id }}</h1>
    <p>This page will display the tournament bracket.</p>
  </div>
</template>

<script>
export default {
  name: 'BracketPage',
  mounted() {
    console.log('Tournament ID:', this.$route.params.id);
  }
}
</script>

<style scoped>
.bracket-page {
  padding: 20px;
  text-align: center;
}
</style>
