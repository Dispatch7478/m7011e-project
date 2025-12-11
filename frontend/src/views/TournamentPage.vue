<template>
  <div class="tournament-container">
    <header class="hero-section">
      <h1>Tournaments</h1>
      <p>Browse public tournaments or create your own.</p>
    </header>

    <section class="action-cards">
      <div class="card">
        <h2>Public Tournaments</h2>
        <ul v-if="tournaments.length > 0" class="tournament-list">
          <li v-for="tournament in tournaments" :key="tournament.id">
            <strong>{{ tournament.name }}</strong> - {{ tournament.size }} teams
          </li>
        </ul>
        <p v-else>No public tournaments available at the moment.</p>
      </div>
    </section>

    <!-- Creation form for logged-in users -->
    <section v-if="isLoggedIn" class="creation-section">
      <div class="card">
        <h2>Create a New Tournament</h2>
        <form @submit.prevent="createTournament">
          <div>
            <label for="name">Tournament Name:</label>
            <input type="text" id="name" v-model="tournament.name" required>
          </div>
          <div>
            <label for="type">Type:</label>
            <select id="type" v-model="tournament.type" required>
              <option value="public">Public</option>
              <option value="private">Private</option>
            </select>
          </div>
          <div>
            <label for="teams">Number of Teams:</label>
            <select id="teams" v-model="tournament.teams" required>
              <option value="16">16</option>
              <option value="32">32</option>
            </select>
          </div>
          <button type="submit" class="btn primary-btn">Create Tournament</button>
        </form>
      </div>
    </section>
  </div>
</template>

<script>
import securedApi from '@/axios-auth.js';

export default {
  name: 'TournamentPage',
  data() {
    return {
      tournaments: [],
      isLoggedIn: false,
      tournament: {
        name: '',
        type: 'public',
        teams: 16
      }
    };
  },
  methods: {
    async getPublicTournaments() {
      // Mock data for now
      this.tournaments = [
        { id: 1, name: 'Summer League', size: 16 },
        { id: 2, name: 'Winter Cup', size: 32 },
      ];
      // try {
      //   const response = await securedApi.get('/tournaments');
      //   this.tournaments = response.data;
      // } catch (error) {
      //   console.error('Failed to fetch tournaments:', error);
      // }
    },
    async createTournament() {
      try {
        const response = await securedApi.post('/tournaments', {
          name: this.tournament.name,
          private: this.tournament.type === 'private',
          size: this.tournament.teams,
        });
        this.tournaments.push(response.data);
        this.tournament.name = '';
        this.tournament.type = 'public';
        this.tournament.teams = 16;
        alert('Tournament created successfully!');
      } catch (error) {
        console.error('Failed to create tournament:', error);
        alert('Failed to create tournament.');
      }
    }
  },
  created() {
    this.isLoggedIn = this.$keycloak && this.$keycloak.authenticated;
    this.getPublicTournaments();
  }
}
</script>

<style scoped>
.tournament-container {
  padding: 0;
  text-align: center;
}
.hero-section {
  background-color: #007bff; /* Primary blue color */
  color: white;
  padding: 60px 20px;
  margin-bottom: 40px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}
.hero-section h1 {
  font-size: 2.5em;
  margin-bottom: 10px;
}
.action-cards {
  display: flex;
  justify-content: center;
  gap: 30px;
  margin: 0 20px 50px;
}
.card {
  flex: 1;
  max-width: 900px;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  background-color: white;
  text-align: left;
}
.creation-section {
  margin: 40px 20px;
}
.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 5px;
  font-weight: bold;
  cursor: pointer;
  margin-top: 15px;
  transition: background-color 0.3s;
}
.primary-btn {
  background-color: #28a745; /* Green for creation */
  color: white;
}
.primary-btn:hover {
  background-color: #1e7e34;
}
.tournament-list {
  list-style: none;
  padding: 0;
}
.tournament-list li {
  padding: 10px;
  border-bottom: 1px solid #eee;
}
form div {
  margin-bottom: 15px;
}
label {
  display: block;
  margin-bottom: 5px;
}
input[type="text"], select {
  width: 100%;
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
}
</style>
