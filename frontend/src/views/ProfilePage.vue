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
      <form class="profile-form" @submit.prevent="onSave">
        <div class="form-row">
          <label for="username">Username</label>
          <input
            id="username"
            v-model="editableProfile.username"
            :readonly="!isEditing"
            :class="{ readonly: !isEditing }"
            type="text"
          />
        </div>

        <div class="form-row">
          <label for="email">Email</label>
          <input
            id="email"
            v-model="editableProfile.email"
            :readonly="!isEditing"
            :class="{ readonly: !isEditing }"
            type="email"
          />
        </div>

        <div class="form-row">
          <label for="team">Team</label>
          <input
            id="team"
            v-model="editableProfile.team"
            :readonly="!isEditing"
            :class="{ readonly: !isEditing }"
            type="text"
          />
        </div>

        <!-- Buttons -->
        <div class="button-row">
          <button
            type="button"
            class="btn-secondary"
            @click="onChangePassword"
          >
            Change password
          </button>

          <div class="button-row-right">
            <button
              v-if="!isEditing"
              type="button"
              class="btn-primary"
              @click="startEditing"
            >
              Edit profile
            </button>

            <template v-else>
              <button
                type="button"
                class="btn-ghost"
                @click="cancelEditing"
              >
                Cancel
              </button>
              <button type="submit" class="btn-primary">
                Save changes
              </button>
            </template>
          </div>
        </div>
      </form>

      <!-- Upcoming tournaments  -->
      <div class="tournaments-section">
        <div class="tournaments-header">
          <h2>Upcoming tournaments</h2>
          <p>These are the tournaments you’re registered for.</p>
        </div>

        <div v-if="upcomingTournaments.length === 0" class="tournaments-empty">
          <p>You are not registered for any upcoming tournaments yet.</p>
          <!-- Later you can turn this into a link/button to the tournaments page -->
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
      </div>
      <!-- end tournaments section -->
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed, onMounted, getCurrentInstance } from "vue";

const { proxy } = getCurrentInstance();
const keycloak = proxy.$keycloak;

const profile = reactive({
  username: "",
  email: "",
  team: "Red Team", // This would typically come from a backend user service
});

// Local editable copy so you can cancel changes
const editableProfile = reactive({ ...profile });

const isEditing = ref(false);

// Mock upcoming tournaments (replace with API/Keycloak backed data later)
const upcomingTournaments = ref([
  {
    id: 1,
    name: "Winter Clash 2025",
    date: "2025-12-05 19:00",
    location: "Online",
    team: "Red Team",
    status: "Registered",
  },
  {
    id: 2,
    name: "Holiday Cup",
    date: "2025-12-18 20:30",
    location: "Local Arena",
    team: "Red Team",
    status: "Confirmed",
  },
]);

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

function startEditing() {
  Object.assign(editableProfile, profile);
  isEditing.value = true;
}

function cancelEditing() {
  Object.assign(editableProfile, profile);
  isEditing.value = false;
}

function onSave() {
  // In a real app, you’d call your backend API here.
  // For now, just update the base profile with editable changes
  Object.assign(profile, editableProfile);
  isEditing.value = false;
  console.log("Profile saved (no backend yet):", profile);
}

/**
 * Keycloak password change placeholder.
 * Later you can:
 *   keycloak.accountManagement();
 * or redirect to your Keycloak account console.
 */
function onChangePassword() {
  if (keycloak && keycloak.authenticated) {
    // Redirect to Keycloak's account management console for password changes
    keycloak.accountManagement();
  } else {
    console.log("Not authenticated. Cannot change password.");
  }
}

/**
 * Placeholder for navigating to a specific tournament:
 * - router.push({ name: 'tournament-details', params: { id: t.id }})
 */
function viewTournament(tournament) {
  console.log("View tournament clicked:", tournament);
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

/* Header */
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

/* Profile form */
.profile-form {
  margin-top: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.form-row label {
  font-size: 0.9rem;
  font-weight: 500;
  color: #4b5563;
}

.form-row input {
  padding: 0.55rem 0.7rem;
  border-radius: 8px;
  border: 1px solid #d1d5db;
  font-size: 0.95rem;
  outline: none;
  transition: border-color 0.15s ease, box-shadow 0.15s ease, background-color 0.15s ease;
}

.form-row input:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 1px rgba(37, 99, 235, 0.2);
}

.form-row input.readonly {
  background-color: #f9fafb;
  color: #4b5563;
  cursor: default;
}

/* Buttons */
.button-row {
  margin-top: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.button-row-right {
  display: flex;
  gap: 0.5rem;
}

button {
  border-radius: 999px;
  border: 1px solid transparent;
  padding: 0.45rem 1rem;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.btn-primary {
  background-color: #2563eb;
  color: #ffffff;
  border-color: #2563eb;
}

.btn-primary:hover {
  background-color: #1d4ed8;
  border-color: #1d4ed8;
  box-shadow: 0 6px 18px rgba(37, 99, 235, 0.35);
}

.btn-secondary {
  background-color: #ffffff;
  color: #111827;
  border-color: #d1d5db;
}

.btn-secondary:hover {
  background-color: #f3f4f6;
}

.btn-ghost {
  background-color: transparent;
  color: #4b5563;
  border-color: transparent;
}

.btn-ghost:hover {
  background-color: #f3f4f6;
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

/* Tournaments section */
.tournaments-section {
  margin-top: 2rem;
  border-top: 1px solid #e5e7eb;
  padding-top: 1.5rem;
}

.tournaments-header h2 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 600;
  color: #111827;
}

.tournaments-header p {
  margin: 0.25rem 0 0;
  font-size: 0.9rem;
  color: #6b7280;
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

.tournaments-list {
  margin-top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
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

.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.15rem 0.6rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 500;
}

/* status colors */
.status-registered {
  background-color: #eff6ff;
  color: #1d4ed8;
}

.status-confirmed {
  background-color: #ecfdf3;
  color: #15803d;
}

.status-waiting {
  background-color: #fef9c3;
  color: #92400e;
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