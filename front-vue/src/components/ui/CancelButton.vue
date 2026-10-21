<script setup>
defineProps({
  text: {
    type: String,
    default: 'Cancel',
  },
})

const emit = defineEmits(['onClick'])
</script>

<template>
  <button
    type="button"
    class="rounded-lg bg-red-500 px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-red-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-500"
    @click="emit('onClick', $event)"
  >
    {{ text }}
  </button>
</template>
