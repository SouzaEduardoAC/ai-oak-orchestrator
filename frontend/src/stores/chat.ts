import { defineStore } from 'pinia';

export type MessageRole = 'user' | 'agent' | 'thinking' | 'approval';

export interface BaseMessage {
  role: MessageRole;
  timestamp: number;
}

export interface UserMessage extends BaseMessage {
  role: 'user';
  content: string;
}

export interface AgentMessage extends BaseMessage {
  role: 'agent';
  content: string;
}

export interface ThinkingMessage extends BaseMessage {
  role: 'thinking';
}

export interface ApprovalRequest extends BaseMessage {
  role: 'approval';
  data: {
    callId: string;
    name: string;
    args: Record<string, any>;
  };
  status: 'pending' | 'approved' | 'rejected';
  output?: string;
  queuePosition: number;
  totalInQueue: number;
}

export type Message = UserMessage | AgentMessage | ThinkingMessage | ApprovalRequest;

export const useChatStore = defineStore('chat', {
  state: () => ({
    messages: [] as Message[],
    isThinking: false,
    availableModels: [] as { id: string; name: string }[],
    selectedModel: '',
  }),
  actions: {
    addMessage(message: Message) {
      this.messages.push(message);
    },
    appendAgentToken(token: string) {
      const last = this.messages[this.messages.length - 1];
      if (last && last.role === 'agent') {
        last.content += token;
      } else {
        this.messages.push({ role: 'agent', content: token, timestamp: Date.now() });
      }
    },
    removeThinking() {
      this.messages = this.messages.filter(m => m.role !== 'thinking');
      this.isThinking = false;
    },
    setThinking() {
      if (!this.messages.some(m => m.role === 'thinking')) {
        this.messages.push({ role: 'thinking', timestamp: Date.now() });
      }
      this.isThinking = true;
    },
    updateApproval(callId: string, status: 'approved' | 'rejected', output?: string) {
      const msg = this.messages.find(m => m.role === 'approval' && m.data.callId === callId) as ApprovalRequest;
      if (msg) {
        msg.status = status;
        if (output) msg.output = output;
      }
    }
  },
});
