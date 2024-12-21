<template>
  <div class="database-list">
    <div class="database-header">
      <span class="title">数据库列表</span>
    </div>

    <el-tree
      :data="databases"
      :props="defaultProps"
      @node-click="handleNodeClick"
    >
      <template #default="{ node, data }">
        <TreeNodeContent 
          :data="data" 
          @refresh="handleRefresh"
        />
      </template>
    </el-tree>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import TreeNodeContent from './TreeNodeContent.vue'
import type { DatabaseConfig } from '../types/database'

const props = defineProps<{
  config: DatabaseConfig
}>()

const emit = defineEmits(['database-selected', 'refresh'])

const databases = ref([])

// 树形控件配置
const defaultProps = {
  children: 'children',
  label: 'name'
}

// 处理节点点击
const handleNodeClick = (data: any) => {
  emit('database-selected', data)
}

// 处理刷新
const handleRefresh = () => {
  emit('refresh')
}
</script>

<style scoped>
.database-list {
  padding: 16px;
}

.database-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.title {
  font-size: 14px;
  font-weight: 600;
  color: #24292f;
}

.custom-tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  font-size: 12px;
  padding-right: 8px;
}
</style> 