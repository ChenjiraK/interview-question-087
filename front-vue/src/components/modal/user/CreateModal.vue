<script setup lang="ts">
import { computed, ref, watch, watchPostEffect } from 'vue'
import type { PersonForm, PersonSubmission } from '@/types/person'
import { calculateAge } from '@/utils/person'
import InputText from '@/components/ui/InputText.vue'
import InputBirthDay from '@/components/ui/InputBirthDay.vue'
import TextArea from '@/components/ui/TextArea.vue'
import PrimaryButton from '@/components/ui/PrimaryButton.vue'
import CancelButton from '@/components/ui/CancelButton.vue'

const props = withDefaults(
  defineProps<{
    data?: Partial<PersonForm> | null
    saving?: boolean
    error?: string
  }>(),
  {
    data: null,
    saving: false,
    error: '',
  },
)

const emit = defineEmits<{ close: []; save: [person: PersonSubmission] }>()
const detailsDialog = ref<HTMLDialogElement | null>(null)
const form = ref<PersonForm>({ firstName: '', lastName: '', birthday: '', address: '' })
const validationError = ref('')

watch(
  () => props.data,
  (data) => {
    validationError.value = ''
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
const today = computed(() => {
  const date = new Date()
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
})

function savePerson() {
  if (props.saving) return
  validationError.value = ''
  if (!form.value.birthday || !form.value.address.trim()) {
    validationError.value = 'กรุณากรอกวันเกิดและที่อยู่ให้ครบถ้วน'
    return
  }
  emit('save', {
    ...form.value,
    address: form.value.address.trim(),
    name: `${form.value.firstName} ${form.value.lastName}`,
    age: age.value,
  })
}

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
    aria-labelledby="create-title"
    class="fixed inset-0 m-auto w-[calc(100%-2rem)] max-w-3xl overflow-y-auto rounded-2xl bg-white p-0 text-slate-900 shadow-xl backdrop:bg-black/25"
    @cancel="props.saving && $event.preventDefault()"
    @close="emit('close')"
  >
    <form v-if="data" class="px-6 pt-8 pb-4 sm:px-12 sm:pt-10" @submit.prevent="savePerson">
      <h2 id="create-title" class="sr-only">เพิ่มข้อมูล</h2>
      <p v-if="validationError || error" role="alert" class="mb-6 text-sm text-red-600">
        {{ validationError || error }}
      </p>

      <fieldset
        :disabled="saving"
        class="grid min-w-0 grid-cols-1 items-start gap-x-6 gap-y-6 sm:grid-cols-[7rem_minmax(0,1fr)] sm:gap-y-8"
      >
        <label for="create-first-name" class="font-medium sm:pt-3 text-slate-500"
          >ชื่อ - สกุล</label
        >
        <div class="grid min-w-0 grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-8">
          <InputText
            id="create-first-name"
            v-model.trim="form.firstName"
            name="firstName"
            autocomplete="given-name"
            aria-label="ชื่อ"
            required
            autofocus
          />
          <InputText
            id="create-last-name"
            v-model.trim="form.lastName"
            name="lastName"
            autocomplete="family-name"
            aria-label="นามสกุล"
            required
          />
        </div>

        <label for="create-birthday" class="font-medium sm:pt-3 text-slate-500">วันเกิด</label>
        <InputBirthDay
          id="create-birthday"
          v-model="form.birthday"
          name="birthday"
          :max="today"
          required
        />

        <span id="create-age-label" class="font-medium text-slate-500">อายุ</span>
        <output
          aria-labelledby="create-age-label"
          for="create-birthday"
          aria-live="polite"
          class="flex gap-4"
        >
          <span>{{ age ?? 'xx' }}</span>
          <span>ปี</span>
        </output>

        <label for="create-address" class="font-medium sm:pt-2 text-slate-500">ที่อยู่</label>
        <TextArea
          id="create-address"
          v-model.trim="form.address"
          name="address"
          autocomplete="street-address"
          required
        />
      </fieldset>

      <footer class="flex justify-end gap-3 sm:mt-14">
        <PrimaryButton
          :text="saving ? 'กำลังบันทึก...' : 'บันทึก'"
          type="submit"
          :disabled="saving"
        />
        <CancelButton text="ยกเลิก" :disabled="saving" @on-click="detailsDialog?.close()" />
      </footer>
    </form>
  </dialog>
</template>
