<template>
  <div class="tournament-container">
    <header class="hero-section">
      <h1>Tournaments</h1>
      <p>Browse public tournaments or create your own.</p>
    </header>

    <section class="action-cards">
      <div class="card">
        <h2>Public Tournaments</h2>

        <div v-if="tournaments.length === 0" class="tournaments-empty">
          <p>No public tournaments available at the moment.</p>
        </div>

        <div v-else class="tournaments-list">
          <div
            v-for="tournament in tournaments"
            :key="tournament.id"
            class="tournament-card"
          >
            <div class="tournament-main">
              <h3 class="tournament-name">
                {{ tournament.name }}
              </h3>
              <p class="tournament-meta">
                <span>{{ tournament.current_participants }} / {{ tournament.max_participants }} participants</span>
              </p>
            </div>
                <div class="tournament-side">
                  <button
                    type="button"
                    class="btn-link"
                    @click="viewBracket(tournament.id)"
                  >
                    View Bracket
                  </button>
                  <button
                    v-if="isLoggedIn && tournament.status === 'registration' && !isRegistered(tournament.id) && tournament.current_participants < tournament.max_participants"
                    type="button"
                    class="btn-link"
                    @click="registerForTournament(tournament.id)"
                  >
                    Register
                  </button>
                  <span v-if="isRegistered(tournament.id)">Registered</span>
                  <span v-else-if="tournament.current_participants >= tournament.max_participants" style="color: #dc3545; font-weight: bold;">Full</span>
                </div>
              </div>
        </div>
      </div>
    </section>

    <!-- Button to create a new tournament -->
    <section v-if="isLoggedIn" class="creation-section">
      <router-link to="/tournaments/create" class="btn primary-btn">
        Create a New Tournament
      </router-link>
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
      registrations: [],
      isLoggedIn: false,
    };
  },
  methods: {
    async getTournaments() {
      try {
        const response = await securedApi.get('/api/tournaments');
        this.tournaments = response.data || [];
      } catch (error) {
        console.error('Failed to fetch tournaments:', error);
        alert('Failed to fetch tournaments.');
      }
    },
    isRegistered(tournamentId) {
      return this.registrations.includes(tournamentId);
    },
    viewBracket(tournamentId) {
      this.$router.push({ name: 'Bracket', params: { id: tournamentId } });
    },
  },
  created() {
    this.isLoggedIn = this.$keycloak && this.$keycloak.authenticated;
    this.getTournaments();
  },
};
</script>

<style scoped>
.tournament-container {
  padding: 0;
  text-align: center;
}

.hero-section {
  background-color: #007bff;
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
  background-color: #28a745;
  color: white;
}

.primary-btn:hover {
  background-color: #1e7e34;
}

form div {
  margin-bottom: 15px;
}

label {
  display: block;
  margin-bottom: 5px;
}

input[type='text'],
select {
  width: 100%;
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
}

/* New styles from ProfilePage.vue */
.tournaments-list {
  margin-top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.tournaments-empty {
  margin-top: 1rem;
  padding: 0.75rem 0.9rem;
  border-radius: 10px;
  background-color: #f9fafb;
  border: 1px dashed #d1d5db;
  font-size: 0.9rem;
  color: #6b7280;
}

.tournament-card {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.9rem 1rem;
  border-radius: 10px;
  border: 1px solid #e5e7eb;
  background-color: #f9fafb;
  align-items: center;
  flex-wrap: wrap;
}

.tournament-main {
  flex: 1;
  min-width: 220px;
}

.tournament-name {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: #111827;
}

.tournament-meta {
  margin: 0.2rem 0 0;
  font-size: 0.85rem;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.tournament-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.35rem;
  min-width: 130px;
}

.btn-link {
  background: transparent;
  border: none;
  padding: 0;
  font-size: 0.85rem;
  font-weight: 500;
  color: #2563eb;
  border-radius: 999px;
}

.btn-link:hover {
  text-decoration: underline;
}

@media (max-width: 640px) {
  .tournament-card {
    align-items: flex-start;
  }

  .tournament-side {
    align-items: flex-start;
  }
}
</style>