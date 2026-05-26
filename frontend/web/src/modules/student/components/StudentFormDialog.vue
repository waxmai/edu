<template>
  <el-dialog :model-value="modelValue" title="学员信息" width="640px" @close="$emit('update:modelValue', false)">
    <el-form :model="formModel" label-width="96px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="学员姓名" required>
            <el-input v-model="formModel.studentName" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="性别">
            <el-select v-model="formModel.gender" style="width: 100%">
              <el-option label="男" value="male" />
              <el-option label="女" value="female" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="年级" required>
            <el-input v-model="formModel.grade" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="手机号">
            <el-input v-model="formModel.phone" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="家长姓名" required>
            <el-input v-model="formModel.parentName" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="家长电话" required>
            <el-input v-model="formModel.parentPhone" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="报读科目" required>
            <el-input v-model="formModel.subject" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="授课形式">
            <el-select v-model="formModel.teachingType" style="width: 100%">
              <el-option
                v-for="option in teachingTypeOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="状态">
            <el-select v-model="formModel.status" style="width: 100%">
              <el-option
                v-for="option in studentStatusOptions.slice(1)"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="备注">
            <el-input v-model="formModel.remark" type="textarea" :rows="3" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" @click="$emit('submit', formModel)">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'
import { studentStatusOptions, teachingTypeOptions } from '@/modules/student/constants/student'
import type { StudentFormModel } from '@/modules/student/types/student'

const props = defineProps<{
  modelValue: boolean
  initialValue: StudentFormModel
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [value: StudentFormModel]
}>()

const formModel = reactive<StudentFormModel>({ ...props.initialValue })

watch(
  () => props.initialValue,
  (value) => {
    Object.assign(formModel, value)
  },
  { deep: true, immediate: true },
)
</script>
