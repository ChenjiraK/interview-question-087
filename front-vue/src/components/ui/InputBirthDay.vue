<script setup>
const model = defineModel({ type: String, default: '' })
const emit = defineEmits(['onChange'])
</script>

<template>
  <input
    v-model="model"
    type="date"
    autocomplete="bday"
    min="0001-01-01"
    class="h-11 w-full min-w-0 rounded-md border border-gray-400 bg-white px-3 outline-none focus:border-sky-600 focus:ring-2 focus:ring-sky-200 disabled:bg-gray-50"
    @change="emit('onChange', model)"
  />
</template>
