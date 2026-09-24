import request from '@/utils/request'
import type { QuizQuestion, QuizResultItem, QuizSubmitResult, WrongQuestion, WrongQuestionStats } from '@/constants/quiz'

// GET /quiz/questions — public question bank (no correct answer).
export function listQuizQuestions() {
  return request.get<never, QuizQuestion[]>('/quiz/questions')
}

// POST /quiz/submit — works for everyone; only logged-in users get wrong
// answers persisted into their wrong-question book.
export function submitQuiz(answers: { question_id: number; selected_index: number }[]) {
  return request.post<never, QuizSubmitResult>('/quiz/submit', { answers })
}

// GET /quiz/wrong-questions — login required.
export function listWrongQuestions(mastered?: boolean) {
  return request.get<never, WrongQuestion[]>('/quiz/wrong-questions', {
    params: mastered === undefined ? {} : { mastered },
  })
}

// GET /quiz/wrong-questions/stats — login required.
export function getWrongQuestionStats() {
  return request.get<never, WrongQuestionStats>('/quiz/wrong-questions/stats')
}

// PUT /quiz/wrong-questions/:id/retry — grade one retry.
export function retryWrongQuestion(questionId: number, selectedIndex: number) {
  return request.put<never, QuizResultItem>(
    `/quiz/wrong-questions/${questionId}/retry`,
    { selected_index: selectedIndex },
  )
}
