<script setup lang="ts">
import { useChatStore } from '../stores/chat';
import { useWebSocket } from '../composables/useWebSocket';
import ChatWindow from '../components/ChatWindow.vue';

const chatStore = useChatStore();
const { send } = useWebSocket();

const handleSendMessage = (content: string) => {
  chatStore.addMessage({
    role: 'user',
    content,
    timestamp: Date.now()
  });
  send('message', content);
};

const handleApproveTool = (callId: string) => {
  chatStore.updateApproval(callId, 'approved');
  send('approval', { callId, approved: true });
};

const handleRejectTool = (callId: string) => {
  chatStore.updateApproval(callId, 'rejected');
  send('approval', { callId, approved: false });
};
</script>

<template>
  <ChatWindow 
    @send="handleSendMessage" 
    @approve="handleApproveTool" 
    @reject="handleRejectTool" 
  />
</template>
