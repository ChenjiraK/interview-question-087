<script setup lang="ts">
import { computed } from 'vue'
import type { Pagination } from '@/types/person'

const props = withDefaults(
  defineProps<{
    pagination: Pagination
    disabled?: boolean
  }>(),
  { disabled: false },
)

const emit = defineEmits<{ pageChange: [page: number] }>()
const totalPages = computed(() => Math.max(1, props.pagination.total_pages))
const firstRecord = computed(() =>
  props.pagination.total > 0 ? (props.pagination.page - 1) * props.pagination.limit + 1 : 0,
)
const lastRecord = computed(() =>
  Math.min(props.pagination.page * props.pagination.limit, props.pagination.total),
)

function changePage(page: number) {
  if (props.disabled || page < 1 || page > totalPages.value || page === props.pagination.page)
    return
  emit('pageChange', page)
}
</script>

<template>
  <div
    class="flex flex-wrap items-center justify-between gap-4 border-t border-slate-200 px-6 py-4 text-xs text-slate-500"
  >
    <p aria-live="polite">
      แสดง {{ firstRecord }}–{{ lastRecord }} จาก {{ pagination.total }} รายการ
    </p>
    <nav aria-label="เปลี่ยนหน้ารายชื่อผู้ใช้" class="flex items-center gap-3">
      <button
        type="button"
        class="rounded-lg border border-slate-300 px-3 py-2 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
        :disabled="disabled || pagination.page <= 1"
        @click="changePage(pagination.page - 1)"
      >
        ก่อนหน้า
      </button>
      <span aria-live="polite">หน้า {{ pagination.page }} / {{ totalPages }}</span>
      <button
        type="button"
        class="rounded-lg border border-slate-300 px-3 py-2 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
        :disabled="disabled || pagination.page >= totalPages"
        @click="changePage(pagination.page + 1)"
      >
        ถัดไป
      </button>
    </nav>
  </div>
</template>
