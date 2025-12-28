<template>
  <div class="user-list">
    <h2>Users</h2>
    <div v-if="loading">Loading...</div>
    <div v-if="error">{{ error }}</div>
    <ul v-if="users.length">
      <li v-for="user in users" :key="user.id">
        {{ user.username }} ({{ user.email }})
      </li>
    </ul>
  </div>
</template>

<script>
import api from '../axios-auth';

export default {
  name: 'UserList',
  data() {
    return {
      users: [],
      loading: false,
      error: null,
    };
  },
  created() {
    this.fetchUsers();
  },
  methods: {
    fetchUsers() {
      this.loading = true;
      api.get('/api/v1/users')
        .then(response => {
          this.users = response.data;
        })
        .catch(error => {
          this.error = 'Failed to fetch users';
          console.error(error);
        })
        .finally(() => {
          this.loading = false;
        });
    },
  },
};
</script>

<style scoped>
.user-list {
  margin-top: 20px;
}
</style>
