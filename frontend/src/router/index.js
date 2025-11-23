import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '../views/HomePage.vue'
import SignupPage from '../views/SignupPage.vue'
import LoginPage from '../views/LoginPage.vue'

const routes = [
  { path: '/',      name: 'Home',   component: HomePage },
  { path: '/signup', name: 'Signup', component: SignupPage },
  { path: '/login',  name: 'Login',  component: LoginPage }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router