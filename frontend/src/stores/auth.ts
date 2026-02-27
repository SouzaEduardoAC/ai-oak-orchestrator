import { defineStore } from 'pinia';
import Keycloak from 'keycloak-js';
import config from '../utils/config';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    isAuthenticated: false,
    user: null as any,
    token: null as string | null,
    keycloak: null as Keycloak | null,
  }),
  actions: {
    async init() {
      if (!config.enableAuth) {
        this.isAuthenticated = true;
        this.user = { name: 'Local Admin', preferred_username: 'admin' };
        return true;
      }

      this.keycloak = new Keycloak({
        url: config.keycloak.url,
        realm: config.keycloak.realm,
        clientId: config.keycloak.clientId,
      });

      try {
        const authenticated = await this.keycloak.init({
          onLoad: 'login-required',
          checkLoginIframe: false,
        });

        this.isAuthenticated = authenticated;
        if (authenticated) {
          this.token = this.keycloak.token || null;
          this.user = this.keycloak.tokenParsed;
        }
        return authenticated;
      } catch (error) {
        console.error('Failed to initialize Keycloak', error);
        return false;
      }
    },
    logout() {
      if (this.keycloak) {
        this.keycloak.logout();
      }
      this.isAuthenticated = false;
      this.user = null;
      this.token = null;
    }
  },
});
