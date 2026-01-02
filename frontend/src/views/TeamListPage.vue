<template>
  <div class="page-container">
    <header class="hero-section">
      <h1>My Teams</h1>
      <p>Manage your squads or join a new one.</p>
    </header>

    <div class="actions-bar">
      <router-link to="/teams/create" class="btn primary-btn">
        + Create New Team
      </router-link>
    </div>

    <div v-if="loading" class="loading-state">Loading teams...</div>

    <div v-else-if="myTeams.length === 0" class="empty-state">
      <p>You are not a member of any team yet.</p>
    </div>

    <div v-else class="teams-grid">
      <div v-for="team in myTeams" :key="team.id" class="team-card">
        <div class="team-header">
          <div class="team-logo">
            <img :src="getTeamLogo(team)" alt="Logo" />
          </div>
          <div class="team-info">
            <h3>{{ team.name }}</h3>
            <span class="team-tag">[{{ team.tag }}]</span>
          </div>
        </div>
        
        <div class="team-role">
          <span v-if="team.captain_id === currentUserId" class="badge captain">Captain</span>
          <span v-else class="badge member">Member</span>
        </div>

        <div class="card-actions">
           <router-link :to="{ name: 'TeamDetail', params: { id: team.id }}" class="btn-link">
             Manage Team
           </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import securedApi from '@/axios-auth.js';

export default {
  name: 'TeamListPage',
  data() {
    return {
      myTeams: [],
      loading: true,
      currentUserId: null
    };
  },
  methods: {
    async fetchMyTeams() {
      try {
        this.loading = true;
        // Fetch teams where the user is a member (includes captaincy)
        const response = await securedApi.get('/api/teams/me/teams');
        this.myTeams = response.data || [];
      } catch (error) {
        console.error("Failed to fetch teams", error);
      } finally {
        this.loading = false;
      }
    },
    getTeamLogo(team) {
    if (team.logo_url) return team.logo_url;
    
    // Simple grey circle with "T" text placeholder
    return `data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='64' height='64' viewBox='0 0 64 64'><rect width='64' height='64' fill='%23e0e0e0'/><text x='50%' y='50%' font-family='Arial' font-size='24' fill='%23888' dy='.3em' text-anchor='middle'>T</text></svg>`;
  }
  },
  created() {
    if (this.$keycloak && this.$keycloak.tokenParsed) {
      this.currentUserId = this.$keycloak.tokenParsed.sub;
    }
    this.fetchMyTeams();
  }
};
</script>

<style scoped>
.page-container { padding: 0 20px; max-width: 1000px; margin: 0 auto; text-align: center; }
.hero-section { background: #6f42c1; color: white; padding: 40px 20px; border-radius: 0 0 10px 10px; margin-bottom: 30px; }
.actions-bar { margin-bottom: 30px; display: flex; justify-content: flex-end; }

.primary-btn {
  background-color: #28a745; color: white; padding: 10px 20px;
  text-decoration: none; border-radius: 5px; font-weight: bold;
}
.primary-btn:hover { background-color: #218838; }

.teams-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 20px; }

.team-card {
  background: white; border: 1px solid #e0e0e0; border-radius: 8px;
  padding: 20px; display: flex; flex-direction: column; align-items: center;
  box-shadow: 0 2px 5px rgba(0,0,0,0.05); transition: transform 0.2s;
}
.team-card:hover { transform: translateY(-3px); box-shadow: 0 5px 15px rgba(0,0,0,0.1); }

.team-header { display: flex; align-items: center; gap: 15px; margin-bottom: 15px; width: 100%; }
.team-logo img { width: 50px; height: 50px; border-radius: 50%; object-fit: cover; border: 1px solid #ddd; }
.team-info { text-align: left; }
.team-info h3 { margin: 0; font-size: 1.1rem; }
.team-tag { color: #666; font-family: monospace; font-weight: bold; }

.team-role { margin-bottom: 15px; width: 100%; text-align: left; }
.badge { padding: 4px 8px; border-radius: 4px; font-size: 0.8rem; font-weight: bold; text-transform: uppercase; }
.badge.captain { background-color: #ffd700; color: #856404; }
.badge.member { background-color: #e2e3e5; color: #383d41; }

.card-actions { border-top: 1px solid #f0f0f0; width: 100%; padding-top: 15px; }
.btn-link { color: #007bff; text-decoration: none; font-weight: 600; }
.btn-link:hover { text-decoration: underline; }

.loading-state, .empty-state { color: #888; margin-top: 40px; font-size: 1.1rem; }
</style>