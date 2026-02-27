import { ref } from 'vue';
import config from '../utils/config';
import { useAuthStore } from '../stores/auth';

export interface WsMessage {
  type: 'message' | 'approval' | 'ping' | 'agent:thinking' | 'agent:response' | 'tool:approval_required' | 'tool:output' | 'agent:error';
  payload: any;
  metadata?: {
    timestamp: number;
    version: string;
  };
}

// Shared state
const socket = ref<WebSocket | null>(null);
const isConnected = ref(false);
const reconnectAttempts = ref(0);
const maxReconnectDelay = 30000;
const onMessageCallbacks = new Set<(msg: WsMessage) => void>();

export function useWebSocket() {
  const connect = () => {
    if (socket.value && (socket.value.readyState === WebSocket.OPEN || socket.value.readyState === WebSocket.CONNECTING)) {
      return;
    }

    const authStore = useAuthStore();
    let url = config.wsUrl;
    
    if (!url) {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      url = `${protocol}//${window.location.host}/ws`;
    }
    
    if (config.enableAuth && authStore.token) {
      url += (url.includes('?') ? '&' : '?') + `token=${authStore.token}`;
    }

    socket.value = new WebSocket(url);

    socket.value.onopen = () => {
      console.log('WebSocket connected');
      isConnected.value = true;
      reconnectAttempts.value = 0;
    };

    socket.value.onmessage = (event) => {
      try {
        const msg: WsMessage = JSON.parse(event.data);
        onMessageCallbacks.forEach(cb => cb(msg));
      } catch (e) {
        console.error('Failed to parse WebSocket message', e);
      }
    };

    socket.value.onclose = (event) => {
      isConnected.value = false;
      if (event.code !== 1000) { // Not normal closure
        console.log('WebSocket closed unexpectedly, attempting reconnect...');
        scheduleReconnect();
      }
    };

    socket.value.onerror = (error) => {
      console.error('WebSocket error', error);
      socket.value?.close();
    };
  };

  const scheduleReconnect = () => {
    const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.value), maxReconnectDelay);
    reconnectAttempts.value++;
    setTimeout(connect, delay);
  };

  const send = (type: string, payload: any) => {
    if (socket.value && isConnected.value) {
      const msg: WsMessage = {
        type: type as any,
        payload,
        metadata: {
          timestamp: Date.now(),
          version: '1.0',
        },
      };
      socket.value.send(JSON.stringify(msg));
    } else {
      console.warn('Cannot send message: WebSocket is not connected');
    }
  };

  const onMessage = (callback: (msg: WsMessage) => void) => {
    onMessageCallbacks.add(callback);
    // Return unsubscribe function
    return () => onMessageCallbacks.delete(callback);
  };

  const disconnect = () => {
    if (socket.value) {
      const s = socket.value;
      socket.value = null;
      s.close(1000);
    }
  };

  return {
    isConnected,
    connect,
    disconnect,
    send,
    onMessage,
  };
}