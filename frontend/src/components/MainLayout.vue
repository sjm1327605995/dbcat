<template>
  <div class="main-layout">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="tool-group">
        <el-tooltip content="新建连接" placement="bottom">
          <div class="tool-item" @click="handleNewConnection">
            <el-icon><Connection /></el-icon>
            <span class="tool-label">连接</span>
          </div>
        </el-tooltip>
        <el-tooltip content="新建查询" placement="bottom">
          <div 
            class="tool-item" 
            :class="{ disabled: !canQuery }"
            @click="handleToolbarNewQuery"
          >
            <el-icon><Document /></el-icon>
            <span class="tool-label">查询</span>
          </div>
        </el-tooltip>
      </div>

      <div class="tool-group">
        <el-tooltip content="刷新" placement="bottom">
          <div class="tool-item" @click="handleRefresh">
            <el-icon><Refresh /></el-icon>
            <span class="tool-label">刷新</span>
          </div>
        </el-tooltip>
        <el-tooltip content="设置" placement="bottom">
          <div class="tool-item">
            <el-icon><Setting /></el-icon>
            <span class="tool-label">设置</span>
          </div>
        </el-tooltip>
      </div>
    </div>

    <!-- 主体内容区 -->
    <div class="content">
      <!-- 左侧边栏 -->
      <Sidebar 
        ref="sidebarRef"
        class="sidebar" 
        @select-table="handleSelectTable"
        @new-query="handleNewQuery"
        @connection-select="handleConnectionSelect"
      />

      <!-- 右侧内容区 -->
      <div class="main">
        <QueryEditor
          v-if="activeTab === 'query' && currentQuery"
          :config="currentQuery.config"
          :database="currentQuery.database"
        />
        <TableContent
          v-else-if="selectedTable"
          :config="selectedTable.config"
          :database="selectedTable.database"
          :table="selectedTable.table"
        />
        <div v-else class="welcome">
          <el-empty description="选择表或创建新查询" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Connection, Document, Refresh, Setting } from '@element-plus/icons-vue'
import Sidebar from './Sidebar.vue'
import TableContent from './TableContent.vue'
import QueryEditor from './QueryEditor.vue'
import type { DatabaseConfig } from '../types/database'
import type { TreeNodeData } from '../types/tree'

// 定义事件
const emit = defineEmits<{
  (e: 'select-table', data: { config: DatabaseConfig; database: string; table: string }): void
  (e: 'new-query', data: { config: DatabaseConfig; database: string }): void
}>()

const activeTab = ref<'table' | 'query'>('table')
const selectedTable = ref<{
  config: DatabaseConfig
  database: string
  table: string
} | null>(null)

const currentQuery = ref<{
  config: DatabaseConfig
  database: string
} | null>(null)

// 添加选中的连接节点引用
const selectedConnection = ref<TreeNodeData | null>(null)

// 修改 canQuery 的计算逻辑
const canQuery = computed(() => selectedConnection.value?.config !== undefined)

// 处理表格选择
const handleSelectTable = (data: { config: DatabaseConfig; database: string; table: string }) => {
  selectedTable.value = data
  activeTab.value = 'table'
}

// 处理新建查询
const handleNewQuery = (data: { config: DatabaseConfig; database: string }) => {
  currentQuery.value = data
  activeTab.value = 'query'
}

// 工具栏的新建查询处理函数
const handleToolbarNewQuery = () => {
  if (selectedConnection.value?.config) {
    handleNewQuery({
      config: selectedConnection.value.config,
      database: selectedConnection.value.label
    })
  }
}

// 处理新建连接
const handleNewConnection = () => {
  // 调用 Sidebar 组件的方法
  if (sidebarRef.value) {
    sidebarRef.value.handleContextMenuCommand('newConnection')
  }
}

// 处理刷新
const handleRefresh = () => {
  // 根据当前活动页面刷新数据
}

// 处理连接选择
const handleConnectionSelect = (connection: TreeNodeData) => {
  selectedConnection.value = connection
}

// 添加 Sidebar 组件引用
const sidebarRef = ref<InstanceType<typeof Sidebar> | null>(null)
</script>

<style scoped>
.main-layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f5f7fa;
}

.toolbar {
  height: 64px;
  background-color: #ffffff;
  border-bottom: 1px solid #dcdfe6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
}

.tool-group {
  display: flex;
  gap: 16px;
}

.tool-item {
  width: 48px;
  height: 48px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  cursor: pointer;
  color: #606266;
  transition: all 0.3s;
  gap: 4px;
}

.tool-item:hover {
  background-color: #f0f2f5;
  color: #409EFF;
}

.tool-item.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.tool-item .el-icon {
  font-size: 20px;
}

.tool-label {
  font-size: 12px;
  line-height: 1;
}

.content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.sidebar {
  width: 250px;
  border-right: 1px solid #dcdfe6;
}

.main {
  flex: 1;
  overflow: hidden;
  background-color: #ffffff;
}

.welcome {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style> 