<template>
  <div class="page">
    <h1>养护知识小测验</h1>
    <el-tabs v-model="activeTab">
      <!-- ============ 测验 ============ -->
      <el-tab-pane name="quiz">
        <template #label>测验</template>
        <el-alert
          :title="`共 ${questions.length} 题，交卷后展示成绩与错题解析`"
          :description="auth.isLoggedIn ? '答错的题目会自动收入错题本，同一道题只保留一条' : '当前未登录，可以交卷查看成绩，但不会留下学习记录'"
          :type="auth.isLoggedIn ? 'info' : 'warning'"
          :closable="false"
          show-icon
        />
        <div v-loading="loading" class="quiz-list">
          <QuizCard
            v-for="(q, i) in questions"
            :key="q.id"
            :question="q"
            :index="i"
            :submitted="submitted"
            :model-value="answers[q.id]"
            :answer-index="resultMap[q.id]?.answer_index"
            :explanation="resultMap[q.id]?.explanation"
            @update:model-value="answers[q.id] = $event"
          />
        </div>
        <div class="actions">
          <el-button v-if="!submitted" type="primary" size="large" :loading="submitting" :disabled="loading || questions.length === 0" @click="submitQuizPaper">交卷</el-button>
          <template v-else>
            <el-result :icon="score >= questions.length * 0.6 ? 'success' : 'warning'" :title="`得分：${score} / ${questions.length}`">
              <template #sub-title>
                正确 {{ correctCount }} 题，错题解析见每道题下方
                <span v-if="wrongCount > 0 && auth.isLoggedIn" class="hint-wrong">，{{ wrongCount }} 道错题已收入错题本</span>
                <span v-else-if="wrongCount > 0" class="hint-wrong">，登录后错题可自动收录</span>
              </template>
              <template #extra>
                <el-button type="primary" @click="restart">重新作答</el-button>
                <el-button v-if="wrongCount > 0 && auth.isLoggedIn" plain @click="goWrongBook">查看错题本</el-button>
              </template>
            </el-result>
          </template>
        </div>
      </el-tab-pane>

      <!-- ============ 错题本 ============ -->
      <el-tab-pane name="wrong" :disabled="!auth.isLoggedIn">
        <template #label>
          错题本
          <el-badge v-if="wrongBook.unmastered > 0" :value="wrongBook.unmastered" class="tab-badge" />
        </template>

        <el-empty v-if="!auth.isLoggedIn" description="登录后错题会自动收录，可随时重做">
          <el-button type="primary" @click="$router.push({ path: '/login', query: { redirect: '/quiz' } })">去登录</el-button>
        </el-empty>

        <template v-else>
          <el-card class="stats-card">
            <div class="stats-line">
              <el-tag type="warning">未掌握 {{ wrongBook.unmastered }} 题</el-tag>
              <el-tag type="success" class="mastered-tag">已掌握 {{ wrongBook.mastered }} 题</el-tag>
              <span class="progress-text">掌握进度 {{ wrongBook.total ? Math.round((wrongBook.mastered / wrongBook.total) * 100) : 0 }}%</span>
            </div>
            <el-progress
              :percentage="wrongBook.total ? Math.round((wrongBook.mastered / wrongBook.total) * 100) : 0"
              :stroke-width="14"
              :text-inside="true"
            />
          </el-card>

          <el-radio-group v-model="masteredFilter" class="filter" @change="loadWrongList">
            <el-radio-button :value="''">全部 {{ wrongBook.total }}</el-radio-button>
            <el-radio-button :value="'false'">未掌握 {{ wrongBook.unmastered }}</el-radio-button>
            <el-radio-button :value="'true'">已掌握 {{ wrongBook.mastered }}</el-radio-button>
          </el-radio-group>

          <div v-loading="wrongLoading" class="wrong-list">
            <WrongQuestionCard
              v-for="wq in wrongList"
              :key="wq.id"
              :question="wq"
              @retried="onRetried"
            />
            <el-empty v-if="!wrongLoading && wrongList.length === 0" description="这里还没有题目" />
          </div>
        </template>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import QuizCard from '@/components/common/QuizCard.vue'
import WrongQuestionCard from '@/components/common/WrongQuestionCard.vue'
import { listQuizQuestions, submitQuiz, listWrongQuestions } from '@/api/quiz'
import type { QuizQuestion, QuizResultItem, WrongQuestion } from '@/constants/quiz'
import { useQuiz } from '@/hooks/useQuiz'
import { useAuthStore } from '@/stores/authStore'
import { useWrongBookStore } from '@/stores/wrongBookStore'

const route = useRoute()
const auth = useAuthStore()
const wrongBook = useWrongBookStore()

const { answers, submitted, reset } = useQuiz()
const activeTab = ref<'quiz' | 'wrong'>((route.query.tab as string) === 'wrong' ? 'wrong' : 'quiz')

const questions = ref<QuizQuestion[]>([])
const loading = ref(false)
const submitting = ref(false)
const results = ref<QuizResultItem[]>([])

const wrongList = ref<WrongQuestion[]>([])
const wrongLoading = ref(false)
const masteredFilter = ref<'' | 'true' | 'false'>('')

const resultMap = computed(() => {
  const m: Record<number, QuizResultItem> = {}
  for (const r of results.value) m[r.question_id] = r
  return m
})
const score = computed(() => results.value.filter((r) => r.correct).length)
const correctCount = computed(() => score.value)
const wrongCount = computed(() => results.value.length - score.value)

async function loadQuestions() {
  loading.value = true
  try {
    questions.value = await listQuizQuestions()
  } finally {
    loading.value = false
  }
}

async function submitQuizPaper() {
  const unanswered = questions.value.filter((q) => answers.value[q.id] === undefined)
  if (unanswered.length > 0) {
    ElMessage.warning(`还有 ${unanswered.length} 题未作答`)
    return
  }
  submitting.value = true
  try {
    const payload = questions.value.map((q) => ({ question_id: q.id, selected_index: answers.value[q.id] }))
    const res = await submitQuiz(payload)
    results.value = res.results
    submitted.value = true
    if (auth.isLoggedIn) {
      await wrongBook.load(true)
    }
  } finally {
    submitting.value = false
  }
}

function restart() {
  reset()
  results.value = []
}

async function loadWrongList() {
  wrongLoading.value = true
  try {
    const mastered = masteredFilter.value === '' ? undefined : masteredFilter.value === 'true'
    wrongList.value = await listWrongQuestions(mastered)
  } finally {
    wrongLoading.value = false
  }
}

async function onRetried({ questionId, correct }: { questionId: number; correct: boolean }) {
  await wrongBook.load(true)
  if (correct && masteredFilter.value === 'false') {
    // A newly mastered question leaves the unmastered view.
    wrongList.value = wrongList.value.filter((w) => w.question_id !== questionId)
  } else {
    await loadWrongList()
  }
}

function goWrongBook() {
  activeTab.value = 'wrong'
  loadWrongList()
}

// Lazily load the wrong-question book the first time the tab opens.
watch(activeTab, (tab) => {
  if (tab === 'wrong' && auth.isLoggedIn && wrongList.value.length === 0 && !wrongLoading.value) {
    loadWrongList()
  }
})

onMounted(async () => {
  await loadQuestions()
  if (auth.isLoggedIn) {
    await wrongBook.load(true)
    if (activeTab.value === 'wrong') await loadWrongList()
  }
})
</script>

<style scoped>
.page { max-width: 900px; margin: 0 auto; }
.quiz-list { margin-top: 12px; }
.actions { margin: 20px 0; text-align: center; }
.hint-wrong { color: #e6a23c; }
.tab-badge { margin-left: 6px; }
.stats-card { margin: 8px 0 16px; }
.stats-line { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }
.mastered-tag { margin-right: 8px; }
.progress-text { margin-left: auto; color: #606266; font-size: 13px; }
.filter { margin-bottom: 12px; }
</style>
