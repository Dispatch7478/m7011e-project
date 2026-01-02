<template>
  <div class="change-tournament-container">
    <div class="card">
      <h1 class="card-header">Edit Tournament</h1>
      <div class="card-body">
        <form v-if="tournament" @submit.prevent="saveTournament">
          <div class="form-group">
            <label for="name">Tournament Name</label>
            <input type="text" id="name" v-model="tournament.name" required>
          </div>
          <div class="form-group">
            <label for="description">Description</label>
            <textarea id="description" v-model="tournament.description"></textarea>
          </div>
          <div class="form-group">
            <label for="game">Game</label>
            <input type="text" id="game" v-model="tournament.game" required>
          </div>
          <div class="form-group">
            <label for="format">Format</label>
            <input type="text" id="format" v-model="tournament.format" required>
          </div>
          <div class="form-group">
            <label for="status">Status</label>
            <select id="status" v-model="tournament.status">
              <option value="draft">Draft</option>
              <option value="registration_open">Registration Open</option>
              <option value="registration_closed">Registration Closed</option>
              <option value="ongoing">Ongoing</option>
              <option value="completed">Completed</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>
          <div class="form-group">
            <label for="start_date">Start Time</label>
            <input type="datetime-local" id="start_date" v-model="tournament.start_date">
          </div>
          <div class="form-row">
            <div class="form-group">
              <label for="min_participants">Min Teams</label>
              <input type="number" id="min_participants" v-model.number="tournament.min_participants" min="2">
            </div>
            <div class="form-group">
              <label for="max_participants">Max Teams</label>
              <input type="number" id="max_participants" v-model.number="tournament.max_participants" min="2">
            </div>
          </div>
          <div class="form-actions">
            <button type="submit" class="btn primary-btn">Save Changes</button>
            <router-link to="/tournaments" class="btn-cancel">Cancel</router-link>
          </div>
        </form>
        <div v-else class="loading">
          Loading tournament details...
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import securedApi from '@/axios-auth.js';

export default {
  name: 'ChangeTournamentPage',
  data() {
    return {
      tournament: null,
    };
  },
  methods: {
    async fetchTournament() {
      const tournamentId = this.$route.params.id;
      try {
        const response = await securedApi.get(`/api/tournaments/${tournamentId}`);
        this.tournament = response.data;
        // Format date for datetime-local input
        if (this.tournament.start_date) {
            const date = new Date(this.tournament.start_date);
            const year = date.getFullYear();
            const month = (date.getMonth() + 1).toString().padStart(2, '0');
            const day = date.getDate().toString().padStart(2, '0');
            const hours = date.getHours().toString().padStart(2, '0');
            const minutes = date.getMinutes().toString().padStart(2, '0');
            this.tournament.start_date = `${year}-${month}-${day}T${hours}:${minutes}`;
        }
      } catch (error) {
        console.error('Failed to fetch tournament details:', error);
        alert('Failed to load tournament details. Please try again.');
      }
    },
    async saveTournament() {
      const tournamentId = this.$route.params.id;
      
      // Ensure numbers are not negative
      if (this.tournament.min_participants < 2 || this.tournament.max_participants < 2) {
          alert("Minimum and maximum participants must be at least 2.");
          return;
      }
      // Ensure min is not greater than max
      if (this.tournament.min_participants > this.tournament.max_participants) {
          alert("Minimum participants cannot be greater than maximum participants.");
          return;
      }

      // Convert date back to ISO format before sending
      const payload = {
        ...this.tournament,
        start_date: new Date(this.tournament.start_date).toISOString(),
      };

      // Backend check prevents changing game/format if tournament has started
      // Assume if the current status is ongoing or completed then do not send game/format.
      if (['ongoing', 'completed'].includes(this.tournament.status)) {
        // Prevent sending game and format in the payload if the tournament has started/completed
        delete payload.game;
        delete payload.format;
      }

      try {
        await securedApi.put(`/api/tournaments/${tournamentId}`, payload);
        alert('Tournament updated successfully!');
        this.$router.push('/tournaments');
      } catch (error) {
        console.error('Failed to update tournament:', error);
        const msg = error.response?.data?.error || 'An error occurred while saving.';
        alert(`Save failed: ${msg}`);
      }
    },
  },
  created() {
    this.fetchTournament();
  },
};
</script>

<style scoped>
.change-tournament-container {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 40px 20px;
  background-color: #f4f7f6;
  min-height: 100vh;
}

.card {
  width: 100%;
  max-width: 700px;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  overflow: hidden;
}

.card-header {
  background-color: #007bff;
  color: white;
  padding: 20px 25px;
  margin: 0;
  font-size: 1.5em;
}

.card-body {
  padding: 30px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #333;
}

.form-group input,
.form-group textarea,
.form-group select {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 1em;
}

.form-group textarea {
  min-height: 100px;
  resize: vertical;
}

.form-row {
  display: flex;
  gap: 20px;
}

.form-row .form-group {
  flex: 1;
}

.form-actions {
  margin-top: 30px;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 15px;
}

.btn {
  padding: 10px 20px;
  border-radius: 5px;
  font-weight: bold;
  cursor: pointer;
  border: none;
  transition: background-color 0.3s;
}

.primary-btn {
  background-color: #007bff;
  color: white;
}

.primary-btn:hover {
  background-color: #0056b3;
}

.btn-cancel {
  text-decoration: none;
  color: #6c757d;
  font-weight: bold;
}

.btn-cancel:hover {
  text-decoration: underline;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #666;
}
</style>