<template>
  <div>

    
    <div class="homepage-container">
      <header class="hero-section">
        <h1>🏆 Welcome to Ersena Tournament Hub 🏆</h1>
        <p>Your centralized platform for creating, joining, and managing all your local sports tournaments.</p>
      </header>
      
      <section class="action-cards">
        <div class="card create-card">
          <h2>Create a New Tournament</h2>
          <p>Organize your league, cup, or bracket event in minutes. Set the rules, schedule the matches, and send out invitations!</p>
          <button @click="goToSignup" class="btn primary-btn">Get Started</button>
        </div>
        
        <div class="card join-card">
          <h2>Find a Tournament to Join</h2>
          <p>Browse local events, view brackets, and register your team. Time to compete!</p>
          <button @click="handleBrowse" class="btn secondary-btn">Browse Tournaments</button>
        </div>
      </section>
      
      <section class="features">
        <h3>Why Choose Ersena?</h3>
        <ul>
          <li>✅ **Easy Setup:** Simple, step-by-step tournament creation.</li>
          <li>📅 **Integrated Scheduling:** Never miss a match time.</li>
          <li>📊 **Real-time Leaderboards:** Track scores and standings instantly.</li>
        </ul>
      </section>
    </div>
  </div>
</template>

<script>
// Assuming NavBar is correctly located in components
import NavBar from '@/components/NavBar.vue' 

export default {
  name: 'HomePage',
  components: {
    NavBar
  },
  methods: {
    goToSignup() {
      // Directs users to the sign-up page as a first step
      this.$router.push('/signup')
    },
    handleBrowse() {
      // Placeholder for browsing functionality
      alert('Browsing functionality coming soon! Please sign up or log in.')
    }
  }
}
</script>

<style scoped>
.homepage-container {
  padding: 0;
  text-align: center;
}

/* --- Hero Section Styling --- */
.hero-section {
  background-color: #007bff; /* Primary blue color */
  color: white;
  padding: 60px 20px;
  margin-bottom: 40px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.hero-section h1 {
  font-size: 2.5em;
  margin-bottom: 10px;
}

.hero-section p {
  font-size: 1.2em;
  opacity: 0.9;
}

/* --- Action Cards Styling --- */
.action-cards {
  display: flex;
  justify-content: center;
  gap: 30px;
  margin: 0 20px 50px;
}

.card {
  flex: 1;
  max-width: 450px;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  background-color: white;
  transition: transform 0.3s;
}

.card:hover {
  transform: translateY(-5px);
}

.card h2 {
  color: #333;
  margin-bottom: 15px;
}

/* --- Buttons --- */
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
  background-color: #28a745; /* Green for creation */
  color: white;
}

.primary-btn:hover {
  background-color: #1e7e34;
}

.secondary-btn {
  background-color: #ffc107; /* Yellow for browsing */
  color: #333;
}

.secondary-btn:hover {
  background-color: #e0a800;
}

/* --- Features Section --- */
.features {
    margin: 40px 20px;
    padding: 30px;
    background-color: #f8f9fa;
    border-radius: 8px;
}

.features ul {
    list-style: none;
    padding: 0;
    display: flex;
    justify-content: center;
    gap: 30px;
    margin-top: 20px;
}

.features li {
    font-size: 1.1em;
    color: #555;
    text-align: left;
}
</style>