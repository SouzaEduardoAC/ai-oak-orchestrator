<script setup lang="ts">
import { ref } from 'vue';

const emit = defineEmits(['submit', 'cancel']);

const form = ref({
  name: '',
  transport: 'http',
  description: '',
  url: '',
  headers: [] as { key: string; value: string }[],
  command: '',
  args: '',
  cwd: '',
  containerImage: '',
  containerMemory: 512,
  containerCpu: 0.5
});

const transportTypes = [
  { value: 'http', label: 'HTTP', description: 'REST/HTTP-based server' },
  { value: 'sse', label: 'SSE', description: 'Server-Sent Events' },
  { value: 'stdio', label: 'Stdio', description: 'Local process' },
  { value: 'stdio-docker', label: 'Stdio-Docker', description: 'Docker container' }
];

const addHeader = () => {
  form.value.headers.push({ key: '', value: '' });
};

const removeHeader = (index: number) => {
  form.value.headers.splice(index, 1);
};

const handleSubmit = () => {
  // Build payload
  const config: any = {
    transport: form.value.transport,
    enabled: true,
    description: form.value.description
  };

  if (form.value.transport === 'http' || form.value.transport === 'sse') {
    config.url = form.value.url;
    if (form.value.headers.length > 0) {
      const headers: Record<string, string> = {};
      form.value.headers.forEach(h => {
        if (h.key && h.value) headers[h.key] = h.value;
      });
      if (Object.keys(headers).length > 0) config.headers = headers;
    }
  } else if (form.value.transport === 'stdio') {
    config.command = form.value.command;
    if (form.value.args) config.args = form.value.args.split(' ').filter(a => a.length > 0);
    if (form.value.cwd) config.cwd = form.value.cwd;
  } else if (form.value.transport === 'stdio-docker') {
    config.containerImage = form.value.containerImage;
    config.containerMemory = form.value.containerMemory;
    config.containerCpu = form.value.containerCpu;
  }

  emit('submit', { name: form.value.name, config });
};
</script>

<template>
  <form @submit.prevent="handleSubmit" class="space-y-4">
    <!-- Name -->
    <div>
      <label class="block text-sm font-medium text-gray-300 mb-1">Name *</label>
      <input v-model="form.name" type="text" required placeholder="my-mcp-server" class="w-full bg-gray-900 border border-gray-700 rounded-md px-3 py-2 text-white focus:ring-2 focus:ring-brand-primary focus:border-transparent">
    </div>

    <!-- Transport Type -->
    <div>
      <label class="block text-sm font-medium text-gray-300 mb-2">Transport Type *</label>
      <div class="grid grid-cols-2 gap-3">
        <button v-for="type in transportTypes" :key="type.value" type="button" @click="form.transport = type.value"
          :class="form.transport === type.value ? 'border-brand-primary bg-brand-primary/10 ring-1 ring-brand-primary' : 'border-gray-700 bg-gray-900'"
          class="p-3 border rounded-lg text-left transition hover:border-gray-500">
          <div class="font-semibold text-sm">{{ type.label }}</div>
          <div class="text-xs text-gray-500 line-clamp-1">{{ type.description }}</div>
        </button>
      </div>
    </div>

    <!-- Description -->
    <div>
      <label class="block text-sm font-medium text-gray-300 mb-1">Description</label>
      <input v-model="form.description" type="text" placeholder="Short description" class="w-full bg-gray-900 border border-gray-700 rounded-md px-3 py-2 text-white">
    </div>

    <!-- HTTP/SSE Specific -->
    <div v-if="form.transport === 'http' || form.transport === 'sse'" class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-300 mb-1">URL *</label>
        <input v-model="form.url" type="url" required placeholder="http://localhost:8080" class="w-full bg-gray-900 border border-gray-700 rounded-md px-3 py-2 text-white">
      </div>
      <div>
        <div class="flex justify-between items-center mb-2">
          <label class="text-sm font-medium text-gray-300">Custom Headers</label>
          <button type="button" @click="addHeader" class="text-xs text-brand-primary hover:underline">+ Add Header</button>
        </div>
        <div v-for="(header, i) in form.headers" :key="i" class="flex space-x-2 mb-2">
          <input v-model="header.key" placeholder="Key" class="flex-1 bg-gray-900 border border-gray-700 rounded-md px-2 py-1 text-sm text-white">
          <input v-model="header.value" placeholder="Value" class="flex-1 bg-gray-900 border border-gray-700 rounded-md px-2 py-1 text-sm text-white">
          <button type="button" @click="removeHeader(i)" class="text-red-400 p-1">×</button>
        </div>
      </div>
    </div>

    <!-- Stdio Specific -->
    <div v-if="form.transport === 'stdio'" class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-300 mb-1">Command *</label>
        <input v-model="form.command" type="text" required placeholder="python" class="w-full bg-gray-900 border border-gray-700 rounded-md px-3 py-2 text-white">
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-300 mb-1">Arguments</label>
        <input v-model="form.args" type="text" placeholder="-m mcp_server" class="w-full bg-gray-900 border border-gray-700 rounded-md px-3 py-2 text-white">
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-300 mb-1">Working Directory</label>
        <input v-model="form.cwd" type="text" placeholder="/home/user/mcp" class="w-full bg-gray-900 border border-gray-700 rounded-md px-3 py-2 text-white">
      </div>
    </div>

    <!-- Docker Specific -->
    <div v-if="form.transport === 'stdio-docker'" class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-300 mb-1">Container Image *</label>
        <input v-model="form.containerImage" type="text" required placeholder="mcp/sqlite:latest" class="w-full bg-gray-900 border border-gray-700 rounded-md px-3 py-2 text-white">
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Memory Limit (MB)</label>
          <input v-model.number="form.containerMemory" type="number" class="w-full bg-gray-900 border border-gray-700 rounded-md px-3 py-2 text-white">
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">CPU Limit</label>
          <input v-model.number="form.containerCpu" type="number" step="0.1" class="w-full bg-gray-900 border border-gray-700 rounded-md px-3 py-2 text-white">
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex space-x-3 pt-4">
      <button type="button" @click="emit('cancel')" class="flex-1 px-4 py-2 border border-brand-secondary rounded-lg hover:bg-brand-secondary/20 transition text-gray-300">Cancel</button>
      <button type="submit" class="flex-1 px-4 py-2 bg-brand-primary hover:bg-brand-primary/80 text-brand-dark font-bold rounded-lg transition shadow-lg shadow-brand-primary/10">Add Server</button>
    </div>
  </form>
</template>
