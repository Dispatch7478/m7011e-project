<template>
  <div class="team-page-root">
    <header class="hero-section">
      <h1>My Teams</h1>
      <p>Manage your squads or join a new one.</p>
    </header>

    <div class="content-container">
      
      <div v-if="myInvites.length > 0" class="invites-section">
        <h2>Pending Invites</h2>
        <div class="invites-grid">
          <div v-for="invite in myInvites" :key="invite.id" class="invite-card">
            <div class="invite-info">
               <span class="invite-msg">You have been invited to join:</span>
               <div class="invite-team">
                 <strong>{{ invite.team_name }}</strong> 
                 <span class="tag">[{{ invite.team_tag }}]</span>
               </div>
            </div>
            <div class="invite-actions">
              <button @click="acceptInvite(invite)" class="btn-accept">Accept</button>
              <button @click="declineInvite(invite.id)" class="btn-decline">Decline</button>
            </div>
          </div>
        </div>
      </div>

      <div class="actions-bar">
        <router-link to="/teams/create" class="btn primary-btn">
          + Create New Team
        </router-link>
      </div>

      <div v-if="loading" class="loading-state">Loading teams...</div>

      <div v-else-if="myTeams.length === 0 && myInvites.length === 0" class="empty-state">
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
  </div>
</template>

<script>
import securedApi from '@/axios-auth.js';

export default {
  name: 'TeamListPage',
  data() {
    return {
      myTeams: [],
      myInvites: [],
      loading: true,
      currentUserId: null
    };
  },
  methods: {
    async fetchMyTeams() {
      try {
        this.loading = true;
        // Fetch teams where the user is a member (includes captaincy)
        let apiUrl = '/api/teams/me/teams';
        if (this.$keycloak && this.$keycloak.hasRealmRole('SuperAdmin')) {
          apiUrl = '/api/teams'; // Fetch all teams for SuperAdmin
        }
        const response = await securedApi.get(apiUrl);
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
    },
    async fetchMyInvites() {
      try {
        const response = await securedApi.get('/api/teams/me/invites');
        this.myInvites = response.data || [];
      } catch (error) {
        console.warn("Failed to fetch invites (backend might be missing endpoint):", error);
        this.myInvites = [];
      }
    },
    async acceptInvite(invite) {
      try {
        // Calls POST /teams/{team_id}/members
        await securedApi.post(`/api/teams/${invite.team_id}/members`, {
          invite_id: invite.id
        });
        alert(`You have joined ${invite.team_name}!`);
        // Refresh data
        this.myInvites = this.myInvites.filter(i => i.id !== invite.id);
        this.fetchMyTeams(); 
      } catch (error) {
        console.error("Failed to join team", error);
        alert(error.response?.data?.error || "Failed to accept invite.");
      }
    },
    async declineInvite(inviteId) {
      if (!confirm("Are you sure you want to decline this invite?")) return;
      try {
        await securedApi.delete(`/api/teams/invites/${inviteId}`);
        this.myInvites = this.myInvites.filter(i => i.id !== inviteId);
      } catch (error) {
        console.error("Failed to decline", error);
      }
    }
  },
  created() {
    if (this.$keycloak && this.$keycloak.tokenParsed) {
      this.currentUserId = this.$keycloak.tokenParsed.sub;
    }
    this.fetchMyTeams();
    this.fetchMyInvites();
  }
};
</script>

<style scoped>
/* Root container handles the full width background */
.team-page-root {
  width: 100%;
  min-height: 100vh;
  padding: 0;
  background-color: #fcfcfc;
}

/* Hero Section: Full Width, Blue Background (Matches TournamentPage) */
.hero-section {
  background-color: #007bff; /* Changed from Purple to Blue */
  color: white;
  padding: 60px 20px;
  margin-bottom: 40px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  text-align: center;
}

.hero-section h1 {
  font-size: 2.5em;
  margin-bottom: 10px;
  margin-top: 0;
}

/* Content Container: Centered, Constrained Width */
.content-container {
  padding: 0 20px 40px 20px;
  max-width: 1000px;
  margin: 0 auto;
  text-align: center;
}

.actions-bar { margin-bottom: 30px; display: flex; justify-content: flex-end; }

/* Invites Styles */
.invites-section { margin-bottom: 40px; text-align: left; }
.invites-section h2 { border-bottom: 2px solid #eee; padding-bottom: 10px; margin-bottom: 20px; color: #333; }
.invites-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 15px; }

.invite-card {
  background: #fff; border: 1px solid #cce5ff; border-left: 5px solid #007bff;
  border-radius: 5px; padding: 15px; display: flex; justify-content: space-between; align-items: center;
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}
.invite-info { display: flex; flex-direction: column; }
.invite-msg { font-size: 0.85rem; color: #666; margin-bottom: 4px; }
.invite-team { font-size: 1.1rem; }
.tag { font-family: monospace; font-weight: bold; color: #555; margin-left: 5px; }

.invite-actions { display: flex; gap: 10px; }
.btn-accept { background: #28a745; color: white; border: none; padding: 6px 12px; border-radius: 4px; cursor: pointer; font-weight: bold; }
.btn-accept:hover { background: #218838; }
.btn-decline { background: #dc3545; color: white; border: none; padding: 6px 12px; border-radius: 4px; cursor: pointer; font-weight: bold; }
.btn-decline:hover { background: #c82333; }

/* Button Styles */
.primary-btn {
  background-color: #28a745; color: white; padding: 10px 20px;
  text-decoration: none; border-radius: 5px; font-weight: bold;
}
.primary-btn:hover { background-color: #218838; }

/* Grid Styles */
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