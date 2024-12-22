<template>
  <div class="query-editor">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="connection-info">
        <el-select 
          v-model="selectedConnection" 
          placeholder="选择连接"
          size="small"
          @change="handleConnectionChange"
        >
          <el-option
            v-for="conn in connections"
            :key="conn.id"
            :label="conn.label"
            :value="conn.id"
          >
            <div class="connection-option">
              <DatabaseIcon :type="conn.dbType" />
              <span>{{ conn.label }}</span>
            </div>
          </el-option>
        </el-select>

        <el-select 
          v-model="selectedDatabase" 
          placeholder="选择数据库"
          size="small"
          :disabled="!selectedConnection || connecting"
          @change="handleDatabaseChange"
        >
          <el-option
            v-for="db in databases"
            :key="db.Name"
            :label="db.Name"
            :value="db.Name"
          />
        </el-select>
      </div>

      <el-button 
        type="primary" 
        @click="executeQuery" 
        :loading="loading"
        :disabled="!canExecuteQuery"
        size="small"
      >
        {{ getExecuteButtonText }}
      </el-button>
    </div>

    <!-- SQL 编辑器 -->
    <div class="editor-container">
      <el-input
        v-model="sql"
        type="textarea"
        :rows="8"
        placeholder="请输入 SQL 语句..."
        class="sql-input"
      />
    </div>
    
    <!-- 查询结果 -->
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
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import type { DatabaseConfig } from '../types/database'
import type { TreeNodeData } from '../types/tree'
import { ExecuteQuery, GetDatabases, TestConnection } from '../../wailsjs/go/main/App'
import DatabaseIcon from './DatabaseIcon.vue'
import { StorageManager } from '../utils/storage'

const props = defineProps<{
  config: DatabaseConfig
  database: string
}>()

const sql = ref('')
const loading = ref(false)
const results = ref<any[]>([])
const columns = ref<string[]>([])

// 连接和数据库选择
const connections = ref<TreeNodeData[]>([])
const selectedConnection = ref('')
const databases = ref<{ Name: string }[]>([])
const selectedDatabase = ref('')

// 添加状态变量
const connecting = ref(false)
const connected = ref(false)

// 计算属性：是否可以执行查询
const canExecuteQuery = computed(() => {
  return selectedConnection.value && 
         selectedDatabase.value && 
         !connecting.value &&
         connected.value
})

// 计算属性：执行按钮文本
const getExecuteButtonText = computed(() => {
  if (connecting.value) return '连接中...'
  if (!selectedConnection.value) return '请选择连接'
  if (!selectedDatabase.value) return '请选择数据库'
  if (!connected.value) return '未连接'
  if (loading.value) return '执行中...'
  return '执行查询'
})

// 加载所有连接
onMounted(() => {
  const storageManager = StorageManager.getInstance()
  const treeData = storageManager.loadFromStorage()
  
  // 提取所有连接节点
  const extractConnections = (nodes: TreeNodeData[]) => {
    let result: TreeNodeData[] = []
    for (const node of nodes) {
      if (node.type === 'connection' && node.config) {
        result.push(node)
      }
      if (node.children) {
        result = result.concat(extractConnections(node.children))
      }
    }
    return result
  }
  
  connections.value = extractConnections(treeData.groups)
  
  // 设置当前连接和数据库
  if (props.config) {
    const currentConn = connections.value.find(
      conn => conn.config?.Host === props.config.Host && 
              conn.config?.Port === props.config.Port &&
              conn.config?.Database === props.config.Database
    )
    if (currentConn && currentConn.config) {
      selectedConnection.value = currentConn.id
      selectedDatabase.value = props.database
      loadDatabases(currentConn.config)
    }
  }
})

// 加载数据库列表
const loadDatabases = async (config: DatabaseConfig) => {
  try {
    const dbs = await GetDatabases(config)
    databases.value = dbs || []
  } catch (error: any) {
    // 显示详细的错误信���
    const errorMessage = error.message || error.toString()
    ElMessage.error({
      message: `获取数据库列表失败: ${errorMessage}`,
      duration: 5000,  // 显示时间延长到5秒
      showClose: true  // 显示关闭按钮
    })
    databases.value = []
  }
}

// 处理连接变更
const handleConnectionChange = async (connId: string) => {
  const conn = connections.value.find(c => c.id === connId)
  if (conn?.config) {
    try {
      connecting.value = true
      connected.value = false
      selectedDatabase.value = ''
      databases.value = []
      
      // 测试连接
      await TestConnection(conn.config)
      await loadDatabases(conn.config)
      connected.value = true
    } catch (error: any) {
      const errorMessage = error.message || error.toString()
      ElMessage.error({
        message: `连接失败: ${errorMessage}`,
        duration: 5000,
        showClose: true
      })
      selectedConnection.value = ''
    } finally {
      connecting.value = false
    }
  }
}

// 处理数据库变更
const handleDatabaseChange = async (dbName: string) => {
  if (!dbName) return
  
  const conn = connections.value.find(c => c.id === selectedConnection.value)
  if (!conn?.config) return

  try {
    connecting.value = true
    connected.value = false
    
    // 测试数据库连接
    await TestConnection({
      ...conn.config,
      Database: dbName
    })
    
    connected.value = true
  } catch (error: any) {
    const errorMessage = error.message || error.toString()
    ElMessage.error({
      message: `切换数据库失败: ${errorMessage}`,
      duration: 5000,
      showClose: true
    })
    selectedDatabase.value = ''
  } finally {
    connecting.value = false
  }
}

// 执行查询
const executeQuery = async () => {
  if (!sql.value.trim()) {
    ElMessage.warning('请输入 SQL 语句')
    return
  }

  const conn = connections.value.find(c => c.id === selectedConnection.value)
  if (!conn?.config || !selectedDatabase.value) {
    ElMessage.warning('请选择连接和数据库')
    return
  }

  loading.value = true
  try {
    const data = await ExecuteQuery(conn.config, selectedDatabase.value, sql.value)
    if (data && data.length > 0) {
      results.value = data
      columns.value = Object.keys(data[0])
      ElMessage.success('查询执行成功')
    } else {
      results.value = []
      columns.value = []
      ElMessage.info('查询执行成功，但没有返回数据')
    }
  } catch (error: any) {
    console.error('Query failed:', error)
    // 显示详细的错误信息
    const errorMessage = error.message || error.toString()
    ElMessage.error({
      dangerouslyUseHTMLString: true,
      message: `
        <div style="text-align: left;">
          <div>SQL执行失败:</div>
          <div style="color: #F56C6C; margin-top: 5px; word-break: break-all;">
            ${errorMessage}
          </div>
        </div>
      `,
      duration: 0,  // 不自动关闭
      showClose: true
    })
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

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.connection-info {
  display: flex;
  gap: 8px;
  flex: 1;
  position: relative;
}

.connection-info :deep(.el-select) {
  width: 200px;
}

.connection-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.editor-container {
  flex: 0 0 auto;
}

.sql-input {
  font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace;
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

.el-button:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

/* 添加连接状态指示器 */
.connection-info::after {
  content: '';
  position: absolute;
  right: -20px;
  top: 50%;
  transform: translateY(-50%);
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: v-bind(connected ? '#67C23A' : '#F56C6C');
  transition: background-color 0.3s;
}
</style> 