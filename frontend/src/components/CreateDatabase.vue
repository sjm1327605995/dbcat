<template>
  <el-dialog
    v-model="dialogVisible"
    :title="'新建' + (props.config?.Type === 'mysql' ? 'MySQL' : 
                     props.config?.Type === 'postgres' ? 'PostgreSQL' : 
                     props.config?.Type === 'sqlite' ? 'SQLite' : '') + '数据库'"
    width="500px"
    :close-on-click-modal="false"
    class="create-database-dialog"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      size="default"
    >
      <el-form-item label="数据库名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入数据库名称" />
      </el-form-item>
      
      <template v-if="props.config?.Type !== 'sqlite'">
        <el-form-item label="字符集" prop="charset">
          <el-select v-model="form.charset" placeholder="请选择字符集" style="width: 100%">
            <el-option
              v-for="item in charsetOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="排序规则" prop="collation">
          <el-select 
            v-model="form.collation" 
            placeholder="请选择排序规则"
            style="width: 100%"
            :disabled="!form.charset"
          >
            <el-option
              v-for="item in collationOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </template>
    </el-form>
    
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="loading">
          创建
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import type { DatabaseConfig } from '../types/database'
import { CreateDatabase, GetDatabaseCharsets } from '../../wailsjs/go/main/App'

const props = defineProps<{
  visible: boolean
  config: DatabaseConfig
}>()

const emit = defineEmits(['update:visible', 'success'])

// 表单数据
const form = ref({
  name: '',
  charset: 'utf8mb4',
  collation: 'utf8mb4_general_ci'
})

// 字符集选项
const charsetOptions = ref<{ label: string; value: string }[]>([])
const collationOptions = ref<{ label: string; value: string }[]>([])

// 加载字符集和排序规则
const loadCharsets = async () => {
  if (!props.config || props.config.Type === 'sqlite') {
    // SQLite 不需要选择字符集和排序规则
    charsetOptions.value = []
    collationOptions.value = []
    return
  }
  
  try {
    const charsets = await GetDatabaseCharsets(props.config)
    charsetOptions.value = charsets.map(charset => ({
      label: `${charset.description} (${charset.name})`,
      value: charset.name
    }))

    // 如果有字符集，自动选择第一个字符集的第一个排序规则
    if (charsets.length > 0) {
      form.value.charset = charsets[0].name
      if (charsets[0].collations.length > 0) {
        form.value.collation = charsets[0].collations[0]
      }
    }
  } catch (error) {
    console.error('Failed to load charsets:', error)
    ElMessage.error('获取字符集失败: ' + error)
  }
}

// 监听字符集变化
watch(() => form.value.charset, async (newCharset) => {
  if (!newCharset || !props.config || props.config.Type === 'sqlite') return
  
  try {
    const charsets = await GetDatabaseCharsets(props.config)
    const currentCharset = charsets.find(c => c.name === newCharset)
    if (currentCharset) {
      collationOptions.value = currentCharset.collations.map(collation => ({
        label: collation,
        value: collation
      }))
      // 自动选择第一个排序规则
      if (currentCharset.collations.length > 0) {
        form.value.collation = currentCharset.collations[0]
      }
    }
  } catch (error) {
    console.error('Failed to load collations:', error)
    ElMessage.error('获取排序规则失败: ' + error)
  }
})

// 组件挂载时加载字符集
onMounted(() => {
  loadCharsets()
})

// 修改表单验证规则
const rules = computed(() => ({
  name: [
    { required: true, message: '请输入数据库名称', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_]*$/, message: '数据库名称只能包含字母、数字和下划线，且必须以字母开头', trigger: 'blur' }
  ],
  charset: [
    { 
      required: props.config?.Type !== 'sqlite', 
      message: '请选择字符集', 
      trigger: 'change' 
    }
  ],
  collation: [
    { 
      required: props.config?.Type !== 'sqlite', 
      message: '请选择排序规则', 
      trigger: 'change' 
    }
  ]
}))

const formRef = ref<FormInstance>()
const loading = ref(false)
const dialogVisible = ref(false)

// 监听visible属性变化
watch(() => props.visible, (newVal) => {
  dialogVisible.value = newVal
})

// 监听对话框关闭
watch(dialogVisible, (newVal) => {
  if (newVal !== props.visible) {
    emit('update:visible', newVal)
  }
  if (!newVal) {
    form.value = {
      name: '',
      charset: 'utf8mb4',
      collation: 'utf8mb4_general_ci'
    }
    if (formRef.value) {
      formRef.value.resetFields()
    }
  }
})

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    const valid = await formRef.value.validate()
    if (valid) {
      loading.value = true
      try {
        await CreateDatabase(props.config, {
          name: form.value.name,
          charset: form.value.charset,
          collation: form.value.collation
        })
        
        ElMessage.success('数据库创建成功')
        dialogVisible.value = false
        // 触发刷新事件
        emit('success')
      } catch (error) {
        console.error('Failed to create database:', error)
        ElMessage.error('创建数据库失败: ' + error)
      } finally {
        loading.value = false
      }
    }
  } catch (error) {
    console.error('Form validation failed:', error)
  }
}
</script>

<style scoped>
.create-database-dialog {
  :deep(.el-dialog__body) {
    padding: 20px;
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style> 