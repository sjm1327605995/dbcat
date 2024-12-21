<template>
  <span class="custom-tree-node">
    <el-icon v-if="data.type === 'group'" class="folder-icon">
      <Folder />
    </el-icon>
    <DatabaseIcon
      v-else-if="data.type === 'connection' && isValidDbType(data.dbType)"
      :type="data.dbType"
    />
    <el-icon v-else-if="data.type === 'database'" class="database-icon">
      <Collection />
    </el-icon>
    <el-icon v-else-if="data.type === 'table'" class="table-icon">
      <Grid />
    </el-icon>
    <span class="node-label">{{ data.label }}</span>
    
    <!-- 连接状态和操作按钮 -->
    <span v-if="data.type === 'connection'" class="node-actions">
      <!-- 新建数据库按钮 -->
      <el-button
        v-if="data.isConnected && data.config"
        type="primary"
        size="small"
        link
        @click.stop="handleCreateDatabase"
      >
        <el-icon><Plus /></el-icon>
      </el-button>
      
      <!-- 连接状态图标 -->
      <el-icon class="status-icon" :class="{ 'is-connected': data.isConnected }">
        <Link />
      </el-icon>
    </span>
  </span>

  <!-- 创建数据库对话框 -->
  <CreateDatabase
    v-if="data.config"
    :visible="createDatabaseVisible"
    @update:visible="createDatabaseVisible = $event"
    :config="data.config"
    @success="handleDatabaseCreated"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Folder, Collection, Grid, Link, Plus } from '@element-plus/icons-vue'
import DatabaseIcon from './DatabaseIcon.vue'
import CreateDatabase from './CreateDatabase.vue'
import type { TreeNodeData } from '../types/tree'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  data: TreeNodeData
}>()

const emit = defineEmits(['refresh'])

const createDatabaseVisible = ref(false)

// 添加类型检查函数
const isValidDbType = (type?: string): type is 'mysql' | 'postgres' | 'sqlite' => {
  return type === 'mysql' || type === 'postgres' || type === 'sqlite'
}

// 处理创建数据库按钮点击
const handleCreateDatabase = (event: MouseEvent) => {
  if (!props.data.config) {
    ElMessage.warning('配置信息不存在')
    return
  }
  event.stopPropagation() // 阻止事件冒泡
  createDatabaseVisible.value = true
  console.log('Opening create database dialog', createDatabaseVisible.value)
}

// 处理数据库创建成功
const handleDatabaseCreated = () => {
  console.log('Database created successfully')
  createDatabaseVisible.value = false
  // 触发刷新事件，让父组件重新加载数据库列表
  emit('refresh')
}
</script>

<style scoped>
.custom-tree-node {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 24px;
  padding: 0 4px;
  width: 100%;
}

.node-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: auto;
  z-index: 1;
}

.status-icon {
  font-size: 14px;
  opacity: 0.5;
}

.status-icon.is-connected {
  color: #67C23A;
}

.folder-icon {
  color: #909399;
}

.database-icon {
  color: #409EFF;
}

.table-icon {
  color: #67C23A;
}

/* 鼠标悬停时显示新建按钮 */
.custom-tree-node:not(:hover) .node-actions :deep(.el-button) {
  display: none;
}

:deep(.el-button) {
  padding: 2px 4px;
  height: 20px;
}
</style> 