<template>
  <div class="query-editor">
    <div class="editor-container">
      <el-input
        v-model="sql"
        type="textarea"
        :rows="8"
        placeholder="请输入 SQL 语句..."
        class="sql-input"
      />
      <div class="toolbar">
        <el-button type="primary" @click="executeQuery" :loading="loading">
          执行查询
        </el-button>
      </div>
    </div>
    
    <div class="result-container">
      <el-table
        v-if="results.length > 0"
        :data="results"
        border
        style="width: 100%"
        height="calc(100% - 40px)"
      >
        <el-table-column
          v-for="col in columns"
          :key="col"
          :prop="col"
          :label="col"
        />
      </el-table>
      <div v-else class="no-data">
        {{ loading ? '查询中...' : '暂无数据' }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { DatabaseConfig } from '../types/database'
import { ExecuteQuery } from '../../wailsjs/go/main/App'

const props = defineProps<{
  config: DatabaseConfig
  database: string
}>()

const sql = ref('')
const loading = ref(false)
const results = ref<any[]>([])
const columns = ref<string[]>([])

const executeQuery = async () => {
  if (!sql.value.trim()) {
    ElMessage.warning('请输入 SQL 语句')
    return
  }

  loading.value = true
  try {
    const data = await ExecuteQuery(props.config, props.database, sql.value)
    if (data && data.length > 0) {
      results.value = data
      columns.value = Object.keys(data[0])
    } else {
      results.value = []
      columns.value = []
      ElMessage.success('查询执行成功，但没有返回数据')
    }
  } catch (error) {
    console.error('Query failed:', error)
    ElMessage.error('查询失败: ' + error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.query-editor {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
}

.editor-container {
  flex: 0 0 auto;
}

.sql-input {
  margin-bottom: 8px;
  font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace;
}

.toolbar {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.result-container {
  flex: 1;
  min-height: 0;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.no-data {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}
</style> 