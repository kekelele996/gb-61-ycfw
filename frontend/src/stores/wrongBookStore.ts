import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getWrongBookStats } from '@/api/quiz'
import type { WrongBookStats } from '@/types/api'

const EMPTY_STATS: WrongBookStats = { total: 0, unmastered: 0, mastered: 0, progress: 0 }

export const useWrongBookStore = defineStore('wrongBook', () => {
  const stats = ref<WrongBookStats>({ ...EMPTY_STATS })

  async function loadStats() {
    stats.value = await getWrongBookStats()
  }

  function reset() {
    stats.value = { ...EMPTY_STATS }
  }

  return { stats, loadStats, reset }
})
