<template>
  <el-card class="wrong-card">
    <div class="q-head">
      <span class="q-title">{{ question.question }}</span>
      <el-tag :type="question.mastered ? 'success' : 'warning'" size="small">
        {{ question.mastered ? '已掌握' : '未掌握' }}
      </el-tag>
    </div>
    <el-radio-group v-model="selected" :disabled="disabled">
      <el-radio v-for="(opt, i) in question.options" :key="i" :value="i" class="option">
        {{ opt }}
        <el-tag v-if="reveal && i === question.answer_index" type="success" size="small">正确答案</el-tag>
        <el-tag v-else-if="reveal && selected === i" type="danger" size="small">你的选择</el-tag>
        <el-tag v-else-if="!reveal && i === question.selected_index" type="danger" size="small">上次错选</el-tag>
      </el-radio>
    </el-radio-group>
    <div v-if="reveal" class="explanation">解析：{{ question.explanation }}</div>
    <div v-if="result === 'correct'" class="result-ok">🎉 回答正确，已记为掌握</div>
    <div v-else-if="result === 'wrong'" class="result-ng">回答错误，已更新你的选择，本题保持未掌握</div>

    <div class="card-actions">
      <template v-if="!question.mastered">
        <el-button v-if="!retrying && !result" type="primary" plain size="small" @click="startRetry">重做本题</el-button>
        <template v-else-if="result">
          <el-button v-if="result === 'wrong'" type="primary" plain size="small" @click="startRetry">再做一次</el-button>
          <el-button type="primary" size="small" @click="finish">完成</el-button>
        </template>
        <template v-else>
          <el-button type="primary" size="small" :loading="loading" @click="submitRetry">提交答案</el-button>
          <el-button size="small" @click="cancelRetry">取消</el-button>
        </template>
      </template>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { retryWrongQuestion } from '@/api/quiz'
import type { WrongQuestion } from '@/constants/quiz'

const props = defineProps<{ question: WrongQuestion }>()
const emit = defineEmits<{ (e: 'retried', payload: { questionId: number; correct: boolean }): void }>()

const retrying = ref(false)
const loading = ref(false)
const result = ref<'' | 'correct' | 'wrong'>('')
const selected = ref<number | undefined>(props.question.selected_index)

const reveal = computed(() => props.question.mastered || result.value !== '')
// Selection is only enabled while actively redoing an unmastered question.
const disabled = computed(() => props.question.mastered || !retrying.value || result.value !== '')

function startRetry() {
  retrying.value = true
  result.value = ''
  selected.value = undefined
}

function cancelRetry() {
  retrying.value = false
  result.value = ''
  selected.value = props.question.selected_index
}

async function submitRetry() {
  if (selected.value === undefined) {
    ElMessage.warning('请先选择一个答案')
    return
  }
  loading.value = true
  try {
    const res = await retryWrongQuestion(props.question.question_id, selected.value)
    result.value = res.correct ? 'correct' : 'wrong'
    retrying.value = false
    emit('retried', { questionId: props.question.question_id, correct: res.correct })
  } finally {
    loading.value = false
  }
}

function finish() {
  result.value = ''
  selected.value = props.question.selected_index
}
</script>

<style scoped>
.wrong-card { margin-bottom: 12px; }
.q-head { display: flex; justify-content: space-between; align-items: center; gap: 8px; margin-bottom: 8px; }
.q-title { font-weight: 700; }
.option { display: block; margin: 6px 0; }
.explanation { margin-top: 8px; color: #3c8d5c; }
.result-ok { margin-top: 8px; color: #3c8d5c; font-weight: 600; }
.result-ng { margin-top: 8px; color: #e6532f; font-weight: 600; }
.card-actions { margin-top: 10px; }
</style>
