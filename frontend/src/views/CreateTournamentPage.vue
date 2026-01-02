<template>
  <div class="tournament-container">
    <header class="hero-section">
      <h1>Create a New Tournament</h1>
      <p>Fill in the details below to create your tournament.</p>
    </header>

    <section class="creation-section">
      <div class="card">
        <h2>Tournament Details</h2>

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
            <label for="description">Description:</label>
            <textarea
              id="description"
              v-model="tournament.description"
            ></textarea>
          </div>

          <div>
            <label for="game">Game:</label>
            <input
              id="game"
              type="text"
              v-model="tournament.game"
              required
            />
          </div>

          <div>
            <label for="format">Format:</label>
            <select
              id="format"
              v-model="tournament.format"
              required
            >
              <option value="single-elimination">Single Elimination</option>
            </select>
          </div>

          <div>
            <label for="participantType">Participant Type:</label>
            <select
              id="participantType"
              v-model="tournament.participant_type"
              required
            >
              <option value="individual">1v1 (Individual)</option>
              <option value="team">Team Based</option>
            </select>
          </div>

          <div>
            <label for="startDate">Start Date:</label>
            <input
              id="startDate"
              type="datetime-local"
              v-model="tournament.start_date"
            />
          </div>

          <div>
            <label for="status">Status:</label>
            <select
              id="status"
              v-model="tournament.status"
              required
            >
              <option value="draft">Draft</option>
              <option value="registration">Registration</option>
              <option value="active">Active</option>
              <option value="completed">Completed</option>
            </select>
          </div>

          <div>
            <label for="minParticipants">Minimum Participants:</label>
            <input
              id="minParticipants"
              type="number"
              v-model="tournament.min_participants"
              min="2"
              required
            />
          </div>

          <div>
            <label for="maxParticipants">Maximum Participants:</label>
            <input
              id="maxParticipants"
              type="number"
              v-model="tournament.max_participants"
              min="2"
              required
            />
          </div>

          <div>
            <label for="public">
              <input
                id="public"
                type="checkbox"
                v-model="tournament.public"
              />
              Public Tournament
            </label>
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
  name: 'CreateTournamentPage',
  data() {
    return {
      tournament: {
        name: '',
        description: '',
        game: '',
        format: 'single-elimination',
        participant_type: 'individual',
        start_date: '',
        status: 'draft',
        min_participants: 2,
        max_participants: 16,
        public: true,
      },
    };
  },
  methods: {
    async createTournament() {
      try {
        await securedApi.post('/api/tournaments', {
          name: this.tournament.name,
          description: this.tournament.description,
          game: this.tournament.game,
          format: this.tournament.format,
          participant_type: this.tournament.participant_type,
          start_date: this.tournament.start_date ? new Date(this.tournament.start_date).toISOString() : '',
          status: this.tournament.status,
          min_participants: parseInt(this.tournament.min_participants),
          max_participants: parseInt(this.tournament.max_participants),
          public: this.tournament.public,
        });

        alert('Tournament created successfully!');
        this.$router.push({ name: 'Tournaments' });
      } catch (error) {
        console.error('Failed to create tournament:', error);
        alert('Failed to create tournament.');
      }
    },
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

.creation-section {
  display: flex;
  justify-content: center;
  margin: 40px 20px;
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
input[type='number'],
input[type='datetime-local'],
textarea,
select {
  width: 100%;
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
}
</style>
