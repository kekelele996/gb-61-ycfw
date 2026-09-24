<template>
  <el-card class="quiz-card">
    <div class="q-title">
      {{ index + 1 }}. {{ question.question }}
      <el-tag v-if="showWrongTag" type="danger" size="small" class="prev-tag">上次误选</el-tag>
      <el-tag v-if="mastered" type="success" size="small" class="prev-tag">已掌握</el-tag>
    </div>
    <el-radio-group v-model="selected" :disabled="disabled">
      <el-radio v-for="opt in question.options" :key="opt.index" :value="opt.index" class="option">
        {{ opt.text }}
        <el-tag v-if="submitted && opt.index === question.answer" type="success" size="small">正确答案</el-tag>
        <el-tag v-else-if="submitted && selected === opt.index" type="danger" size="small">你的选择</el-tag>
      </el-radio>
    </el-radio-group>
    <div v-if="submitted" class="explanation">解析：{{ question.explanation }}</div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { QuizQuestion } from '@/types/api'

const props = withDefaults(
  defineProps<{
    question: QuizQuestion
    index: number
    submitted: boolean
    modelValue?: number
    disabled?: boolean
    showWrongTag?: boolean
    mastered?: boolean
  }>(),
  { modelValue: undefined, disabled: false, showWrongTag: false, mastered: false },
)
const emit = defineEmits<{ (e: 'update:modelValue', v: number): void }>()
const selected = ref<number | undefined>(props.modelValue)
watch(() => props.modelValue, (v) => { selected.value = v })
watch(selected, (v) => { if (v !== undefined) emit('update:modelValue', v) })
</script>

<style scoped>
.quiz-card { margin-bottom: 12px; }
.q-title { font-weight: 700; margin-bottom: 8px; }
.prev-tag { margin-left: 8px; }
.option { display: block; margin: 6px 0; }
.explanation { margin-top: 8px; color: #3c8d5c; }
</style>
