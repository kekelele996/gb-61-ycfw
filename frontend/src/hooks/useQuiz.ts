import { ref } from 'vue'

// Local answer-sheet state for one quiz attempt. The question bank itself is
// served by the backend; see @/api/quiz and @/constants/quiz.
export function useQuiz() {
  const answers = ref<Record<number, number>>({})
  const submitted = ref(false)

  function reset() {
    answers.value = {}
    submitted.value = false
  }

  return { answers, submitted, reset }
}
