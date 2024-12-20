<template>
  <el-container class="app-container">
    <Sidebar @select-table="handleTableSelect" />
    <el-main v-if="selectedTable">
      <TableContent
        :config="selectedTable.config"
        :database="selectedTable.database"
        :table="selectedTable.table"
      />
    </el-main>
    <el-main v-else class="empty-state">
      <el-empty description="请选择一个表" />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Sidebar from './components/Sidebar.vue'
import TableContent from './components/TableContent.vue'
import type { DatabaseConfig } from './types/database'

interface SelectedTable {
  config: DatabaseConfig
  database: string
  table: string
}

const selectedTable = ref<SelectedTable | null>(null)

const handleTableSelect = (data: SelectedTable) => {
  console.log('Selected table:', data)
  selectedTable.value = {
    config: { ...data.config },
    database: data.database,
    table: data.table
  }
}
</script>

<style>
.app-container {
  height: 100vh;
  width: 100vw;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f5f7fa;
}

.el-main {
  padding: 0;
  background-color: #fff;
}
</style>