<script setup>
const [model, modifiers] = defineModel({
  type: String,
  default: '',
  set(value) {
    return modifiers.trim ? value.trim() : value
  },
})

const emit = defineEmits(['onChange'])
</script>

<template>
  <textarea
    v-model="model"
    rows="4"
    class="min-h-28 w-full min-w-0 resize-y rounded-2xl border border-gray-400 bg-white px-3 py-2 outline-none focus:border-sky-600 focus:ring-2 focus:ring-sky-200 disabled:bg-gray-50"
    @change="emit('onChange', model)"
  />
</template>
