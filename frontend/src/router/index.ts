import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/chat',
    },
    {
      path: '/chat',
      name: 'chat',
      component: () => import('../views/ChatView.vue'),
    },
    {
      path: '/mcp',
      name: 'mcp',
      component: () => import('../views/McpView.vue'),
    },
  ],
});

router.beforeEach(async (_to, _from, next) => {
  const authStore = useAuthStore();
  if (!authStore.isAuthenticated) {
    await authStore.init();
  }
  next();
});

export default router;
