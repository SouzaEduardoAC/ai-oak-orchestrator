<script setup lang="ts">
import { onMounted } from 'vue';
import { useMcpStore } from '../stores/mcp';

const mcpStore = useMcpStore();

const getStatusClass = (status: string) => {
  const classes = {
    'healthy': 'bg-green-900/50 text-green-300 border-green-700',
    'unhealthy': 'bg-red-900/50 text-red-300 border-red-700',
    'reconnecting': 'bg-yellow-900/50 text-yellow-300 border-yellow-700',
    'disconnected': 'bg-gray-700 text-gray-400 border-gray-600'
  };
  return (classes as any)[status] || 'bg-gray-700 text-gray-400';
};

const getStatusIcon = (status: string) => {
  const icons = {
    'healthy': '✓',
    'unhealthy': '⚠',
    'reconnecting': '↻',
    'disconnected': '✗'
  };
  return (icons as any)[status] || '?';
};

const formatTime = (timestamp: number) => {
  if (!timestamp) return 'Never';
  return new Date(timestamp).toLocaleString();
};

onMounted(() => {
  mcpStore.fetchHealth();
});
</script>

<template>
  <div class="space-y-6">
    <!-- Health Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div class="bg-brand-secondary/10 rounded-lg p-4 border border-brand-secondary/30 shadow-lg backdrop-blur-sm">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-gray-400 text-sm">Total MCPs</p>
            <p class="text-3xl font-bold text-brand-primary">{{ mcpStore.summary.total }}</p>
          </div>
          <div class="text-brand-primary">
            <svg class="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
            </svg>
          </div>
        </div>
      </div>

      <div class="bg-brand-secondary/10 rounded-lg p-4 border border-brand-status-green/30 shadow-lg backdrop-blur-sm">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-gray-400 text-sm">Healthy</p>
            <p class="text-3xl font-bold text-brand-status-green">{{ mcpStore.summary.healthy }}</p>
          </div>
          <div class="text-brand-status-green">
            <svg class="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
          </div>
        </div>
      </div>

      <div class="bg-brand-secondary/10 rounded-lg p-4 border border-brand-status-red/30 shadow-lg backdrop-blur-sm">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-gray-400 text-sm">Issues</p>
            <p class="text-3xl font-bold text-brand-status-red">{{ mcpStore.summary.unhealthy + mcpStore.summary.reconnecting }}</p>
          </div>
          <div class="text-brand-status-red">
            <svg class="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
          </div>
        </div>
      </div>
    </div>

    <!-- MCP List -->
    <div class="bg-brand-secondary/10 rounded-lg border border-brand-secondary/30 shadow-xl backdrop-blur-sm">
      <div class="p-4 border-b border-brand-secondary/30 flex justify-between items-center">
        <h2 class="text-lg font-semibold">Registered MCP Servers</h2>
        <button @click="mcpStore.fetchHealth()" class="text-sm bg-brand-primary/20 hover:bg-brand-primary text-brand-primary hover:text-brand-dark px-3 py-1 rounded transition flex items-center space-x-1 font-bold border border-brand-primary/30">
          <span v-if="mcpStore.isLoading" class="animate-spin mr-1">🔄</span>
          <span v-else>🔄</span>
          <span>Refresh</span>
        </button>
      </div>

      <div v-if="mcpStore.isLoading && !mcpStore.registry.length" class="p-12 text-center text-gray-400">
        <div class="animate-spin h-8 w-8 border-4 border-brand-primary border-t-transparent rounded-full mx-auto mb-4"></div>
        <p>Loading server registry...</p>
      </div>

      <div v-else-if="!mcpStore.registry.length" class="p-12 text-center text-gray-500">
        <p class="text-lg">No MCPs configured</p>
        <p class="text-sm mt-2">Connect your first MCP server to begin.</p>
      </div>

      <div v-else class="divide-y divide-brand-secondary/30">
        <div v-for="mcp in mcpStore.registry" :key="mcp.name" class="p-4 hover:bg-brand-secondary/20 transition group">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center space-x-3 mb-2">
                <h3 class="text-lg font-medium text-gray-100">{{ mcp.name }}</h3>
                <span class="px-2 py-0.5 rounded text-xs font-semibold border uppercase tracking-wider" :class="getStatusClass(mcp.status)">
                  {{ getStatusIcon(mcp.status) }} {{ mcp.status }}
                </span>
              </div>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-1 text-sm text-gray-400">
                <div class="flex justify-between sm:justify-start sm:space-x-2">
                  <span class="text-gray-500">Last Check:</span>
                  <span>{{ formatTime(mcp.lastCheck) }}</span>
                </div>
                <div class="flex justify-between sm:justify-start sm:space-x-2">
                  <span class="text-gray-500">Last Success:</span>
                  <span>{{ formatTime(mcp.lastSuccess) }}</span>
                </div>
                <div v-if="mcp.consecutiveFailures > 0" class="col-span-full text-brand-accent flex items-center space-x-1 mt-1">
                  <span class="text-lg leading-none">⚠</span>
                  <span>Consecutive Failures: {{ mcp.consecutiveFailures }}</span>
                </div>
                <div v-if="mcp.error" class="col-span-full text-brand-status-red text-xs mt-1 p-2 bg-red-900/20 rounded font-mono break-all">
                  Error: {{ mcp.error }}
                </div>
              </div>
            </div>
            <div class="ml-4 opacity-0 group-hover:opacity-100 transition">
              <button @click="$emit('remove', mcp.name)" class="p-2 text-gray-400 hover:text-brand-status-red transition" title="Remove MCP">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
