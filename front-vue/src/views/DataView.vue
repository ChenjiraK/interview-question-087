<script setup>
import { ref } from 'vue'
import ViewModal from '../components/modal/ViewModal.vue'
import CreateModal from '../components/modal/CreateModal.vue'
import { calculateAge, formatBirthday } from '../utils/person'

// Sample records. Replace with API data when a backend is available.
const people = ref([
  { id: 1, name: 'Anan Srisuk', address: '128 Sukhumvit Road, Bangkok', birthday: '1995-03-15' },
  { id: 2, name: 'Siriporn Jaidee', address: '45 Nimman Road, Chiang Mai', birthday: '1998-11-22' },
  { id: 3, name: 'Nattapong Saetang', address: '76 Phuket Road, Phuket', birthday: '1992-07-08' },
  {
    id: 4,
    name: 'Pimchanok Wongchai',
    address: '23 Mittraphap Road, Khon Kaen',
    birthday: '2000-01-30',
  },
  {
    id: 5,
    name: 'Kittisak Boonmee',
    address: '91 Pattaya Klang Road, Chon Buri',
    birthday: '1988-10-21',
  },
])

const selectedPerson = ref(null)
const createData = ref(null)

function createPerson(person) {
  const id = Math.max(0, ...people.value.map((record) => record.id)) + 1
  people.value.push({ ...person, id })
  createData.value = null
}

function viewPerson(person) {
  selectedPerson.value = person
}
</script>

<template>
  <main class="min-h-screen bg-slate-50 px-4 py-10 text-slate-900 sm:px-8 sm:py-16">
    <div class="mx-auto max-w-6xl">
      <div class="mb-4 flex justify-end">
        <button
          type="button"
          class="rounded-lg bg-green-700 px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-green-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-green-500"
          @click="createData = {}"
        >
          ADD
        </button>
      </div>
      <section
        aria-labelledby="records-title"
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
              <tr
                v-for="person in people"
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
                  {{ calculateAge(person.birthday) }}
                </td>
                <td class="px-6 py-5 text-right">
                  <button
                    type="button"
                    class="inline-flex items-center gap-2 rounded-lg bg-blue-500 px-3 py-2 text-xs font-semibold whitespace-nowrap text-white transition"
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
                    View data<span class="sr-only"> for {{ person.name }}</span>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <footer class="border-t border-slate-200 px-6 py-4 text-xs text-slate-500">
          Showing {{ people.length }} records
        </footer>
      </section>
    </div>
    <ViewModal mode="view" :data="selectedPerson" @close="selectedPerson = null" />
    <CreateModal :data="createData" @save="createPerson" @close="createData = null" />
  </main>
</template>
