<template>
  <div class="create-container">
    <h2>Create a New Team</h2>
    <p class="subtitle">Assemble your squad and start competing.</p>

    <form @submit.prevent="createTeam" class="team-form">
      <div class="form-group">
        <label>Team Name</label>
        <input 
          v-model="form.name" 
          type="text" 
          placeholder="e.g. The Code Warriors" 
          required 
          minlength="3"
        />
      </div>

      <div class="form-group">
        <label>Team Tag (Ticker)</label>
        <input 
          v-model="form.tag" 
          type="text" 
          placeholder="e.g. TCW" 
          required 
          maxlength="5"
          class="uppercase-input"
        />
        <small>Max 5 characters. This will appear next to your name in brackets.</small>
      </div>

      <div class="form-group">
        <label>Logo URL (Optional)</label>
        <input 
          v-model="form.logo_url" 
          type="url" 
          placeholder="https://..." 
        />
      </div>

      <div class="form-actions">
        <button type="button" @click="$router.back()" class="btn-cancel">Cancel</button>
        <button type="submit" class="btn-submit" :disabled="loading">
          {{ loading ? 'Creating...' : 'Create Team' }}
        </button>
      </div>

      <p v-if="error" class="error-msg">{{ error }}</p>
    </form>
  </div>
</template>

<script>
import securedApi from '@/axios-auth.js';

export default {
  name: 'CreateTeamPage',
  data() {
    return {
      form: {
        name: '',
        tag: '',
        logo_url: ''
      },
      loading: false,
      error: null
    };
  },
  methods: {
    async createTeam() {
      this.loading = true;
      this.error = null;

      try {
        // According to your handlers_teams.go -> CreateTeam expects: { name, tag, logo_url }
        await securedApi.post('/api/teams', {
          name: this.form.name,
          tag: this.form.tag.toUpperCase(),
          logo_url: this.form.logo_url || null
        });

        // Redirect back to the list upon success
        this.$router.push({ name: 'Teams' });
      } catch (err) {
        console.error("Creation failed", err);
        this.error = err.response?.data?.error || "Failed to create team. Tag might be taken.";
      } finally {
        this.loading = false;
      }
    }
  }
};
</script>

<style scoped>
.create-container { max-width: 500px; margin: 40px auto; padding: 30px; background: white; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.1); text-align: left; }
h2 { margin-top: 0; text-align: center; color: #333; }
.subtitle { text-align: center; color: #666; margin-bottom: 30px; }

.form-group { margin-bottom: 20px; }
label { display: block; font-weight: 600; margin-bottom: 5px; color: #444; }
input { width: 100%; padding: 10px; border: 1px solid #ccc; border-radius: 4px; font-size: 1rem; box-sizing: border-box; }
input:focus { border-color: #6f42c1; outline: none; }
.uppercase-input { text-transform: uppercase; }
small { color: #888; font-size: 0.85rem; }

.form-actions { display: flex; justify-content: space-between; margin-top: 30px; }
.btn-submit { background: #28a745; color: white; border: none; padding: 10px 25px; border-radius: 5px; font-weight: bold; cursor: pointer; }
.btn-submit:disabled { background: #94d3a2; cursor: not-allowed; }
.btn-cancel { background: transparent; border: 1px solid #ccc; padding: 10px 20px; border-radius: 5px; cursor: pointer; }

.error-msg { color: #dc3545; text-align: center; margin-top: 15px; }
</style>