<template>
  <el-card class="quiz-card">
    <div class="q-title">{{ index !== undefined ? `${index + 1}. ` : '' }}{{ question.question }}</div>
    <el-radio-group v-model="selected" :disabled="submitted">
      <el-radio v-for="(opt, i) in question.options" :key="i" :value="i" class="option">
        {{ opt }}
        <el-tag v-if="submitted && showAnswer && i === answerIndex" type="success" size="small">正确答案</el-tag>
        <el-tag v-else-if="submitted && selected === i" :type="showAnswer ? 'danger' : 'info'" size="small">你的选择</el-tag>
      </el-radio>
    </el-radio-group>
    <div v-if="submitted && showAnswer" class="explanation">解析：{{ explanation }}</div>
  </el-card>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  question: { id: number; question: string; options: string[] }
  index?: number
  submitted: boolean
  modelValue: number | undefined
  answerIndex?: number
  explanation?: string
}>()
const emit = defineEmits<{ (e: 'update:modelValue', v: number): void }>()

const selected = ref<number | undefined>(props.modelValue)
// During grading the correct answer is only shown after submission.
const showAnswer = computed(() => props.submitted && props.answerIndex !== undefined)

watch(() => props.modelValue, (v) => { selected.value = v })
watch(selected, (v) => { if (v !== undefined) emit('update:modelValue', v) })
</script>

<style scoped>
.quiz-card { margin-bottom: 12px; }
.q-title { font-weight: 700; margin-bottom: 8px; }
.option { display: block; margin: 6px 0; }
.explanation { margin-top: 8px; color: #3c8d5c; }
</style>
