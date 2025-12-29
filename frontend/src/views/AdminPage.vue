<template>
  <div class="admin-page">
    <h1>Admin Page</h1>
    <p>This page is only visible to SuperAdmins.</p>
    <UserList />
  </div>
</template>

<script>
import UserList from '../components/UserList.vue';

export default {
  name: 'AdminPage',
  components: {
    UserList,
  },
};
</script>

<style scoped>
.admin-page {
  padding: 20px;
}
</style>
