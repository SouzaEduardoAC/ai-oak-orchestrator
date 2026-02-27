import axios from 'axios';
import config from './config';
import { useAuthStore } from '../stores/auth';

const api = axios.create({
  baseURL: config.apiBaseUrl,
});

api.interceptors.request.use((axiosConfig) => {
  const authStore = useAuthStore();
  if (config.enableAuth && authStore.token) {
    axiosConfig.headers.Authorization = `Bearer ${authStore.token}`;
  }
  return axiosConfig;
});

export default api;
