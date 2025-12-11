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
                <span>{{ tournament.size }} teams</span>
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
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Creation form for logged-in users -->
    <section v-if="isLoggedIn" class="creation-section">
      <div class="card">
        <h2>Create a New Tournament</h2>

        <form @submit.prevent="createTournament">
          <div>
            <label for="name">Tournament Name:</label>
            <input
              id="name"
              type="text"
              v-model="tournament.name"
              required
            />
          </div>

          <div>
            <label for="type">Type:</label>
            <select
              id="type"
              v-model="tournament.type"
              required
            >
              <option value="public">Public</option>
              <option value="private">Private</option>
            </select>
          </div>

          <div>
            <label for="teams">Number of Teams:</label>
            <select
              id="teams"
              v-model="tournament.teams"
              required
            >
              <option :value="16">16</option>
              <option :value="32">32</option>
            </select>
          </div>

          <button
            type="submit"
            class="btn primary-btn"
          >
            Create Tournament
          </button>
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
        teams: 16,
      },
    };
  },
  methods: {
    async getPublicTournaments() {
      this.tournaments = [
        {
          id: 1,
          name: 'Summer League',
          size: 16,
          date: '2025-12-05 19:00',
          location: 'Online',
        },
        {
          id: 2,
          name: 'Winter Cup',
          size: 32,
          date: '2025-12-18 20:30',
          location: 'Local Arena',
        },
      ];
      // Real API call example:
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

        this.tournament = {
          name: '',
          type: 'public',
          teams: 16,
        };

        alert('Tournament created successfully!');
      } catch (error) {
        console.error('Failed to create tournament:', error);
        alert('Failed to create tournament.');
      }
    },
    viewBracket(tournamentId) {
      this.$router.push({ name: 'Bracket', params: { id: tournamentId } });
    },
  },
  created() {
    this.isLoggedIn = this.$keycloak && this.$keycloak.authenticated;
    this.getPublicTournaments();
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