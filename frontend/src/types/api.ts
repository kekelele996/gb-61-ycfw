export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export interface PageData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface UserInfo {
  id: number
  username: string
  email: string
  nickname: string
  avatar: string
  bio: string
  role: 'user' | 'admin'
  created_at: string
}

export interface CareReminder {
  id: number
  user_id: number
  plant_species_id: number
  task_title: string
  remind_date: string
  frequency: string
  status: 'pending' | 'done' | 'overdue'
  created_at: string
}

export interface UserGarden {
  id: number
  user_id: number
  plant_species_id: number
  nickname: string
  owned_since: string
  location: string
  care_reminder_id: number
  created_at: string
}

export interface DiseasePest {
  id: number
  plant_species_id: number
  name: string
  symptoms: string
  cause: string
  treatment: string
  recommended_medicine: string
  images: string
  keywords: string
  created_at: string
}

export interface Question {
  id: number
  user_id: number
  title: string
  content: string
  images: string
  status: string
  created_at: string
}

export interface Answer {
  id: number
  question_id: number
  user_id: number
  content: string
  is_best: boolean
  like_count: number
  created_at: string
}

export interface QuizOption {
  index: number
  text: string
}

export interface QuizQuestion {
  id: number
  question: string
  options: QuizOption[]
  answer: number
  explanation: string
}

export interface QuizSubmitAnswer {
  question_id: number
  selected_index: number
}

export interface QuizSubmitResult {
  total: number
  correct: number
  score: number
  wrong_count: number
  wrong_ids: number[]
  saved: boolean
}

export interface QuizRetryResult {
  question_id: number
  correct: boolean
  mastered: boolean
}

export interface WrongQuestion {
  id: number
  question_id: number
  selected_index: number
  mastered: boolean
  question: string
  options: QuizOption[]
  answer: number
  explanation: string
  created_at: string
  updated_at: string
}

export interface WrongBookStats {
  total: number
  unmastered: number
  mastered: number
  progress: number
}
