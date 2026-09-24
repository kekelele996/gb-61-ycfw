<template>
  <div class="page">
    <h1>养护知识小测验</h1>
    <el-alert
      :title="auth.isLoggedIn
        ? `共 ${bank.length} 题，交卷后答错的题目会自动收入错题本`
        : `共 ${bank.length} 题，未登录也可交卷查看成绩，但不会留下学习记录`"
      :type="auth.isLoggedIn ? 'success' : 'info'"
      :closable="false"
      show-icon
    />
    <div v-loading="loading">
      <QuizCard
        v-for="(q, i) in bank"
        :key="q.id"
        :question="q"
        :index="i"
        :submitted="submitted"
        :disabled="submitted"
        :model-value="answers[q.id]"
        @update:model-value="setAnswer(q.id, $event)"
      />
    </div>
    <div class="actions">
      <el-button v-if="!submitted" type="primary" size="large" :loading="submitting" @click="submitQuizAction">交卷</el-button>
      <template v-else>
        <el-result :icon="result && result.correct >= result.total * 0.6 ? 'success' : 'warning'" :title="`得分：${result?.score ?? 0} / ${bank.length}`">
          <template #sub-title>
            正确 {{ result?.correct ?? 0 }} 题，错题 {{ result?.wrong_count ?? 0 }} 题，解析见每道题下方
            <div v-if="auth.isLoggedIn && result && result.wrong_count > 0">
              答错的题目已加入<router-link to="/wrong-book">错题本</router-link>，可前往重做巩固
            </div>
            <div v-else-if="!auth.isLoggedIn">
              <router-link to="/login">登录</router-link>后交卷，错题可自动收录进错题本
            </div>
            <div v-else-if="result && result.wrong_count === 0">全部答对，错题本无新增 🎉</div>
          </template>
          <template #extra>
            <el-button type="primary" @click="restart">重新作答</el-button>
          </template>
        </el-result>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import QuizCard from '@/components/common/QuizCard.vue'
import { listQuizQuestions, submitQuiz } from '@/api/quiz'
import type { QuizQuestion, QuizSubmitResult } from '@/types/api'
import { useAuthStore } from '@/stores/authStore'
import { useWrongBookStore } from '@/stores/wrongBookStore'

const auth = useAuthStore()
const wrongBook = useWrongBookStore()

const bank = ref<QuizQuestion[]>([])
const answers = ref<Record<number, number>>({})
const submitted = ref(false)
const submitting = ref(false)
const loading = ref(true)
const result = ref<QuizSubmitResult | null>(null)

onMounted(async () => {
  try {
    bank.value = await listQuizQuestions()
  } finally {
    loading.value = false
  }
})

function setAnswer(qid: number, v: number) {
  answers.value[qid] = v
}

async function submitQuizAction() {
  const unanswered = bank.value.filter((q) => answers.value[q.id] === undefined)
  if (unanswered.length > 0) {
    ElMessage.warning(`还有 ${unanswered.length} 题未作答`)
    return
  }
  submitting.value = true
  try {
    const payload = bank.value.map((q) => ({ question_id: q.id, selected_index: answers.value[q.id] }))
    result.value = await submitQuiz(payload)
    submitted.value = true
    if (auth.isLoggedIn && result.value.wrong_count > 0) {
      ElMessage.success('答错的题目已收入错题本')
      await wrongBook.loadStats()
    }
  } finally {
    submitting.value = false
  }
}

function restart() {
  answers.value = {}
  submitted.value = false
  result.value = null
}
</script>

<style scoped>
.page { max-width: 900px; margin: 0 auto; }
.actions { margin: 20px 0; text-align: center; }
</style>
