<template>
  <el-icon class="db-icon">
    <img v-if="type === 'mysql'" :src="mysqlIcon" :alt="type">
    <img v-else-if="type === 'postgres'" :src="postgresIcon" :alt="type">
    <img v-else-if="type === 'sqlite'" :src="sqliteIcon" :alt="type">
    <DataLine v-else />
  </el-icon>
</template>

<script setup lang="ts">
import { DataLine } from '@element-plus/icons-vue'
import mysqlIcon from '../assets/icons/mysql.svg?url'
import postgresIcon from '../assets/icons/postgres.svg?url'
import sqliteIcon from '../assets/icons/sqlite.svg?url'

defineProps<{
  type?: 'mysql' | 'postgres' | 'sqlite'
}>()
</script>

<style scoped>
.db-icon {
  display: flex;
  align-items: center;
  justify-content: center;
}

.db-icon img {
  width: 16px;
  height: 16px;
}
</style> 