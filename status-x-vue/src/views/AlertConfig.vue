<template>
  <div class="alert-config">
    <el-page-header content="警报配置" />
    
    <el-card class="config-card">
      <el-form 
        ref="formRef"
        :model="configForm"
        :rules="rules"
        label-width="180px"
        label-position="left"
      >
        <el-form-item label="CPU 使用率阈值 (%)" prop="cpu_threshold">
          <el-input-number 
            v-model="configForm.cpu_threshold"
            :min="0"
            :max="100"
            :step="1"
            :precision="0"
          />
        </el-form-item>

        <el-form-item label="CPU 占用持续时间 (分钟)" prop="cpu_duration">
          <el-input-number 
            v-model="configForm.cpu_duration"
            :min="0"
            :precision="0"
            :step="1"
          />
        </el-form-item>

        <el-form-item label="内存使用率阈值 (%)" prop="memory_threshold">
          <el-input-number 
            v-model="configForm.memory_threshold"
            :min="0"
            :max="100"
            :precision="0"
            :step="1"
          />
        </el-form-item>

        <el-form-item label="内存占用持续时间 (分钟)" prop="memory_duration">
          <el-input-number 
            v-model="configForm.memory_duration"
            :min="0"
            :precision="0"
            :step="1"
          />
        </el-form-item>

        <el-form-item label="磁盘使用率阈值 (%)" prop="disk_threshold">
          <el-input-number 
            v-model="configForm.disk_threshold"
            :min="0"
            :max="100"
            :precision="0"
            :step="1"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="submitForm">更新配置</el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 添加密码验证对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="管理员验证"
      width="30%"
      :close-on-click-modal="false"
    >
      <el-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-width="100px"
      >
        <el-form-item label="管理密码" prop="password">
          <el-input
            v-model="passwordForm.password"
            type="password"
            placeholder="请输入管理密码"
            show-password
            @keyup.enter="confirmPassword"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmPassword">确认</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useMetricsStore } from '../stores/metrics'
import { api } from '../api'

const store = useMetricsStore()
const formRef = ref(null)

// 表单数据
const configForm = ref({
  cpu_threshold: 90,
  cpu_duration: 5,
  memory_threshold: 90,
  memory_duration: 5,
  disk_threshold: 85
})

// 表单验证规则
const rules = {
  cpu_threshold: [
    { required: true, message: '请输入CPU阈值', trigger: 'blur' },
    { type: 'number', min: 0, max: 100, message: '阈值必须在0-100之间', trigger: 'blur' }
  ],
  cpu_duration: [
    { required: true, message: '请输入持续时间', trigger: 'blur' },
    { type: 'number', min: 0, message: '持续时间必须大于0', trigger: 'blur' }
  ],
  memory_threshold: [
    { required: true, message: '请输入内存阈值', trigger: 'blur' },
    { type: 'number', min: 0, max: 100, message: '阈值必须在0-100之间', trigger: 'blur' }
  ],
  memory_duration: [
    { required: true, message: '请输入持续时间', trigger: 'blur' },
    { type: 'number', min: 0, message: '持续时间必须大于0', trigger: 'blur' }
  ],
  disk_threshold: [
    { required: true, message: '请输入磁盘阈值', trigger: 'blur' },
    { type: 'number', min: 0, max: 100, message: '阈值必须在0-100之间', trigger: 'blur' }
  ]
}

// 添加密码验证相关的响应式变量
const dialogVisible = ref(false)
const passwordFormRef = ref(null)
const passwordForm = ref({
  password: ''
})

// 密码验证规则
const passwordRules = {
  password: [
    { required: true, message: '请输入管理密码', trigger: 'blur' }
  ]
}

// 提交表单
async function submitForm() {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      // 显示密码验证对话框
      dialogVisible.value = true
    }
  })
}

// 添加密码验证函数
async function confirmPassword() {
  if (!passwordFormRef.value) return

  await passwordFormRef.value.validate(async (valid) => {
    if (valid) {
      if (passwordForm.value.password === 'XR') {
        // 密码正确，执行更新操作
        try {
          const data = await api.updateAlertConfig(configForm.value)
          
          if (data.success) {
            ElMessage.success('警报配置更新成功！')
            dialogVisible.value = false
            passwordForm.value.password = '' // 清空密码
          } else {
            ElMessage.error(`更新失败: ${data.message}`)
          }
        } catch (error) {
          console.error('请求失败:', error)
          ElMessage.error('更新过程中出现错误')
        }
      } else {
        ElMessage.error('管理密码错误')
      }
    }
  })
}

// 重置表单
function resetForm() {
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

// 加载初始配置
onMounted(async () => {
  try {
    const data = await api.getAlertConfig()
    
    if (data.success) {
      configForm.value = {
        cpu_threshold: data.cpu_threshold,
        cpu_duration: data.cpu_duration,
        memory_threshold: data.memory_threshold,
        memory_duration: data.memory_duration,
        disk_threshold: data.disk_threshold
      }
    } else {
      ElMessage.warning('获取配置失败: ' + data.message)
    }
  } catch (error) {
    console.error('请求失败:', error)
    ElMessage.error('获取配置过程中出现错误')
  }
})
</script>

<style scoped>
.alert-config {
  padding: 24px;
  background-color: #f9fafb;
  min-height: 100vh;
}

.config-card {
  margin-top: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
  background-color: #ffffff;
  max-width: 800px;
}

:deep(.el-form-item__label) {
  color: #4b5563;
  font-weight: 500;
}

:deep(.el-input-number) {
  width: 180px;
}

:deep(.el-input-number__decrease),
:deep(.el-input-number__increase) {
  background-color: #f9fafb;
  border-color: #e5e7eb;
}

:deep(.el-button) {
  border-radius: 6px;
  font-weight: 500;
  padding: 10px 24px;
}

:deep(.el-button--primary) {
  background-color: #60a5fa;
  border-color: #60a5fa;
}

:deep(.el-button--primary:hover) {
  background-color: #3b82f6;
  border-color: #3b82f6;
}

/* 添加对话框相关样式 */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

:deep(.el-dialog) {
  border-radius: 8px;
}

:deep(.el-dialog__header) {
  margin-right: 0;
  padding: 20px 20px 10px;
  border-bottom: 1px solid #f0f0f0;
}

:deep(.el-dialog__body) {
  padding: 30px 20px;
}

:deep(.el-dialog__footer) {
  padding: 10px 20px 20px;
  border-top: 1px solid #f0f0f0;
}

:deep(.el-input__wrapper) {
  border-radius: 6px;
}
</style> 