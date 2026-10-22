import { createRouter, createWebHistory } from 'vue-router'
import DataView from '@/views/DataView.vue'

export default createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [{ path: '/', name: 'data', component: DataView }],
})
