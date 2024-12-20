<template>
  <div class="toolbar">
    <el-button-group>
      <el-button @click="$emit('newGroup')" type="primary" plain>
        <el-icon><Folder /></el-icon>
        新建组
      </el-button>
      <el-button @click="$emit('newConnection')" type="primary" plain>
        <el-icon><Plus /></el-icon>
        新建连接
      </el-button>
    </el-button-group>
  </div>
</template>

<script setup lang="ts">
import { Folder, Plus } from '@element-plus/icons-vue'
import type { TreeNodeData } from '../types/tree'

defineProps<{
  treeData: TreeNodeData[]
}>()

defineEmits<{
  (e: 'update:treeData', value: TreeNodeData[]): void
  (e: 'newGroup'): void
  (e: 'newConnection'): void
}>()
</script>

<style scoped>
.toolbar {
  padding: 8px;
  border-bottom: 1px solid #dcdfe6;
  background-color: #fff;
}
</style> 