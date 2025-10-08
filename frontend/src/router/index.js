// router/index.js
import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/store/auth';
import Home from '@/pages/Home.vue'
import Login from '@/pages/Login.vue'
import Register from '@/pages/Register.vue'
import Chat from '@/pages/Chat.vue'
import Profile from '@/pages/Profile.vue'
import GroupsList from '@/pages/GroupsList.vue'
import Group from '@/pages/Group.vue'

const routes = [
  { path: '/', name: 'Home', component: Home, meta: { requiresAuth: true } },
  { path: '/groups', name: 'Groups', component: GroupsList, meta: { requiresAuth: true } },
  { path: '/group/:id', name: 'Group', component: Group, meta: { requiresAuth: true } },
  { path: '/login', name: 'Login', component: Login },
  { path: '/register', name: 'Register', component: Register },
  { path: '/chat', name: 'Chat', component: Chat, meta: { requiresAuth: true } },
    { path: '/profile', name: 'Profile', component: Profile, meta: { requiresAuth: true } },
]

const router = createRouter({ history: createWebHistory(), routes });

// Navigation Guard
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore();
  // Fetch user on first load to check the session cookie
  if (authStore.user === null) {
      await authStore.fetchUser();
  }

  if (to.meta.requiresAuth && !authStore.user) {
    // If route is protected and user is not logged in
    next('/login');
  } else {
    next();
  }
});

export default router;