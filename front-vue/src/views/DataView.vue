<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { createUser, getUserById, getUsers } from '@/services/users'
import { getApiErrorMessage } from '@/services/api'
import type { Pagination as PaginationMetadata, Person, PersonForm } from '@/types/person'
import { calculateAge, formatBirthday } from '@/utils/person'
import Pagination from '@/components/ui/Pagination.vue'
import ViewModal from '@/components/modal/user/ViewModal.vue'
import CreateModal from '@/components/modal/user/CreateModal.vue'

const people = ref<Person[]>([])
const pagination = ref<PaginationMetadata>({ page: 1, limit: 10, total: 0, total_pages: 0 })
const selectedPerson = ref<Person | null>(null)
const createData = ref<Partial<PersonForm> | null>(null)
const isLoading = ref(false)
const isSaving = ref(false)
const viewingId = ref<Person['id'] | null>(null)
const listError = ref('')
const detailError = ref('')
const createError = ref('')

async function loadPeople(page = pagination.value.page) {
  if (isLoading.value) return
  isLoading.value = true
  listError.value = ''
  try {
    let result = await getUsers(page, pagination.value.limit)
    // The requested page may disappear if records were removed by another client.
    const lastPage = Math.max(1, result.pagination.total_pages)
    if (result.pagination.page > lastPage) {
      result = await getUsers(lastPage, result.pagination.limit)
    }
    people.value = result.data
    pagination.value = result.pagination
  } catch (error) {
    listError.value = getApiErrorMessage(error, 'โหลดรายชื่อไม่สำเร็จ กรุณาลองอีกครั้ง')
  } finally {
    isLoading.value = false
  }
}

function openCreate() {
  createError.value = ''
  createData.value = {}
}

async function createPerson(person: PersonForm) {
  if (isSaving.value) return
  isSaving.value = true
  createError.value = ''
  try {
    await createUser(person)
    createData.value = null
    // Refresh so the table also includes records created by other clients.
    await loadPeople()
  } catch (error) {
    createError.value = getApiErrorMessage(error, 'บันทึกข้อมูลไม่สำเร็จ กรุณาลองอีกครั้ง')
  } finally {
    isSaving.value = false
  }
}

async function viewPerson(person: Person) {
  if (viewingId.value !== null) return
  viewingId.value = person.id
  detailError.value = ''
  try {
    selectedPerson.value = await getUserById(person.id)
  } catch (error) {
    detailError.value = getApiErrorMessage(error, 'โหลดข้อมูลผู้ใช้ไม่สำเร็จ กรุณาลองอีกครั้ง')
  } finally {
    viewingId.value = null
  }
}

onMounted(() => loadPeople())
</script>

<template>
  <main class="min-h-screen bg-slate-50 px-4 py-10 text-slate-900 sm:px-8 sm:py-16">
    <div class="mx-auto max-w-6xl">
      <div class="mb-4 flex justify-end">
        <button
          type="button"
          :disabled="isLoading || isSaving"
          class="rounded-lg bg-green-700 px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-green-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-green-500"
          @click="openCreate"
        >
          ADD
        </button>
      </div>
      <p v-if="detailError" role="alert" class="mb-4 text-sm text-red-600">{{ detailError }}</p>
      <div v-if="listError" role="alert" class="mb-4 flex items-center gap-3 text-sm text-red-600">
        <span>{{ listError }}</span>
        <button type="button" class="underline" :disabled="isLoading" @click="loadPeople()">
          ลองอีกครั้ง
        </button>
      </div>
      <section
        aria-label="รายชื่อผู้ใช้"
        :aria-busy="isLoading"
        class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm"
      >
        <div class="overflow-x-auto">
          <table class="w-full min-w-212.5 text-left text-sm">
            <thead
              class="border-b border-slate-200 bg-slate-50 text-xs font-semibold tracking-wide text-slate-500 uppercase"
            >
              <tr>
                <th scope="col" class="px-6 py-4">ID</th>
                <th scope="col" class="px-6 py-4">ชื่อ-นามสกุล</th>
                <th scope="col" class="px-6 py-4">ที่อยู่</th>
                <th scope="col" class="px-6 py-4">วันเกิด</th>
                <th scope="col" class="px-6 py-4">อายุ</th>
                <th scope="col" class="px-6 py-4 text-center">Action</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="isLoading">
                <td colspan="6" class="px-6 py-10 text-center text-slate-500" role="status">
                  กำลังโหลดข้อมูล...
                </td>
              </tr>
              <tr v-else-if="!people.length && !listError">
                <td colspan="6" class="px-6 py-10 text-center text-slate-500">
                  ยังไม่มีข้อมูลผู้ใช้
                </td>
              </tr>
              <tr
                v-for="person in people"
                v-show="!isLoading"
                :key="person.id"
                class="transition-colors hover:bg-slate-50/80"
              >
                <td class="px-6 py-5 text-slate-400 tabular-nums">{{ person.id }}</td>
                <th scope="row" class="px-6 py-5 font-medium whitespace-nowrap">
                  {{ person.name }}
                </th>
                <td class="px-6 py-5 leading-6 text-slate-500">{{ person.address }}</td>
                <td class="px-6 py-5 whitespace-nowrap text-slate-600">
                  <time :datetime="person.birthday">{{ formatBirthday(person.birthday) }}</time>
                </td>
                <td class="px-6 py-5 text-slate-600 tabular-nums">
                  {{ calculateAge(person.birthday) ?? '-' }}
                </td>
                <td class="px-6 py-5 text-right">
                  <button
                    type="button"
                    class="inline-flex items-center gap-2 rounded-lg bg-blue-500 px-3 py-2 text-xs font-semibold whitespace-nowrap text-white transition"
                    :disabled="viewingId !== null"
                    :aria-busy="viewingId === person.id"
                    @click="viewPerson(person)"
                  >
                    <svg
                      aria-hidden="true"
                      class="size-4"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="1.5"
                    >
                      <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z" />
                      <circle cx="12" cy="12" r="3" />
                    </svg>
                    {{ viewingId === person.id ? 'กำลังโหลด...' : 'View data'
                    }}<span class="sr-only"> for {{ person.name }}</span>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <footer>
          <Pagination
            :pagination="pagination"
            :disabled="isLoading || isSaving"
            @page-change="loadPeople"
          />
        </footer>
      </section>
    </div>
    <ViewModal :data="selectedPerson" @close="selectedPerson = null" />
    <CreateModal
      :data="createData"
      :saving="isSaving"
      :error="createError"
      @save="createPerson"
      @close="createData = null"
    />
  </main>
</template>
