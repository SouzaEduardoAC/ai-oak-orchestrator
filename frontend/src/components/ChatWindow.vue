<script setup lang="ts">
import { ref, nextTick, watch } from 'vue';
import { useChatStore } from '../stores/chat';
import ToolApprovalQueue from './ToolApprovalQueue.vue';

const chatStore = useChatStore();
const chatContainer = ref<HTMLElement | null>(null);
const inputMessage = ref('');

const emit = defineEmits(['send', 'approve', 'reject']);

const scrollToBottom = () => {
  nextTick(() => {
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
    }
  });
};

watch(() => chatStore.messages.length, scrollToBottom);
watch(() => chatStore.isThinking, scrollToBottom);

const handleSend = () => {
  if (!inputMessage.value.trim() || chatStore.isThinking) return;
  emit('send', inputMessage.value);
  inputMessage.value = '';
};

const handleApprove = (callId: string) => {
  emit('approve', callId);
};

const handleReject = (callId: string) => {
  emit('reject', callId);
};
</script>

<template>
  <div class="flex-1 flex flex-col h-full overflow-hidden bg-brand-dark/50 backdrop-blur-sm rounded-2xl border border-gray-800 shadow-2xl">
    <!-- Messages Area -->
    <div ref="chatContainer" class="flex-1 overflow-y-auto p-6 space-y-6 scrollbar-thin scrollbar-thumb-gray-700 scrollbar-track-transparent">
      <div v-if="chatStore.messages.length === 0" class="h-full flex flex-col items-center justify-center text-gray-500 space-y-4">
        <div class="w-16 h-16 rounded-full bg-gray-800/50 flex items-center justify-center border border-gray-700/50">
          <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"></path>
          </svg>
        </div>
        <div class="text-center">
          <p class="text-lg font-medium text-gray-400">Initialize sequence...</p>
          <p class="text-xs uppercase tracking-widest mt-1 opacity-50">Awaiting user directives</p>
        </div>
      </div>

      <div v-for="(msg, index) in chatStore.messages" :key="index" class="flex flex-col animate-in fade-in slide-in-from-bottom-2 duration-300">
        <!-- User Message -->
        <div v-if="msg.role === 'user'" class="self-end max-w-[85%] group">
          <div class="flex flex-col items-end space-y-1">
            <div class="bg-brand-primary text-white px-5 py-3 rounded-2xl rounded-tr-none font-medium shadow-lg shadow-brand-primary/10">
              {{ msg.content }}
            </div>
            <span class="text-[10px] text-gray-600 font-bold uppercase tracking-widest opacity-0 group-hover:opacity-100 transition pr-1">Directive Sent</span>
          </div>
        </div>

        <!-- Agent Message -->
        <div v-if="msg.role === 'agent'" class="self-start max-w-[85%] group">
          <div class="flex flex-col items-start space-y-1">
            <div class="bg-brand-secondary/40 text-gray-100 px-5 py-3 rounded-2xl rounded-tl-none border border-brand-secondary/50 shadow-xl backdrop-blur-md">
              <p class="whitespace-pre-wrap leading-relaxed text-sm selection:bg-brand-primary selection:text-brand-dark">{{ msg.content }}</p>
            </div>
            <span class="text-[10px] text-gray-600 font-bold uppercase tracking-widest opacity-0 group-hover:opacity-100 transition pl-1">Orchestrator Feedback</span>
          </div>
        </div>

        <!-- Thinking State -->
        <div v-if="msg.role === 'thinking'" class="self-start pl-2 flex items-center space-x-3">
          <div class="flex space-x-1.5">
            <div class="w-1.5 h-1.5 bg-brand-primary rounded-full animate-bounce [animation-delay:-0.3s]"></div>
            <div class="w-1.5 h-1.5 bg-brand-primary rounded-full animate-bounce [animation-delay:-0.15s]"></div>
            <div class="w-1.5 h-1.5 bg-brand-primary rounded-full animate-bounce"></div>
          </div>
          <span class="text-[10px] font-black text-brand-primary uppercase tracking-[0.2em] animate-pulse">Processing</span>
        </div>

        <!-- Tool Approval -->
        <ToolApprovalQueue 
          v-if="msg.role === 'approval'" 
          :request="msg" 
          @approve="handleApprove" 
          @reject="handleReject" 
        />
      </div>
    </div>

    <!-- Input Area -->
    <div class="p-6 bg-gray-900/50 border-t border-gray-800">
      <form @submit.prevent="handleSend" class="relative group">
        <input
          v-model="inputMessage"
          type="text"
          placeholder="ENTER MISSION DIRECTIVE..."
          class="w-full bg-gray-950 border border-gray-700 text-gray-100 rounded-xl px-5 py-4 pr-24 focus:outline-none focus:ring-2 focus:ring-brand-primary/50 focus:border-brand-primary transition-all duration-300 placeholder:text-gray-700 font-mono text-sm tracking-wide shadow-inner"
          :disabled="chatStore.isThinking"
        >
        <div class="absolute right-2 top-2 bottom-2 flex items-center">
          <button
            type="submit"
            class="h-full bg-brand-primary/10 border border-brand-primary/30 hover:bg-brand-primary hover:text-brand-dark disabled:bg-gray-800 disabled:text-gray-600 text-brand-primary px-6 rounded-lg font-black text-xs uppercase tracking-widest transition-all duration-200 active:scale-95 shadow-lg shadow-brand-primary/10"
            :disabled="!inputMessage.trim() || chatStore.isThinking"
          >
            EXECUTE
          </button>
        </div>
      </form>
      <div class="mt-3 flex justify-between items-center px-1">
        <div class="flex items-center space-x-4">
          <div class="flex items-center space-x-1.5">
            <div class="w-1 h-1 rounded-full" :class="chatStore.isThinking ? 'bg-brand-accent animate-pulse' : 'bg-gray-700'"></div>
            <span class="text-[9px] font-bold text-gray-600 uppercase tracking-tighter">System Engine</span>
          </div>
          <div class="flex items-center space-x-1.5">
            <div class="w-1 h-1 rounded-full" :class="!chatStore.isThinking ? 'bg-brand-status-green' : 'bg-gray-700'"></div>
            <span class="text-[9px] font-bold text-gray-600 uppercase tracking-tighter">Input Buffer</span>
          </div>
        </div>
        <span class="text-[9px] font-mono text-gray-700 uppercase tracking-widest">v1.0.0-PROTOTYPE</span>
      </div>
    </div>
  </div>
</template>
