<script setup lang="ts">
import { computed, ref, watch, watchPostEffect } from 'vue'
import type { Person, PersonForm } from '@/types/person'
import { calculateAge } from '@/utils/person'
import InputText from '@/components/ui/InputText.vue'
import InputBirthDay from '@/components/ui/InputBirthDay.vue'
import TextArea from '@/components/ui/TextArea.vue'
import CancelButton from '@/components/ui/CancelButton.vue'

const props = withDefaults(defineProps<{ data?: Person | null }>(), {
  data: null,
})

const emit = defineEmits<{ close: [] }>()
const detailsDialog = ref<HTMLDialogElement | null>(null)
const form = ref<PersonForm>({ firstName: '', lastName: '', birthday: '', address: '' })

watch(
  () => props.data,
  (data) => {
    form.value = {
      firstName: data?.firstName ?? '',
      lastName: data?.lastName ?? '',
      birthday: data?.birthday ?? '',
      address: data?.address ?? '',
    }
  },
  { immediate: true },
)

const age = computed(() => (form.value.birthday ? calculateAge(form.value.birthday) : null))
// Wait for the dialog and its content to render before opening it.
watchPostEffect(() => {
  const dialog = detailsDialog.value
  if (!dialog) return

  if (props.data && !dialog.open) {
    dialog.showModal()
  } else if (!props.data && dialog.open) {
    dialog.close()
  }
})
</script>

<template>
  <dialog
    ref="detailsDialog"
    aria-labelledby="view-title"
    class="fixed inset-0 m-auto w-[calc(100%-2rem)] max-w-3xl overflow-y-auto rounded-2xl bg-white p-8 text-slate-900 shadow-xl backdrop:bg-black/25"
    @close="emit('close')"
  >
    <h2 id="view-title" class="sr-only">ข้อมูลผู้ใช้</h2>
    <div
      class="grid grid-cols-1 items-start gap-x-6 gap-y-6 sm:grid-cols-[7rem_minmax(0,1fr)] sm:gap-y-8"
    >
      <label for="view-first-name" class="font-medium sm:pt-3 text-slate-500">ชื่อ - สกุล</label>
      <div class="grid min-w-0 grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-8">
        <InputText
          id="view-first-name"
          v-model.trim="form.firstName"
          name="firstName"
          aria-label="ชื่อ"
          disabled
        />
        <InputText
          id="view-last-name"
          v-model.trim="form.lastName"
          name="lastName"
          aria-label="นามสกุล"
          disabled
        />
      </div>

      <label for="view-birthday" class="font-medium sm:pt-3 text-slate-500">วันเกิด</label>
      <InputBirthDay id="view-birthday" v-model="form.birthday" name="birthday" disabled />

      <span id="view-age-label" class="font-medium text-slate-500">อายุ</span>
      <output
        aria-labelledby="view-age-label"
        for="view-birthday"
        aria-live="polite"
        class="flex gap-4"
      >
        <span>{{ age ?? '-' }}</span>
        <span>ปี</span>
      </output>

      <label for="view-address" class="font-medium sm:pt-2 text-slate-500">ที่อยู่</label>
      <TextArea id="view-address" v-model.trim="form.address" name="address" disabled />
    </div>

    <footer class="flex justify-end gap-3 sm:mt-14">
      <CancelButton text="ปิด" @on-click="detailsDialog?.close()" />
    </footer>
  </dialog>
</template>
