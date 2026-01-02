import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '../views/HomePage.vue'
import SignupPage from '../views/SignupPage.vue'
import LoginPage from '../views/LoginPage.vue'
import ProfilePage from "../views/ProfilePage.vue"
import TournamentPage from '../views/TournamentPage.vue'
import CreateTournamentPage from '../views/CreateTournamentPage.vue'
import BracketPage from '../views/BracketPage.vue'
import ChangeTournamentPage from '../views/ChangeTournamentPage.vue'
import AdminPage from '../views/AdminPage.vue'
import TeamListPage from '../views/TeamListPage.vue';
import CreateTeamPage from '../views/CreateTeamPage.vue';

const routes = [
  { path: '/',      name: 'Home',   component: HomePage },
  { path: '/signup', name: 'Signup', component: SignupPage },
  { path: '/login',  name: 'Login',  component: LoginPage },
  { path: '/profile', name: 'Profile', component: ProfilePage },
  { path: '/tournaments', name: 'Tournaments', component: TournamentPage },
  { path: '/tournaments/create', name: 'CreateTournament', component: CreateTournamentPage },
  { path: '/tournaments/:id/bracket', name: 'Bracket', component: BracketPage },
  { path: '/tournaments/:id/edit', name: 'ChangeTournament', component: ChangeTournamentPage },
  { 
    path: '/admin',
    name: 'Admin',
    component: AdminPage,
    beforeEnter: (to, from, next) => {
      if (window.keycloakInstance && window.keycloakInstance.authenticated && window.keycloakInstance.hasRealmRole('SuperAdmin')) {
        next();
      } else {
        next('/');
      }
    }
  },
  {
    path: '/teams',
    name: 'Teams',
    component: TeamListPage,
  },
  {
    path: '/teams/create',
    name: 'CreateTeam',
    component: CreateTeamPage,
    meta: { requiresAuth: true }
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router