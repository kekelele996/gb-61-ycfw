// Quiz question as returned before submission (correct answer omitted).
export interface QuizQuestion {
  id: number
  question: string
  options: string[]
  answer_index?: number
  explanation?: string
}

export interface QuizResultItem {
  question_id: number
  question: string
  selected_index: number
  answer_index: number
  correct: boolean
  explanation: string
}

export interface QuizSubmitResult {
  total: number
  correct: number
  score: number
  results: QuizResultItem[]
}

export interface WrongQuestion {
  id: number
  question_id: number
  question: string
  options: string[]
  answer_index: number
  selected_index: number
  mastered: boolean
  explanation: string
  created_at: string
  updated_at: string
}

export interface WrongQuestionStats {
  total: number
  unmastered: number
  mastered: number
}
