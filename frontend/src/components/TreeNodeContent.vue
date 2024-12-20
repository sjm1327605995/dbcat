<template>
  <span class="custom-tree-node">
    <el-icon v-if="data.type === 'group'" class="folder-icon">
      <Folder />
    </el-icon>
    <DatabaseIcon
      v-else-if="data.type === 'connection'"
      :type="data.dbType"
    />
    <el-icon v-else-if="data.type === 'database'" class="database-icon">
      <Collection />
    </el-icon>
    <el-icon v-else-if="data.type === 'table'" class="table-icon">
      <Grid />
    </el-icon>
    <span class="node-label">{{ data.label }}</span>
    <span v-if="data.type === 'connection'" class="connection-status">
      <el-icon class="status-icon" :class="{ 'is-connected': data.isConnected }">
        <Link />
      </el-icon>
    </span>
  </span>
</template>

<script setup lang="ts">
import { Folder, Collection, Grid, Link } from '@element-plus/icons-vue'
import DatabaseIcon from './DatabaseIcon.vue'
import type { TreeNodeData } from '../types/tree'

defineProps<{
  data: TreeNodeData
}>()
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

.connection-status {
  margin-left: auto;
  opacity: 0.5;
}

.status-icon {
  font-size: 14px;
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
</style> 