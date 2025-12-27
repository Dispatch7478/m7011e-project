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
                  <button
                    v-if="isOrganizer(tournament)"
                    type="button"
                    class="btn-link"
                    @click="selectNewStatus(tournament)"
                  >
                    Change Status
                  </button>
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
                    @click="registerForTournament(tournament)"
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
        console.log("Full Response:", response);
        this.tournaments = response.data || [];
      } catch (error) {
        console.error('Failed to fetch tournaments:', error);
        alert('Failed to fetch tournaments.');
      }
    },
    isRegistered(tournamentId) {
      return this.registrations.includes(tournamentId);
    },
    async registerForTournament(tournament) {
      try {
        // Get Username directly from Keycloak Token
        // The tokenParsed object contains the claims from the JWT.
        const username = this.$keycloak.tokenParsed.preferred_username || "Unknown User";
        
        // Handle Team Logic (Temporarily Disabled)
        if (tournament.participant_type === 'team') {
          // Mirroring the Backend's 501 Not Implemented logic
          alert("Team registration is temporarily disabled while we update our team architecture.");
          return;
        } 

        // Individual Logic: the backend gets the ID from the token header and the name from the jwt claims.
        const payload = { };

        // Send Registration to Backend
        await securedApi.post(`/api/tournaments/${tournament.id}/register`, payload);

        // Update UI
        this.registrations.push(tournament.id);
        
        // Update the participant count locally so to avoid a re-fetch of the whole list
        const tIndex = this.tournaments.findIndex(t => t.id === tournament.id);
        if (tIndex !== -1) {
          this.tournaments[tIndex].current_participants += 1;
        }
        
        alert('Successfully registered!');

      } catch (error) {
        console.error('Registration failed:', error);
        // Display the specific error message from the backend (e.g., "Tournament is full")
        const msg = error.response?.data?.error || 'Failed to register.';
        alert(msg);
      }
    },
    viewBracket(tournamentId) {
        this.$router.push({ name: 'Bracket', params: { id: tournamentId } });
      },
    async selectNewStatus(tournament) {
      const validStatuses = ['draft', 'registration_open', 'registration_closed', 'ongoing', 'completed', 'cancelled'];
      const promptMessage = `Enter new status for "${tournament.name}".\n\nValid options: ${validStatuses.join(', ')}`;
      
      const newStatus = prompt(promptMessage, tournament.status);

      if (newStatus === null) {
        // User cancelled the prompt
        return;
      }

      if (newStatus.trim() === tournament.status) {
        // No change
        return;
      }

      if (validStatuses.includes(newStatus.trim())) {
        await this.updateTournamentStatus(tournament.id, newStatus.trim());
      } else {
        alert(`Invalid status: "${newStatus}". Please enter one of the valid options.`);
      }
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
    created() {
      if (this.$keycloak) {
        this.$keycloak.ready.then(authenticated => {
          this.isLoggedIn = authenticated;
          if (authenticated) {
            this.getTournaments();
          }
        });
      }
    },
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
</style>