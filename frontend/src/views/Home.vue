<template>
  <div class="home">
    <div class="sidebar">
      <DatabaseList
        v-if="currentConfig"
        :config="currentConfig"
        @database-selected="handleDatabaseSelected"
        @refresh="loadDatabases"
      />
    </div>
    
    <div class="main-content">
      <TableContent
        v-if="selectedTable && currentConfig"
        :config="currentConfig"
        :database="selectedDatabase"
        :table="selectedTable"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import DatabaseList from '../components/DatabaseList.vue'
import TableContent from '../components/TableContent.vue'
import type { DatabaseConfig } from '../types/database'

const currentConfig = ref<DatabaseConfig>()
const selectedDatabase = ref('')
const selectedTable = ref('')

// 处理数据库选择
const handleDatabaseSelected = (data: any) => {
  selectedDatabase.value = data.name
  // 其他处理逻辑...
}

// 加载数据库列表
const loadDatabases = async () => {
  // 实现加载数据库列表的逻辑...
}
</script>

<style scoped>
.home {
  display: flex;
  height: 100vh;
}

.sidebar {
  width: 250px;
  border-right: 1px solid #d0d7de;
  background-color: #ffffff;
}

.main-content {
  flex: 1;
  overflow: hidden;
}
</style> 