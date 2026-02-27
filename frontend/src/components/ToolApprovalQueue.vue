<script setup lang="ts">
import type { ApprovalRequest } from '../stores/chat';

defineProps<{
  request: ApprovalRequest
}>();

const emit = defineEmits(['approve', 'reject']);
</script>

<template>
  <div class="self-start w-full max-w-md bg-gray-800 border border-brand-accent/30 rounded-xl overflow-hidden shadow-xl my-4">
    <div class="bg-brand-accent/10 px-4 py-2 border-b border-brand-accent/20 flex justify-between items-center">
      <div class="flex items-center space-x-2">
        <span class="text-brand-accent font-bold text-xs uppercase tracking-widest">Tool Request</span>
        <span v-if="request.totalInQueue > 1" class="bg-brand-accent text-brand-dark px-1.5 py-0.5 rounded text-[10px] font-black">
          {{ request.queuePosition }} OF {{ request.totalInQueue }}
        </span>
      </div>
      <span class="text-[10px] font-mono text-gray-500 uppercase tracking-tighter">{{ request.data.callId }}</span>
    </div>
    
    <div class="p-4 bg-brand-secondary/20">
      <div class="flex items-center space-x-2 mb-3">
        <div class="w-2 h-2 bg-brand-accent rounded-full animate-pulse"></div>
        <div class="text-lg font-mono font-bold text-white">{{ request.data.name }}</div>
      </div>
      
      <div class="relative group">
        <pre class="bg-brand-dark/80 p-3 rounded-lg text-xs text-gray-300 overflow-x-auto border border-brand-secondary/50 font-mono leading-relaxed max-h-48 overflow-y-auto">{{ JSON.stringify(request.data.args, null, 2) }}</pre>
        <div class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition text-[10px] text-gray-600 bg-brand-dark px-1 rounded">ARGUMENTS</div>
      </div>

      <div v-if="request.status === 'pending'" class="flex space-x-3 mt-4">
        <button @click="emit('approve', request.data.callId)" 
          class="flex-1 bg-brand-status-green/20 hover:bg-brand-status-green text-brand-status-green hover:text-brand-dark border border-brand-status-green/30 py-2 rounded-lg transition-all duration-200 text-sm font-bold shadow-lg shadow-brand-status-green/5">
          APPROVE
        </button>
        <button @click="emit('reject', request.data.callId)" 
          class="flex-1 bg-brand-status-red/10 hover:bg-brand-status-red text-brand-status-red hover:text-brand-dark border border-brand-status-red/30 py-2 rounded-lg transition-all duration-200 text-sm font-bold">
          REJECT
        </button>
      </div>
      
      <div v-else class="mt-4 flex flex-col items-center justify-center py-2 px-4 rounded-lg border italic text-sm"
        :class="request.status === 'approved' ? 'bg-brand-status-green/5 border-brand-status-green/20 text-brand-status-green' : 'bg-brand-status-red/5 border-brand-status-red/20 text-brand-status-red'">
        <div class="flex items-center space-x-2 font-bold not-italic uppercase tracking-widest text-xs">
          <span>{{ request.status === 'approved' ? '✓ Executing' : '✗ Rejected' }}</span>
        </div>
        <div v-if="request.status === 'approved' && request.totalInQueue > request.queuePosition" class="text-[10px] mt-1 text-brand-primary not-italic font-medium uppercase tracking-tighter">
          Waiting for remaining approvals...
        </div>
      </div>

      <div v-if="request.output" class="mt-4 pt-4 border-t border-brand-secondary/50">
        <div class="text-[10px] font-black text-gray-500 mb-2 uppercase tracking-widest flex items-center space-x-2">
          <span>Output Result</span>
          <div class="flex-1 h-px bg-brand-secondary/30"></div>
        </div>
        <pre class="bg-black/40 p-3 rounded-lg text-xs text-brand-status-green overflow-x-auto font-mono max-h-60 overflow-y-auto leading-normal selection:bg-brand-status-green selection:text-brand-dark">{{ request.output }}</pre>
      </div>
    </div>
  </div>
</template>
