<template>
  <div class="profile-page">
    <div class="profile-card">
      <!-- Header -->
      <div class="profile-header">
        <div class="avatar">
          <span>{{ initials }}</span>
        </div>
        <div>
          <h1 class="profile-title">My Profile</h1>
          <p class="profile-subtitle">
            View and update your account information.
          </p>
        </div>
      </div>

      <hr />

      <!-- Profile form -->
      <div class="profile-info">
        <div class="info-row">
          <label>Username</label>
          <div class="info-value">{{ profile.username }}</div>
        </div>

        <div class="info-row">
          <label>Email</label>
          <div class="info-value">{{ profile.email }}</div>
        </div>

        <!-- <div class="info-row">
          <label>Team</label>
          <div class="info-value">{{ profile.team }}</div>
        </div> -->
      </div>

      <div class="button-row">
        <button
          type="button"
          class="btn-danger-ghost"
          @click="confirmDelete"
        >
          Delete Account
        </button>

        <button
          type="button"
          class="btn-primary"
          @click="goToAccountManagement"
        >
          Edit Account in Keycloak
        </button>
      </div>

      <!-- Upcoming tournaments  -->
      <!-- <div class="tournaments-section">
        <div class="tournaments-header">
          <h2>Upcoming tournaments</h2>
          <p>These are the tournaments you’re registered for.</p>
        </div>

        <div v-if="upcomingTournaments.length === 0" class="tournaments-empty">
          <p>You are not registered for any upcoming tournaments yet.</p>
        </div>

        <div v-else class="tournaments-list">
          <div
            v-for="t in upcomingTournaments"
            :key="t.id"
            class="tournament-card"
          >
            <div class="tournament-main">
              <h3 class="tournament-name">{{ t.name }}</h3>
              <p class="tournament-meta">
                <span>{{ t.date }}</span>
                <span>•</span>
                <span>{{ t.location }}</span>
              </p>
              <p class="tournament-meta" v-if="t.team">
                Playing as: <strong>{{ t.team }}</strong>
              </p>
            </div>

            <div class="tournament-side">
              <span
                class="status-pill"
                :class="`status-${t.status.toLowerCase()}`"
              >
                {{ t.status }}
              </span>
              <button
                type="button"
                class="btn-link"
                @click="viewTournament(t)"
              >
                View details
              </button>
            </div>
          </div>
        </div>
      </div> -->
      <!-- end tournaments section -->
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed, onMounted, getCurrentInstance } from "vue";
import securedApi from '@/axios-auth.js';

const { proxy } = getCurrentInstance();
const keycloak = proxy.$keycloak;

const profile = reactive({
  username: "",
  email: "",
});

// Local editable copy so you can cancel changes
const editableProfile = reactive({ ...profile });

const isEditing = ref(false);

onMounted(() => {
  if (keycloak && keycloak.authenticated) {
    if (keycloak.tokenParsed) {
      profile.username = keycloak.tokenParsed.preferred_username || "";
      profile.email = keycloak.tokenParsed.email || "";
      Object.assign(editableProfile, profile); // Update editable profile too
    }
  }
});

// Simple initials for avatar
const initials = computed(() => {
  if (!editableProfile.username) return "?";
  return editableProfile.username
    .split(/[\s._-]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("");
});

/**
 * Redirects to Keycloak's built-in account management page.
 * Users can update password, email, and OTP settings there.
 */
function goToAccountManagement() {
  if (keycloak) {
    keycloak.accountManagement();
  }
}

/**
 * Handles account deletion via the API Gateway.
 */
async function confirmDelete() {
  const confirmed = window.confirm(
    "Are you sure you want to delete your account? This action cannot be undone."
  );

  if (!confirmed) return;

  try {
    await securedApi.delete('/api/users/me');
    
    // On success, logout locally and from Keycloak
    keycloak.logout();
  } catch (error) {
    console.error("Failed to delete account:", error);
    alert("Could not delete account. Please try again.");
  }
}
</script>

<style scoped>
.profile-page {
  min-height: calc(100vh - 80px);
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 2rem 1rem;
  background: #f3f4f6;
}

.profile-card {
  width: 100%;
  max-width: 840px;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 10px 25px rgba(15, 23, 42, 0.08);
  padding: 1.75rem;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  margin-bottom: 1rem;
}

.avatar {
  width: 56px;
  height: 56px;
  border-radius: 999px;
  background: #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 1.1rem;
  color: #374151;
}

.profile-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: #111827;
}

.profile-subtitle {
  margin: 0.15rem 0 0;
  font-size: 0.9rem;
  color: #6b7280;
}

/* Info Section (Read-only view) */
.profile-info {
  margin-top: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.info-row {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-row label {
  font-size: 0.85rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.025em;
  color: #6b7280;
}

.info-value {
  font-size: 1rem;
  color: #111827;
  padding: 0.5rem 0;
  border-bottom: 1px solid #f3f4f6;
}

/* Buttons */
.button-row {
  margin-top: 2rem;
  display: flex;
  align-items: center;
  justify-content: flex-end; /* Align buttons to right */
  gap: 1rem;
}

.btn-primary {
  background-color: #2563eb;
  color: #ffffff;
  border: 1px solid #2563eb;
  border-radius: 999px;
  padding: 0.5rem 1.25rem;
  font-weight: 500;
  cursor: pointer;
}

.btn-primary:hover {
  background-color: #1d4ed8;
}

.btn-danger-ghost {
  background-color: transparent;
  color: #dc2626;
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0.5rem 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-danger-ghost:hover {
  background-color: #fef2f2;
  border-color: #fecaca;
}

/* Tournaments CSS (Kept from your file) */
.tournaments-section {
  margin-top: 2.5rem;
  border-top: 1px solid #e5e7eb;
  padding-top: 1.5rem;
}
.tournaments-header h2 { margin: 0; font-size: 1.15rem; font-weight: 600; color: #111827; }
.tournaments-header p { margin: 0.25rem 0 0; font-size: 0.9rem; color: #6b7280; }
.tournaments-list { margin-top: 1rem; display: flex; flex-direction: column; gap: 0.75rem; }
.tournament-card { display: flex; justify-content: space-between; padding: 0.9rem 1rem; border-radius: 10px; border: 1px solid #e5e7eb; background-color: #f9fafb; align-items: center; }
.tournament-name { font-weight: 600; color: #111827; margin: 0; }
.status-pill { padding: 0.15rem 0.6rem; border-radius: 999px; font-size: 0.75rem; font-weight: 500; }
.status-registered { background-color: #eff6ff; color: #1d4ed8; }
</style>