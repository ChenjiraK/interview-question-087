<script setup lang="ts">
const [model, modifiers] = defineModel<string, 'trim'>({
  default: '',
  set(value) {
    return modifiers.trim ? value.trim() : value
  },
})

const emit = defineEmits<{ onChange: [value: string] }>()
</script>

<template>
  <input
    v-model="model"
    type="text"
    class="h-11 w-full min-w-0 rounded-md border border-gray-400 bg-white px-3 outline-none focus:border-sky-600 focus:ring-2 focus:ring-sky-200 disabled:bg-gray-50"
    @change="emit('onChange', model)"
  />
</template>
