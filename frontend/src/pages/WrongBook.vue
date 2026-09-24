<template>
  <div class="page" v-loading="loading">
    <div class="header">
      <h1>错题本</h1>
      <el-button @click="$router.push('/quiz')">返回养护测验</el-button>
    </div>

    <el-card class="stats-card">
      <div class="stats">
        <div class="stat">
          <div class="num danger">{{ stats.unmastered }}</div>
          <div class="label">未掌握</div>
        </div>
        <div class="stat">
          <div class="num success">{{ stats.mastered }}</div>
          <div class="label">已掌握</div>
        </div>
        <div class="stat">
          <div class="num">{{ stats.total }}</div>
          <div class="label">错题总数</div>
        </div>
        <div class="progress-box">
          <div class="label">掌握进度 {{ stats.progress }}%</div>
          <el-progress :percentage="stats.progress" :stroke-width="14" status="success" />
        </div>
      </div>
    </el-card>

    <el-radio-group v-model="filter" class="filter">
      <el-radio-button value="all">全部 {{ stats.total }}</el-radio-button>
      <el-radio-button value="unmastered">未掌握 {{ stats.unmastered }}</el-radio-button>
      <el-radio-button value="mastered">已掌握 {{ stats.mastered }}</el-radio-button>
    </el-radio-group>

    <el-empty v-if="!loading && filtered.length === 0" description="错题本里还没有题目，去测验试试吧">
      <el-button type="primary" @click="$router.push('/quiz')">去做测验</el-button>
    </el-empty>

    <el-card v-for="(item, i) in filtered" :key="item.id" class="wrong-card">
      <QuizCard
        :question="toQuestion(item)"
        :index="i"
        :model-value="retrySelection[item.question_id]"
        :submitted="false"
        :show-wrong-tag="!item.mastered"
        :mastered="item.mastered"
        @update:model-value="(v) => onSelect(item.question_id, v)"
      />
      <div class="wrong-meta">
        <template v-if="!item.mastered">
          最近一次误选：<el-tag type="danger" size="small">{{ optionText(item, item.selected_index) }}</el-tag>
        </template>
        <template v-else>
          <el-tag type="success" size="small">已掌握</el-tag>
        </template>
      </div>
      <div class="wrong-actions">
        <el-button
          type="primary"
          :loading="retrying[item.question_id]"
          :disabled="retrySelection[item.question_id] === undefined"
          @click="retry(item)"
        >
          重做此题
        </el-button>
        <el-button text @click="revealAnswers[item.question_id] = !revealAnswers[item.question_id]">
          {{ revealAnswers[item.question_id] ? '收起解析' : '查看解析' }}
        </el-button>
      </div>
      <el-alert
        v-if="revealAnswers[item.question_id]"
        class="retry-result"
        :title="`正确答案：${optionText(item, item.answer)}。解析：${item.explanation}`"
        type="info"
        :closable="false"
        :show-icon="false"
      />
      <el-alert
        v-if="retryResult[item.question_id] === 'correct'"
        class="retry-result"
        title="回答正确，已记为掌握 🎉"
        type="success"
        :closable="false"
      />
      <el-alert
        v-else-if="retryResult[item.question_id] === 'wrong'"
        class="retry-result"
        title="回答错误，已更新你的选择，此题保持未掌握，再试一次吧"
        type="error"
        :closable="false"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import QuizCard from '@/components/common/QuizCard.vue'
import { listWrongBook, retryWrongQuestion } from '@/api/quiz'
import type { QuizQuestion, WrongQuestion } from '@/types/api'
import { useWrongBookStore } from '@/stores/wrongBookStore'

const wrongBook = useWrongBookStore()
const stats = computed(() => wrongBook.stats)

const items = ref<WrongQuestion[]>([])
const loading = ref(true)
const filter = ref<'all' | 'unmastered' | 'mastered'>('unmastered')
const retrySelection = reactive<Record<number, number>>({})
const retrying = reactive<Record<number, boolean>>({})
const revealAnswers = reactive<Record<number, boolean>>({})
const retryResult = reactive<Record<number, 'correct' | 'wrong'>>({})

const filtered = computed(() => {
  if (filter.value === 'unmastered') return items.value.filter((x) => !x.mastered)
  if (filter.value === 'mastered') return items.value.filter((x) => x.mastered)
  return items.value
})

onMounted(async () => {
  try {
    const [list] = await Promise.all([listWrongBook(), wrongBook.loadStats()])
    items.value = list
  } finally {
    loading.value = false
  }
})

function toQuestion(item: WrongQuestion): QuizQuestion {
  return {
    id: item.question_id,
    question: item.question,
    options: item.options,
    answer: item.answer,
    explanation: item.explanation,
  }
}

function optionText(item: WrongQuestion, index: number): string {
  return item.options.find((o) => o.index === index)?.text ?? ''
}

function onSelect(questionId: number, v: number) {
  retrySelection[questionId] = v
}

async function retry(item: WrongQuestion) {
  const selected = retrySelection[item.question_id]
  if (selected === undefined) return
  retrying[item.question_id] = true
  try {
    const res = await retryWrongQuestion(item.question_id, selected)
    retryResult[item.question_id] = res.correct ? 'correct' : 'wrong'
    if (res.correct) {
      ElMessage.success('回答正确，已记为掌握')
      item.mastered = true
      revealAnswers[item.question_id] = false
      delete retrySelection[item.question_id]
    } else {
      ElMessage.error('回答错误，已更新你的选择，此题保持未掌握')
      item.mastered = false
      item.selected_index = selected
    }
    await wrongBook.loadStats()
  } finally {
    retrying[item.question_id] = false
  }
}
</script>

<style scoped>
.page { max-width: 900px; margin: 0 auto; }
.header { display: flex; align-items: center; justify-content: space-between; }
.stats-card { margin: 12px 0; }
.stats { display: flex; align-items: center; gap: 40px; }
.stat { text-align: center; }
.num { font-size: 28px; font-weight: 700; }
.num.danger { color: #e6532f; }
.num.success { color: #3c8d5c; }
.label { color: #666; font-size: 14px; margin-bottom: 4px; }
.progress-box { flex: 1; max-width: 320px; }
.filter { margin-bottom: 12px; }
.wrong-card { margin-bottom: 12px; }
.wrong-meta { color: #666; font-size: 14px; margin: 4px 0; }
.wrong-actions { margin-top: 8px; }
.retry-result { margin-top: 8px; }
</style>
