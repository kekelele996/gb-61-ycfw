import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getWrongQuestionStats } from '@/api/quiz'
import { useAuthStore } from '@/stores/authStore'

// Tracks the unmastered count / mastery progress of the current user's
// wrong-question book. Resets to zero after logout.
export const useWrongBookStore = defineStore('wrongBook', () => {
  const total = ref(0)
  const unmastered = ref(0)
  const mastered = ref(0)
  const loaded = ref(false)

  async function load(force = false) {
    const auth = useAuthStore()
    if (!auth.isLoggedIn) {
      reset()
      return
    }
    if (loaded.value && !force) return
    try {
      const stats = await getWrongQuestionStats()
      total.value = stats.total
      unmastered.value = stats.unmastered
      mastered.value = stats.mastered
      loaded.value = true
    } catch {
      // The interceptor already surfaces the error; keep previous values.
    }
  }

  function reset() {
    total.value = 0
    unmastered.value = 0
    mastered.value = 0
    loaded.value = false
  }

  return { total, unmastered, mastered, loaded, load, reset }
})
