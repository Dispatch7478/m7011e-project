import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '../views/HomePage.vue'
import SignupPage from '../views/SignupPage.vue'
import LoginPage from '../views/LoginPage.vue'
import ProfilePage from "../views/ProfilePage.vue"
import TournamentPage from '../views/TournamentPage.vue'
const routes = [
  { path: '/',      name: 'Home',   component: HomePage },
  { path: '/signup', name: 'Signup', component: SignupPage },
  { path: '/login',  name: 'Login',  component: LoginPage },
  { path: '/profile', name: 'Profile', component: ProfilePage },
  { path: '/tournaments', name: 'Tournaments', component: TournamentPage }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router