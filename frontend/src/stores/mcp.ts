import { defineStore } from 'pinia';
import api from '../utils/api';

export interface McpHealth {
  name: string;
  status: 'healthy' | 'unhealthy' | 'reconnecting' | 'disconnected';
  lastCheck: number;
  lastSuccess: number;
  consecutiveFailures: number;
  error?: string;
}

export interface McpSummary {
  total: number;
  healthy: number;
  unhealthy: number;
  reconnecting: number;
}

export const useMcpStore = defineStore('mcp', {
  state: () => ({
    registry: [] as McpHealth[],
    summary: { total: 0, healthy: 0, unhealthy: 0, reconnecting: 0 } as McpSummary,
    isLoading: false,
  }),
  actions: {
    async fetchHealth() {
      this.isLoading = true;
      try {
        const response = await api.get('/mcp/health');
        this.registry = response.data.mcps || [];
        this.summary = response.data.summary || { total: 0, healthy: 0, unhealthy: 0, reconnecting: 0 };
      } catch (error) {
        console.error('Failed to fetch MCP health', error);
      } finally {
        this.isLoading = false;
      }
    },
    async addMcp(name: string, config: any) {
      try {
        await api.post('/mcp/add', { name, config });
        await this.fetchHealth();
        return true;
      } catch (error) {
        console.error('Failed to add MCP', error);
        throw error;
      }
    },
    async removeMcp(name: string) {
      try {
        await api.delete(`/mcp/${encodeURIComponent(name)}`);
        await this.fetchHealth();
        return true;
      } catch (error) {
        console.error('Failed to remove MCP', error);
        throw error;
      }
    }
  },
});
