<template>
  <div class="host-manager">
    <el-page-header content="主机管理" />
    
    <el-card class="manager-card">
      <template #header>
        <div class="card-header">
          <span>添加新主机</span>
        </div>
      </template>
      
      <el-form 
        ref="formRef"
        :model="hostForm"
        :rules="rules"
        label-width="120px"
        inline
      >
        <el-form-item label="主机IP地址" prop="ip_address">
          <el-input v-model="hostForm.ip_address" placeholder="输入主机IP地址" />
        </el-form-item>
        
        <el-form-item label="标签" prop="label">
          <el-input v-model="hostForm.label" placeholder="输入主机标签" />
        </el-form-item>
        
        <el-form-item label="负责人" prop="owner">
          <el-input v-model="hostForm.owner" placeholder="输入负责人" />
        </el-form-item>
        
        <el-form-item label="开启预警">
          <el-switch v-model="hostForm.alert_enabled" />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="addHost">添加主机</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="manager-card">
      <template #header>
        <div class="card-header">
          <span>主机列表</span>
        </div>
      </template>

      <el-table 
        :data="sortedHosts" 
        style="width: 100%"
        border
      >
        <el-table-column prop="ip_address" label="IP地址" width="200" />
        <el-table-column label="主机标签" width="200">
          <template #default="{ row }">
            <div class="editable-cell">
              <el-input
                v-if="row.editing"
                v-model="row.tempLabel"
                size="small"
                @blur="updateHostInfo(row)"
                @keyup.enter="updateHostInfo(row)"
              />
              <div v-else @click="startEdit(row, 'label')" class="editable-text">
                {{ row.label }}
                <el-icon class="edit-icon"><Edit /></el-icon>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="负责人" width="200">
          <template #default="{ row }">
            <div class="editable-cell">
              <el-input
                v-if="row.editingOwner"
                v-model="row.tempOwner"
                size="small"
                @blur="updateHostInfo(row)"
                @keyup.enter="updateHostInfo(row)"
              />
              <div v-else @click="startEdit(row, 'owner')" class="editable-text">
                {{ row.owner }}
                <el-icon class="edit-icon"><Edit /></el-icon>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="是否开启离线预警" width="120">
          <template #default="{ row }">
            <el-switch
              v-model="row.alert_enabled"
              @change="(val) => toggleAlert(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button 
              type="danger" 
              size="small"
              @click="deleteHost(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '../api'
import { Edit } from '@element-plus/icons-vue'

// 表单数据
const formRef = ref(null)
const hostForm = ref({
  ip_address: '',
  label: '',
  owner: '',
  alert_enabled: true
})

// 主机列表
const hosts = ref([])

// 表单验证规则
const rules = {
  ip_address: [
    { required: true, message: 'IP地址不能为空', trigger: 'blur' },
    { pattern: /^(\d{1,3}\.){3}\d{1,3}$/, message: 'IP地址格式不正确', trigger: 'blur' }
  ],
  label: [
    { required: true, message: '标签不能为空', trigger: 'blur' }
  ],
  owner: [
    { required: true, message: '负责人不能为空', trigger: 'blur' }
  ]
}

// 排序后的主机列表
const sortedHosts = computed(() => {
  return [...hosts.value].sort((a, b) => {
    // 先按标签排序
    const labelA = a.label || ''  // 处理空标签
    const labelB = b.label || ''
    
    if (labelA !== labelB) {
      return labelA.localeCompare(labelB)
    }
    
    // 标签相同时按 IP 排序
    const aIP = a.ip_address.split('.').map(Number)
    const bIP = b.ip_address.split('.').map(Number)
    
    for (let i = 0; i < 4; i++) {
      if (aIP[i] !== bIP[i]) {
        return aIP[i] - bIP[i]
      }
    }
    return 0
  })
})

// 添加主机
async function addHost() {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const data = await api.addHost(hostForm.value)
        
        if (data.success) {
          ElMessage.success('主机添加成功！')
          loadHosts()
          formRef.value.resetFields()
        } else {
          ElMessage.error(`添加失败: ${data.message}`)
        }
      } catch (error) {
        console.error('请求失败:', error)
        ElMessage.error('添加过程中出现错误')
      }
    }
  })
}

// 删除主机
async function deleteHost(host) {
  try {
    await ElMessageBox.confirm(
      `确定要删除主机 ${host.ip_address} 吗？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    const data = await api.deleteHost(`${host.ip_address}`)
    
    if (data.success) {
      ElMessage.success('主机删除成功！')
      loadHosts()
    } else {
      ElMessage.error(`删除失败: ${data.message}`)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('请求失败:', error)
      ElMessage.error('删除过程中出现错误')
    }
  }
}

// 切换预警状态
async function toggleAlert(host, value) {
  try {
    const data = await api.updateHost({
        ip_address: host.ip_address,
        label: host.label,
        owner: host.owner,
        alert_enabled: value
      })

    console.log(data)
    
    if (data.success) {
      ElMessage.success('预警状态更新成功！')
      loadHosts()
    } else {
      ElMessage.error(`更新失败: ${data.message}`)
      host.alert_enabled = !value // 恢复原状态
    }
  } catch (error) {
    console.error('请求失败:', error)
    ElMessage.error('更新过程中出现错误')
    host.alert_enabled = !value // 恢复原状态
  }
}

// 加载主机列表
async function loadHosts() {
  try {
    const data = await api.getHosts()
    
    if (data != null) {
      hosts.value = data
    }
  } catch (error) {
    console.error('请求失败:', error)
    ElMessage.error('加载主机信息时出现错误')
  }
}

// 初始加载
onMounted(() => {
  loadHosts()
})

// 开始编辑
function startEdit(row, field) {
  if (field === 'label') {
    row.editing = true
    row.tempLabel = row.label
  } else if (field === 'owner') {
    row.editingOwner = true
    row.tempOwner = row.owner
  }
}

// 更新主机信息
async function updateHostInfo(row) {
  // 如果没有在编辑状态，直接返回
  if (!row.editing && !row.editingOwner) return

  const newData = {
    ip_address: row.ip_address,
    label: row.editing ? row.tempLabel : row.label,
    owner: row.editingOwner ? row.tempOwner : row.owner,
    alert_enabled: row.alert_enabled
  }

  try {
    const data = await api.updateHost(newData)
    
    if (data.success) {
      ElMessage.success('更新成功！')
      // 更新成功后更新本地数据
      if (row.editing) row.label = row.tempLabel
      if (row.editingOwner) row.owner = row.tempOwner
      loadHosts() // 重新加载数据
    } else {
      ElMessage.error(`更新失败: ${data.message}`)
    }
  } catch (error) {
    console.error('请求失败:', error)
    ElMessage.error('更新过程中出现错误')
  } finally {
    // 无论成功失败都关闭编辑状态
    row.editing = false
    row.editingOwner = false
  }
}
</script>

<style scoped>
.host-manager {
  padding: 24px;
  background-color: #f9fafb;
  min-height: 100vh;
}

.manager-card {
  margin-top: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
  background-color: #ffffff;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
  font-size: 16px;
  color: #1f2937;
}

:deep(.el-form--inline .el-form-item) {
  margin-right: 32px;
  margin-bottom: 16px;
}

:deep(.el-input__inner) {
  border-radius: 6px;
}

:deep(.el-button) {
  border-radius: 6px;
  font-weight: 500;
}

:deep(.el-button--primary) {
  background-color: #60a5fa;
  border-color: #60a5fa;
}

:deep(.el-button--primary:hover) {
  background-color: #3b82f6;
  border-color: #3b82f6;
}

:deep(.el-table) {
  border-radius: 8px;
}

:deep(.el-table th) {
  background-color: #f9fafb;
}

.editable-cell {
  position: relative;
  padding: 5px 12px;
  cursor: pointer;
}

.editable-text {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.editable-text:hover .edit-icon {
  opacity: 1;
}

.edit-icon {
  opacity: 0;
  transition: opacity 0.3s;
  color: #409EFF;
  font-size: 14px;
}

:deep(.el-input__inner) {
  height: 32px;
}

:deep(.el-input__wrapper) {
  padding: 0 8px;
}

.editable-cell :deep(.el-input) {
  margin: -6px -13px;
}
</style> 