import request from '@/utils/request'
import type {
  QuizQuestion,
  QuizRetryResult,
  QuizSubmitAnswer,
  QuizSubmitResult,
  WrongBookStats,
  WrongQuestion,
} from '@/types/api'

export function listQuizQuestions() {
  return request.get<never, QuizQuestion[]>('/quiz/questions')
}

export function submitQuiz(answers: QuizSubmitAnswer[]) {
  return request.post<never, QuizSubmitResult>('/quiz/submit', { answers })
}

export function listWrongBook() {
  return request.get<never, WrongQuestion[]>('/quiz/wrong-book')
}

export function getWrongBookStats() {
  return request.get<never, WrongBookStats>('/quiz/wrong-book/stats')
}

export function retryWrongQuestion(questionId: number, selectedIndex: number) {
  return request.post<never, QuizRetryResult>(`/quiz/wrong-book/${questionId}/retry`, {
    selected_index: selectedIndex,
  })
}
