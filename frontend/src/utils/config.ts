export const config = {
  enableAuth: import.meta.env.VITE_ENABLE_AUTH === 'true',
  keycloak: {
    url: import.meta.env.VITE_KEYCLOAK_URL || '',
    realm: import.meta.env.VITE_KEYCLOAK_REALM || '',
    clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID || '',
  },
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || '/api',
  wsUrl: import.meta.env.VITE_WS_URL || '',
};

export default config;
