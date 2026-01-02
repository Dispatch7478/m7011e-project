<template>
  <div class="team-detail-page">
    <div v-if="loading" class="loading-state">Loading team details...</div>
    
    <div v-else-if="!team" class="error-state">
      <p>Team not found or you do not have permission to view it.</p>
      <router-link to="/teams" class="btn-link">Back to My Teams</router-link>
    </div>
    
    <div v-else class="content-container">
      <header class="team-header-card">
        <div class="header-content">
          <img :src="getTeamLogo(team)" alt="Logo" class="team-logo-large"/>
          <div class="team-title">
            <h1>{{ team.name }}</h1>
            <span class="team-tag-large">[{{ team.tag }}]</span>
            <p class="meta">Created {{ formatDate(team.created_at) }}</p>
          </div>
        </div>
        
        <div class="header-actions">
           <button v-if="isCaptain" @click="deleteTeam" class="btn-danger-outline">Disband Team</button>
           
           <button v-if="!isCaptain" @click="leaveTeam" class="btn-warning-outline">Leave Team</button>
        </div>
      </header>

      <div class="grid-layout">
        <div class="column">
          <section class="panel">
            <div class="panel-header">
              <h3>Roster ({{ members.length }})</h3>
            </div>
            <div class="list-container">
              <div v-for="member in members" :key="member.user_id" class="list-item">
                <div class="member-info">
                   <span :class="['role-badge', member.role]">{{ member.role }}</span>
                   <span class="member-name">{{ formatUserId(member.user_id) }}</span>
                   <small class="join-date">{{ formatDate(member.joined_at) }}</small>
                </div>
                <div class="item-actions">
                  <button 
                    v-if="isCaptain && member.user_id !== currentUserId" 
                    @click="kickMember(member.user_id)" 
                    class="btn-icon delete"
                    title="Kick Member"
                  >
                    ✕
                  </button>
                </div>
              </div>
            </div>
          </section>
        </div>

        <div class="column" v-if="isCaptain">
          <section class="panel invite-section">
            <div class="panel-header">
              <h3>Invite Player</h3>
            </div>
            <div class="panel-body">
              <form @submit.prevent="sendInvite" class="invite-form">
                <div class="input-group">
                  <input 
                    v-model="newInviteEmail" 
                    type="email" 
                    placeholder="Player's email address..." 
                    required 
                  />
                  <button type="submit" :disabled="inviting" class="btn-primary">
                    {{ inviting ? 'Sending...' : 'Invite' }}
                  </button>
                </div>
                <p v-if="inviteMessage" :class="inviteError ? 'msg-error' : 'msg-success'">
                  {{ inviteMessage }}
                </p>
              </form>
            </div>
          </section>

          <section class="panel">
            <div class="panel-header">
              <h3>Pending Invites</h3>
            </div>
            <div v-if="invites.length === 0" class="empty-list">No pending invites.</div>
            <div v-else class="list-container">
              <div v-for="inv in invites" :key="inv.id" class="list-item">
                <div class="invite-info">
                  <span class="email">{{ inv.invitee_email }}</span>
                  <small class="expiry">Expires: {{ formatExpiry(inv.expires_at) }}</small>
                </div>
                <div class="item-actions">
                  <button @click="cancelInvite(inv.id)" class="btn-link-danger">Revoke</button>
                </div>
              </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import securedApi from '@/axios-auth.js';

export default {
  name: 'TeamDetailPage',
  data() {
    return {
      team: null,
      members: [],
      invites: [],
      loading: true,
      currentUserId: null,
      
      // Invite Form State
      newInviteEmail: '',
      inviting: false,
      inviteMessage: '',
      inviteError: false
    };
  },
  computed: {
    isCaptain() {
      return this.team && this.currentUserId === this.team.captain_id;
    }
  },
  methods: {
    getTeamLogo(team) {
      if (team && team.logo_url) return team.logo_url;
      // Consistent fallback SVG
      return `data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='64' height='64' viewBox='0 0 64 64'><rect width='64' height='64' fill='%23e0e0e0'/><text x='50%' y='50%' font-family='Arial' font-size='24' fill='%23888' dy='.3em' text-anchor='middle'>T</text></svg>`;
    },
    async loadData() {
      this.loading = true;
      const teamId = this.$route.params.id;
      
      try {
        // 1. Fetch Team Details via "My Teams" to ensure permissions
        const myTeamsRes = await securedApi.get('/api/teams/me/teams');
        const allTeams = myTeamsRes.data || [];
        this.team = allTeams.find(t => t.id === teamId);

        if (!this.team) {
           throw new Error("Team not found in your list.");
        }

        // 2. Fetch Members (Fixed Route: /api/teams/{id}/members)
        const membersRes = await securedApi.get(`/api/teams/${teamId}/members`);
        this.members = membersRes.data || [];

        // 3. Fetch Invites (Only if Captain)
        if (this.currentUserId === this.team.captain_id) {
           this.fetchInvites();
        }

      } catch (err) {
        console.error("Failed to load team data", err);
        this.team = null;
      } finally {
        this.loading = false;
      }
    },
    async fetchInvites() {
       try {
         // Fixed Route: /api/teams/{id}/invites
         const res = await securedApi.get(`/api/teams/${this.team.id}/invites`);
         this.invites = res.data || [];
       } catch (e) { 
         console.error("Failed to fetch invites", e); 
       }
    },
    async sendInvite() {
       this.inviting = true;
       this.inviteMessage = '';
       this.inviteError = false;
       try {
         // Fixed Route: /api/teams/{id}/invites
         await securedApi.post(`/api/teams/${this.team.id}/invites`, {
           invitee_email: this.newInviteEmail,
           expires_at: null 
         });
         this.inviteMessage = 'Invite sent successfully!';
         this.newInviteEmail = '';
         this.fetchInvites(); 
       } catch (err) {
         this.inviteError = true;
         this.inviteMessage = err.response?.data?.error || 'Failed to send invite.';
       } finally {
         this.inviting = false;
       }
    },
    async kickMember(userId) {
      if(!confirm("Are you sure you want to kick this member?")) return;
      try {
        await securedApi.delete(`/api/teams/${this.team.id}/members/${userId}`);
        this.members = this.members.filter(m => m.user_id !== userId);
      } catch (e) { alert("Failed to kick member"); }
    },
    async cancelInvite(inviteId) {
      if(!confirm("Revoke this invite?")) return;
      try {
         await securedApi.delete(`/api/teams/invites/${inviteId}`);
         this.invites = this.invites.filter(i => i.id !== inviteId);
      } catch (e) { alert("Failed to revoke invite"); }
    },
    async leaveTeam() {
      if(!confirm("Are you sure you want to leave this team?")) return;
      try {
        await securedApi.post(`/api/teams/${this.team.id}/leave`);
        this.$router.push('/teams');
      } catch (e) { 
        alert("Failed to leave team. Captains must delete the team or transfer ownership."); 
      }
    },
    async deleteTeam() {
      if(!confirm("WARNING: This will permanently delete the team and all data. Continue?")) return;
       try {
        await securedApi.delete(`/api/teams/${this.team.id}`);
        this.$router.push('/teams');
      } catch (e) { alert("Failed to delete team"); }
    },
    formatDate(d) {
      if(!d) return '';
      return new Date(d).toLocaleDateString();
    },
    formatExpiry(d) {
       if(!d) return 'Never';
       return new Date(d).toLocaleDateString();
    },
    formatUserId(id) {
       return 'User...' + id.substring(0, 6);
    }
  },
  created() {
    if (this.$keycloak?.tokenParsed) {
      this.currentUserId = this.$keycloak.tokenParsed.sub;
    }
    this.loadData();
  }
}
</script>

<style scoped>
.team-detail-page { max-width: 1000px; margin: 0 auto; padding: 20px; }

/* Header */
.team-header-card {
  background: white; border-radius: 10px; padding: 30px;
  display: flex; justify-content: space-between; align-items: center;
  box-shadow: 0 4px 10px rgba(0,0,0,0.05); margin-bottom: 30px;
}
.header-content { display: flex; align-items: center; gap: 20px; }
.team-logo-large { width: 80px; height: 80px; border-radius: 50%; object-fit: cover; border: 2px solid #eee; }
.team-title h1 { margin: 0; display: inline-block; margin-right: 10px; }
.team-tag-large { font-family: monospace; font-size: 1.5rem; color: #666; font-weight: bold; }
.meta { margin: 5px 0 0; color: #888; font-size: 0.9rem; }

/* Grid Layout */
.grid-layout { display: grid; grid-template-columns: 1fr 1fr; gap: 30px; }
@media(max-width: 768px) { .grid-layout { grid-template-columns: 1fr; } }

/* Panels */
.panel { background: white; border: 1px solid #e0e0e0; border-radius: 8px; overflow: hidden; margin-bottom: 20px; }
.panel-header { background: #f8f9fa; padding: 15px 20px; border-bottom: 1px solid #eee; }
.panel-header h3 { margin: 0; font-size: 1.1rem; }
.panel-body { padding: 20px; }

/* Lists */
.list-container { padding: 0; }
.list-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 15px 20px; border-bottom: 1px solid #f0f0f0;
}
.list-item:last-child { border-bottom: none; }
.empty-list { padding: 20px; color: #999; text-align: center; }

/* Member Items */
.member-info { display: flex; align-items: center; gap: 10px; }
.role-badge { 
  font-size: 0.7rem; text-transform: uppercase; padding: 3px 6px; border-radius: 4px; font-weight: bold; 
}
.role-badge.captain { background: #ffd700; color: #5c4500; }
.role-badge.member { background: #e2e3e5; color: #383d41; }
.join-date { color: #aaa; font-size: 0.8rem; margin-left: 5px; }

/* Invite Form */
.input-group { display: flex; gap: 10px; }
.input-group input { flex: 1; padding: 10px; border: 1px solid #ccc; border-radius: 4px; }
.msg-success { color: #28a745; margin-top: 10px; font-size: 0.9rem; }
.msg-error { color: #dc3545; margin-top: 10px; font-size: 0.9rem; }
.invite-info { display: flex; flex-direction: column; text-align: left; }
.expiry { font-size: 0.8rem; color: #999; }

/* Buttons */
.btn-primary { background: #28a745; color: white; border: none; padding: 0 20px; border-radius: 4px; cursor: pointer; font-weight: bold; }
.btn-primary:disabled { background: #94d3a2; }
.btn-danger-outline { background: white; border: 1px solid #dc3545; color: #dc3545; padding: 8px 16px; border-radius: 5px; cursor: pointer; }
.btn-danger-outline:hover { background: #dc3545; color: white; }
.btn-warning-outline { background: white; border: 1px solid #ffc107; color: #d39e00; padding: 8px 16px; border-radius: 5px; cursor: pointer; }
.btn-link-danger { background: none; border: none; color: #dc3545; cursor: pointer; text-decoration: underline; font-size: 0.9rem; }
.btn-icon { background: none; border: none; cursor: pointer; font-size: 1.1rem; color: #ccc; }
.btn-icon:hover { color: #dc3545; }
.btn-link { color: #007bff; text-decoration: none; font-weight: bold; }

.loading-state, .error-state { text-align: center; margin-top: 50px; color: #666; }
</style>