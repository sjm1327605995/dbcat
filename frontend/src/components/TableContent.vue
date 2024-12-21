<template>
  <div class="table-content">
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
      height="calc(100vh - 120px)"
      :header-cell-style="{
        background: '#f6f8fa',
        borderColor: '#d0d7de',
        color: '#24292f',
        fontWeight: '600',
        fontSize: '12px'
      }"
      :cell-style="{
        borderColor: '#d0d7de',
        fontSize: '12px',
        padding: '4px 8px',
        color: '#24292f'
      }"
      border
      size="small"
    >
      <el-table-column
        v-for="column in columns"
        :key="column.Name"
        :prop="column.Name"
        :label="column.Name"
        :width="getColumnWidth(column)"
        :align="getColumnAlign(column)"
      >
        <template #default="scope">
          <!-- 日期时间类型 -->
          <el-date-picker
            v-if="isDateTime(column.Type) && scope.row[column.Name] !== ''"
            v-model="scope.row[column.Name]"
            :type="getDatePickerType(column.Type)"
            :format="getDateFormat(column.Type)"
            :value-format="getDateFormat(column.Type)"
            disabled
            size="small"
            class="github-date-picker"
          />
          <!-- NULL 值 -->
          <span
            v-else-if="scope.row[column.Name] === ''"
            class="null-value"
          >
            NULL
          </span>
          <!-- 其他类型 -->
          <span
            v-else
            :class="{
              'primary-key': column.IsPrimary,
              'number-value': isNumberType(column.Type)
            }"
          >
            {{ scope.row[column.Name] }}
          </span>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[100, 500, 1000, 2000]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        background
        class="github-pagination"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { DatabaseConfig } from '../types/database'
import { GetTableStructure, GetTableData, GetTableRowCount } from '../../wailsjs/go/main/App'
import type { database } from '../../wailsjs/go/models'

// 定义接口
interface TableData {
  Data: Array<{
    Columns: Array<{
      Name: string
      Value: string
    }>
  }>
  Total: number
}

interface TableColumn extends database.ColumnInfo {
  width?: number
  align?: string
}

const props = defineProps<{
  config: DatabaseConfig
  database: string
  table: string
}>()

const loading = ref(false)
const tableData = ref<Record<string, string>[]>([])
const columns = ref<TableColumn[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(1000)

// 判断是否为日期时间类型
const isDateTime = (type: string): boolean => {
  const lowerType = type.toLowerCase()
  return lowerType.includes('datetime') || 
         lowerType.includes('timestamp') || 
         lowerType.includes('date') || 
         lowerType.includes('time')
}

// 获取日期选择器类型
const getDatePickerType = (type: string): string => {
  const lowerType = type.toLowerCase()
  if (lowerType.includes('datetime') || lowerType.includes('timestamp')) {
    return 'datetime'
  } else if (lowerType.includes('date')) {
    return 'date'
  } else if (lowerType.includes('time')) {
    return 'time'
  }
  return 'datetime'
}

// 获取日期格式
const getDateFormat = (type: string): string => {
  const lowerType = type.toLowerCase()
  if (lowerType.includes('datetime') || lowerType.includes('timestamp')) {
    return 'YYYY-MM-DD HH:mm:ss'
  } else if (lowerType.includes('date')) {
    return 'YYYY-MM-DD'
  } else if (lowerType.includes('time')) {
    return 'HH:mm:ss'
  }
  return 'YYYY-MM-DD HH:mm:ss'
}

// 根据列类型设置合适的宽度
const getColumnWidth = (column: any) => {
  const type = column.Type.toLowerCase()
  if (type.includes('int')) {
    return 100
  } else if (type.includes('datetime') || type.includes('timestamp')) {
    return 180
  } else if (type.includes('date')) {
    return 120
  } else if (type.includes('time')) {
    return 100
  } else if (type.includes('varchar') || type.includes('char')) {
    const length = parseInt(type.match(/\((\d+)\)/)?.[1] || '0')
    return length < 50 ? 120 : 200
  } else if (type.includes('text')) {
    return 300
  }
  return 150
}

// 判断是否为数字类型
const isNumberType = (type: string): boolean => {
  const lowerType = type.toLowerCase()
  return lowerType.includes('int') || 
         lowerType.includes('float') || 
         lowerType.includes('double') || 
         lowerType.includes('decimal') ||
         lowerType.includes('number')
}

// 获取列对齐方式
const getColumnAlign = (column: any): string => {
  const type = column.Type.toLowerCase()
  if (isNumberType(type)) {
    return 'right'
  }
  return 'left'
}

// 获取表结构
const loadTableStructure = async () => {
  console.log('Loading table structure:', props)
  try {
    const structure = await GetTableStructure(props.config, props.database, props.table)
    columns.value = structure.map(col => ({
      ...col,
      width: getColumnWidth(col),
      align: getColumnAlign(col)
    }))
  } catch (error) {
    console.error('Failed to load table structure:', error)
    ElMessage.error('获取表结构失败: ' + error)
  }
}

// 获取表数据
const loadTableData = async () => {
  console.log('Loading table data:', props)
  loading.value = true
  try {
    const result = await GetTableData(
      props.config,
      props.database,
      props.table,
      (currentPage.value - 1) * pageSize.value,
      pageSize.value
    )
    console.log('Table data result:', result)

    // 直接使用返回的数据，因为后端已经返回了正确格式的 map
    tableData.value = result || []
    
    // 获取总行数
    const count = await GetTableRowCount(props.config, props.database, props.table)
    total.value = count || 0

  } catch (error) {
    console.error('Failed to load table data:', error)
    ElMessage.error('获取���数据失败: ' + error)
  } finally {
    loading.value = false
  }
}

const handleSizeChange = (val: number) => {
  pageSize.value = val
  loadTableData()
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val
  loadTableData()
}

// 监听属性变化
watch(
  () => [props.config, props.database, props.table],
  () => {
    console.log('Props changed:', props)
    currentPage.value = 1
    loadTableStructure()
    loadTableData()
  },
  { immediate: true }
)
</script>

<style scoped>
.table-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
  border: 1px solid #d0d7de;
  border-radius: 6px;
}

.pagination-container {
  padding: 16px;
  background: #ffffff;
  border-top: 1px solid #d0d7de;
}

.primary-key {
  color: #0969da;
  font-weight: 600;
}

.number-value {
  font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace;
}

.null-value {
  color: #8c959f;
  font-style: italic;
}

:deep(.el-table) {
  border: none !important;
  border-radius: 6px 6px 0 0;
  overflow: hidden;
}

:deep(.el-table__header) {
  border-radius: 6px 6px 0 0;
}

:deep(.el-table__row) {
  &:hover > td {
    background-color: #f6f8fa !important;
  }
}

:deep(.el-date-editor.el-input) {
  width: 100%;
}

.github-date-picker {
  :deep(.el-input.is-disabled .el-input__wrapper) {
    background-color: transparent;
    box-shadow: none !important;
    border: none;
  }

  :deep(.el-input.is-disabled .el-input__inner) {
    color: #1a7f37;
    -webkit-text-fill-color: #1a7f37;
  }
}

.github-pagination {
  :deep(.el-pagination__total) {
    color: #57606a;
  }

  :deep(.el-pagination__sizes) {
    .el-select .el-input {
      --el-select-input-focus-border-color: #0969da;
    }
  }

  :deep(.btn-prev),
  :deep(.btn-next),
  :deep(.el-pager li) {
    background-color: #f6f8fa;
    border: 1px solid #d0d7de;
    color: #24292f;
    font-weight: 500;

    &:hover {
      background-color: #0969da;
      border-color: #0969da;
      color: #ffffff;
    }

    &.is-active {
      background-color: #0969da;
      border-color: #0969da;
      color: #ffffff;
    }
  }
}
</style> 