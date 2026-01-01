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
              <h3 class="tournament-name">{{ tournament.name }}</h3>
              
              <div class="tournament-meta">
                <span>{{ tournament.game }}</span>
                <span class="separator">•</span>
                <span>{{ tournament.format }}</span>
                <span class="separator">•</span>
                <span class="badge" :class="tournament.participant_type">
                  {{ tournament.participant_type }}
                </span>
                <span class="separator">•</span>
                <span class="tournament-date">{{ formatDate(tournament.start_date) }}</span>
                <span class="separator">•</span>
                <span class="tournament-status" :class="`status-${tournament.status.toLowerCase().replace('_', '-')}`">
                  {{ tournament.status.replace(/_/g, ' ') }}
                </span>
              </div>
              
              <div class="participant-count">
                <small>{{ tournament.current_participants }} / {{ tournament.max_participants }} registered</small>
              </div>
            </div>
                <div class="tournament-actions">
                  <!-- Temporary until the page to edit specific tournaments is ready -->
                  <button
                    v-if="isOrganizer(tournament) && tournament.status === 'registration_closed'"
                    type="button"
                    class="btn-link"
                    @click="generateBracket(tournament)"
                  >
                    Generate Bracket
                  </button>

                  <router-link
                    v-if="($keycloak && $keycloak.authenticated && $keycloak.hasRealmRole('SuperAdmin')) || (isOrganizer(tournament))"
                    :to="{ name: 'ChangeTournament', params: { id: tournament.id } }"
                    class="btn-link"
                  >
                    Settings
                  </router-link>
                  
                  <button
                    v-if="['ongoing', 'completed'].includes(tournament.status)"
                    type="button"
                    class="btn-link"
                    @click="viewBracket(tournament.id)"
                  >
                    View Bracket
                  </button>

                  <button
                    v-if="isLoggedIn && tournament.status === 'registration_open' 
                    && !isRegistered(tournament.id) 
                    && tournament.current_participants < tournament.max_participants"
                    type="button"
                    class="btn-link"
                    @click="initiateRegistration(tournament)"
                  >
                    Register
                  </button>

                  <span v-if="isRegistered(tournament.id)" class="status-text registered">Registered</span>
                  <span v-else-if="tournament.status === 'registration_closed'" class="status-text closed">Registrations Closed</span>
                  <span v-else-if="tournament.current_participants >= tournament.max_participants" class="status-text full">Full</span>
                </div>
              </div>
        </div>
      </div>
    </section>

    <section class="action-cards">
      <div class="card">
        <h2>Previous Tournaments</h2>

        <div v-if="previousTournaments.length === 0" class="tournaments-empty">
          <p>No previous tournaments available.</p>
        </div>

        <div v-else class="tournaments-list">
          <div
            v-for="tournament in previousTournaments"
            :key="tournament.id"
            class="tournament-card"
          >
            <div class="tournament-main">
              <h3 class="tournament-name">{{ tournament.name }}</h3>
              
              <div class="tournament-meta">
                <span>{{ tournament.game }}</span>
                <span class="separator">•</span>
                <span>{{ tournament.format }}</span>
                <span class="separator">•</span>
                <span class="badge" :class="tournament.participant_type">
                  {{ tournament.participant_type }}
                </span>
                <span class="separator">•</span>
                <span class="tournament-date">{{ formatDate(tournament.start_date) }}</span>
                <span class="separator">•</span>
                <span class="tournament-status" :class="`status-${tournament.status.toLowerCase().replace('_', '-')}`">
                  {{ tournament.status.replace(/_/g, ' ') }}
                </span>
              </div>
              
              <div class="participant-count">
                <small>{{ tournament.current_participants }} / {{ tournament.max_participants }} registered</small>
              </div>
            </div>
            <div class="tournament-actions">
              <button
                v-if="['ongoing', 'completed'].includes(tournament.status)"
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

    <!-- Button to create a new tournament -->
    <section v-if="isLoggedIn" class="creation-section">
      <router-link to="/tournaments/create" class="btn primary-btn">
        Create a New Tournament
      </router-link>
    </section>

    <div v-if="showTeamModal" class="modal-overlay">
      <div class="modal-content">
        <h3>Select Team</h3>
        <p>This is a team tournament. Please select which team you want to register.</p>
        
        <div v-if="myTeams.length > 0" class="team-list">
          <div 
            v-for="team in myTeams" 
            :key="team.id" 
            class="team-option"
            @click="registerTeam(team)"
          >
            <span class="team-tag">[{{ team.tag }}]</span>
            <span class="team-name">{{ team.name }}</span>
          </div>
        </div>
        <div v-else class="empty-teams">
          <p>You are not the captain of any teams.</p>
          <router-link to="/teams/create" class="btn-link">Create a Team</router-link>
        </div>

        <button @click="closeTeamModal" class="btn-cancel">Cancel</button>
      </div>
    </div>
  </div>
</template>

<script>
import securedApi from '@/axios-auth.js';

export default {
  name: 'TournamentPage',
  data() {
    return {
      tournaments: [],
      previousTournaments: [],
      registrations: [],
      isLoggedIn: false,
      showTeamModal: false,
      myTeams: [],
      selectedTournament: null,
    };
  },
  methods: {
    async getTournaments() {
      try {
        const response = await securedApi.get('/api/tournaments');
        const allTournaments = response.data || [];
        const now = new Date();

        this.tournaments = allTournaments.filter(t => new Date(t.start_date) >= now);
        this.previousTournaments = allTournaments.filter(t => new Date(t.start_date) < now);

      } catch (error) {
        console.error('Failed to fetch tournaments:', error);
        alert('Failed to fetch tournaments.');
      }
    },
    isRegistered(tournamentId) {
      return this.registrations.includes(tournamentId);
    },
    // Logic to decide if we show modal or register directly
    async initiateRegistration(tournament) {
      if (tournament.participant_type === 'team') {
        this.selectedTournament = tournament;
        await this.fetchMyCaptainTeams();
        this.showTeamModal = true;
      } else {
        await this.registerIndividual(tournament);
      }
    },
    // Fetch teams where I am captain
    async fetchMyCaptainTeams() {
      try {
        const response = await securedApi.get('/api/teams/me/teams/captain');
        this.myTeams = response.data || [];
      } catch (error) {
        console.error("Failed to fetch teams:", error);
        this.myTeams = [];
      }
    },
    async registerTeam(team) {
      if (!this.selectedTournament) return;
      
      const payload = {
        name: team.name, 
        team_id: team.id 
      };

      await this.submitRegistration(this.selectedTournament, payload);
      this.closeTeamModal();
    },
    async registerIndividual(tournament) {
      const payload = {
        name: this.$keycloak.tokenParsed.preferred_username || "Unknown"
      };
      await this.submitRegistration(tournament, payload);
    },
    // Shared submission logic
    async submitRegistration(tournament, payload) {
      try {
        await securedApi.post(`/api/tournaments/${tournament.id}/register`, payload);

        this.registrations.push(tournament.id);
        
        // Optimistic UI update
        const tIndex = this.tournaments.findIndex(t => t.id === tournament.id);
        if (tIndex !== -1) {
          this.tournaments[tIndex].current_participants += 1;
        }
        
        alert(`Successfully registered for ${tournament.name}!`);
      } catch (error) {
        console.error('Registration failed:', error);
        const msg = error.response?.data?.error || 'Failed to register.';
        alert(msg);
      }
    },

    closeTeamModal() {
      this.showTeamModal = false;
      this.selectedTournament = null;
    },
    viewBracket(tournamentId) {
        this.$router.push({ name: 'Bracket', params: { id: tournamentId } });
      },
    async updateTournamentStatus(tournamentId, newStatus) {
      try {
        await securedApi.patch(`/api/tournaments/${tournamentId}/status`, {
          status: newStatus
        });
        alert("Status updated successfully!");
        this.getTournaments(); // Refresh the list to show updated status
      } catch (error) {
        console.error("Update failed:", error);
        const msg = error.response?.data?.error || 'Failed to update tournament status.';
        alert(`Status update failed: ${msg}`);
      }
    },
    isOrganizer(tournament) {
      // 1. Check if user is logged in
      if (!this.isLoggedIn || !this.$keycloak.tokenParsed) {
        return false;
      }
      
      // 2. Get current User ID (subject) from Keycloak
      const currentUserId = this.$keycloak.tokenParsed.sub;
      
      // 3. Compare with Tournament Organizer
      return currentUserId === tournament.organizer_id;
    },
    async generateBracket(tournament) {
      if (!confirm(`Are you sure you want to generate the bracket for "${tournament.name}"? This action cannot be undone.`)) {
        return;
      }

      try {
        // 1. Call Bracket Service
        // Note: We use query param ?tournament_id=... to match your backend handler
        await securedApi.post(`/api/brackets/generate?tournament_id=${tournament.id}`);
        
        alert("Bracket generated successfully!");

        // 2. Auto-start the tournament (Update status to 'ongoing')
        // This makes the "View Bracket" button visible immediately.
        await this.updateTournamentStatus(tournament.id, 'ongoing');

      } catch (error) {
        console.error("Bracket generation failed:", error);
        const msg = error.response?.data?.error || "Failed to generate bracket.";
        alert(`Error: ${msg}`);
      }
    },
    formatDate(dateString) {
      const date = new Date(dateString);
      return date.toISOString().split('T')[0];
    },
  },
  created() {
    this.isLoggedIn = this.$keycloak && this.$keycloak.authenticated;
    this.getTournaments();
  },
};
</script>

<style scoped>
/* Container & Layout */
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

/* Creation Section */
.creation-section {
  margin: 40px 20px;
}

/* Button Base Styles */
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

/* Form Styles */
form div { margin-bottom: 15px; }
label { display: block; margin-bottom: 5px; }
input[type='text'], select {
  width: 100%; padding: 8px; border: 1px solid #ccc; border-radius: 4px;
}

/* --- TOURNAMENT LIST STYLING --- */

.tournaments-list {
  margin-top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem; /* Increased gap between cards */
}

.tournaments-empty {
  margin-top: 1rem;
  padding: 1rem;
  border-radius: 10px;
  background-color: #f9fafb;
  border: 1px dashed #d1d5db;
  color: #6b7280;
  text-align: center;
}

.tournament-card {
  display: flex;
  justify-content: space-between; /* Pushes content to edges */
  align-items: center;            /* Vertically centers everything */
  padding: 1.25rem;
  border-radius: 10px;
  border: 1px solid #e5e7eb;
  background-color: #ffffff;      /* Cleaner white background */
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  gap: 1.5rem;
}

/* Left Side: Info */
.tournament-info-main {
  display: flex;
  flex-direction: column;
  gap: 0.25rem; /* Space between the Title row and the Count row */
  flex-grow: 1; /* Allows this section to take up space */
}

/* The Row containing Name + Metadata */
.tournament-header {
  display: flex;
  align-items: baseline; /* Aligns text cleanly on the baseline */
  flex-wrap: wrap; 
  gap: 1.5rem; /* Creates the separation between Name and Meta */
}

.tournament-name {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 700;
  color: #111827;
  line-height: 1.2;
}

.tournament-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: #6b7280;
  background-color: #f3f4f6; /* Optional: Subtle bubble for meta */
  padding: 4px 10px;
  border-radius: 6px;
}

/* Right Side: Actions */
.tournament-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-shrink: 0; /* Prevents buttons from squishing */
}

/* Badges & Status */
.separator { color: #d1d5db; margin: 0 2px; }

.badge {
  font-weight: 600;
  text-transform: uppercase;
  font-size: 0.7rem;
  letter-spacing: 0.05em;
}

/* Specific colors for types */
.badge.individual { color: #0284c7; }
.badge.team { color: #7e22ce; }

.participant-count {
  font-size: 0.8rem;
  color: #9ca3af;
  margin-top: 2px;
}

.status-text {
  font-size: 0.9rem;
  font-weight: 600;
}
.status-text.full { color: #dc3545; }
.status-text.registered { color: #198754; }
.status-text.closed {
  color: #ffc107; /* A warning/neutral color, like orange/yellow */
}

/* Added styles for Start Date and Status */
.tournament-date {
  font-size: 0.85rem;
  color: #6b7280;
  font-weight: 500;
}

.tournament-status {
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
  text-transform: capitalize;
  font-size: 0.8rem;
}

.tournament-status.status-draft { background-color: #e0e0e0; color: #424242; }
.tournament-status.status-registration-open { background-color: #d4edda; color: #155724; }
.tournament-status.status-registration-closed { background-color: #fff3cd; color: #856404; }
.tournament-status.status-ongoing { background-color: #cce5ff; color: #004085; }
.tournament-status.status-completed { background-color: #d1ecf1; color: #0c5460; }
.tournament-status.status-cancelled { background-color: #f8d7da; color: #721c24; }

/* Link Buttons */
.btn-link {
  background: transparent;
  border: 1px solid #007bff;
  color: #007bff;
  padding: 6px 16px;
  border-radius: 6px;
  font-weight: 600;
  font-size: 0.85rem;
  transition: all 0.2s;
}

.btn-link:hover {
  background-color: #007bff;
  color: white;
  text-decoration: none;
}

/* Mobile Responsiveness */
@media (max-width: 768px) {
  .tournament-card {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  
  .tournament-header {
    gap: 0.5rem;
    flex-direction: column;
    align-items: flex-start;
  }
  
  .tournament-actions {
    width: 100%;
    justify-content: flex-end; /* Align buttons right on mobile too */
    padding-top: 1rem;
    border-top: 1px solid #f3f4f6;
  }
}

/* Modal Styles */
.modal-overlay {
  position: fixed; top: 0; left: 0; width: 100%; height: 100%;
  background: rgba(0,0,0,0.5);
  display: flex; justify-content: center; align-items: center;
  z-index: 1000;
}
.modal-content {
  background: white; padding: 30px; border-radius: 8px;
  width: 400px; text-align: center;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}
.team-list {
  margin: 20px 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 300px;
  overflow-y: auto;
}
.team-option {
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
  text-align: left;
  display: flex;
  align-items: center;
  gap: 10px;
}
.team-option:hover {
  background-color: #f0f9ff;
  border-color: #007bff;
}
.team-tag {
  font-weight: bold;
  color: #555;
  font-family: monospace;
}
.team-name {
  font-weight: 600;
}
.empty-teams {
  margin: 20px 0;
  color: #666;
}
.btn-cancel {
  margin-top: 10px;
  background-color: #e0e0e0;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
}
</style>