<script setup lang="ts">
withDefaults(
  defineProps<{
    text?: string
    type?: 'button' | 'submit' | 'reset'
  }>(),
  {
    text: '',
    type: 'button',
  },
)

const emit = defineEmits<{ onClick: [event: MouseEvent] }>()
</script>

<template>
  <button
    :type="type"
    class="rounded-lg bg-blue-500 px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-blue-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
    @click="emit('onClick', $event)"
  >
    {{ text }}
  </button>
</template>
