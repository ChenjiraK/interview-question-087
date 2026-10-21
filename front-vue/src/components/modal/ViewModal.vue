<script setup>
import { computed, ref, watch, watchPostEffect } from 'vue'
import { calculateAge } from '../../utils/person'
import InputText from '../ui/InputText.vue'
import InputBirthDay from '../ui/InputBirthDay.vue'
import TextArea from '../ui/TextArea.vue'
import PrimaryButton from '../ui/PrimaryButton.vue'
import CancelButton from '../ui/CancelButton.vue'

const props = defineProps({
  data: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['close', 'save'])
const detailsDialog = ref(null)
const form = ref({ firstName: '', lastName: '', birthday: '', address: '' })

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
const today = computed(() => {
  const date = new Date()
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
})

function savePerson() {
  emit('save', {
    ...form.value,
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
    class="fixed inset-0 m-auto w-[calc(100%-2rem)] max-w-3xl overflow-y-auto rounded-2xl bg-white p-8 text-slate-900 shadow-xl backdrop:bg-black/25"
    @close="emit('close')"
  >
      <div
        class="grid grid-cols-1 items-start gap-x-6 gap-y-6 sm:grid-cols-[7rem_minmax(0,1fr)] sm:gap-y-8"
      >
        <label for="create-first-name" class="font-medium sm:pt-3 text-slate-500">ชื่อ - สกุล</label>
        <div class="grid min-w-0 grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-8">
          <InputText
            id="create-first-name"
            v-model.trim="form.firstName"
            name="firstName"
            aria-label="ชื่อ"
            disabled
          />
          <InputText
            id="create-last-name"
            v-model.trim="form.lastName"
            name="lastName"
            aria-label="นามสกุล"
            disabled
          />
        </div>

        <label for="create-birthday" class="font-medium sm:pt-3 text-slate-500">วันเกิด</label>
        <InputBirthDay
          id="create-birthday"
          v-model="form.birthday"
          name="birthday"
          :max="today"
          disabled
        />

        <span id="create-age-label" class="font-medium text-slate-500">อายุ</span>
        <output
          aria-labelledby="create-age-label"
          for="create-birthday"
          aria-live="polite"
          class="flex gap-4"
        >
          <span>{{ age }}</span>
          <span>ปี</span>
        </output>

        <label for="create-address" class="font-medium sm:pt-2 text-slate-500">ที่อยู่</label>
        <TextArea
          id="create-address"
          v-model.trim="form.address"
          name="address"
          disabled        />
      </div>

      <footer class="flex justify-end gap-3 sm:mt-14">
        <CancelButton text="ปิด" @on-click="detailsDialog.close()" />
      </footer>
  </dialog>
</template>
