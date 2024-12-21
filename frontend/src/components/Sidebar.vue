<template>
  <el-aside class="sidebar" :style="{ width: sidebarWidth }">
    <!-- 移除顶部菜单栏 -->
    <div class="resize-bar" @mousedown="handleMousedown"></div>
    <div 
      class="tree-container" 
      @contextmenu.prevent="handleTreeContainerContextMenu"
    >
      <el-tree
        ref="treeRef"
        :data="treeData"
        :props="defaultProps"
        draggable
        :allow-drag="handleDragStart"
        :allow-drop="handleDragEnter"
        @node-drag-end="handleDrop"
        @node-click="handleNodeClick"
        @node-contextmenu="handleContextMenu"
      >
        <template #default="{ node, data }">
          <TreeNodeContent :data="data" />
        </template>
      </el-tree>
    </div>

    <!-- 对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form v-if="dialogType === 'name'" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
      </el-form>

      <el-form v-else label-width="80px">
        <el-form-item label="类型">
          <el-select v-model="connectionForm.type">
            <el-option label="MySQL" value="mysql" />
            <el-option label="PostgreSQL" value="postgres" />
            <el-option label="SQLite" value="sqlite" />
          </el-select>
        </el-form-item>

        <template v-if="connectionForm.type !== 'sqlite'">
          <el-form-item label="主机">
            <el-input v-model="connectionForm.host" />
          </el-form-item>

          <el-form-item label="端口">
            <el-input-number v-model="connectionForm.port" :min="1" :max="65535" />
          </el-form-item>

          <el-form-item label="用户名">
            <el-input v-model="connectionForm.user" />
          </el-form-item>

          <el-form-item label="密码">
            <el-input v-model="connectionForm.password" type="password" show-password />
          </el-form-item>
        </template>

        <el-form-item label="数据库">
          <el-input v-model="connectionForm.database" />
        </el-form-item>

        <el-form-item v-if="connectionForm.type === 'postgres'" label="SSL模式">
          <el-select v-model="connectionForm.sslMode">
            <el-option label="禁用" value="disable" />
            <el-option label="必需" value="require" />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleDialogConfirm">确认</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 右键菜单 -->
    <ul v-show="contextMenuVisible" :style="contextMenuStyle" class="context-menu">
      <!-- 组节点菜单 -->
      <template v-if="selectedNode?.type === 'group'">
        <li @click="handleContextMenuCommand('newGroup')">新建组</li>
        <li @click="handleContextMenuCommand('newConnection')">新建连接</li>
        <li @click="handleContextMenuCommand('rename')">重命名</li>
        <li @click="handleContextMenuCommand('delete')">删除</li>
      </template>

      <!-- 连接节点菜单 -->
      <template v-else-if="selectedNode?.type === 'connection'">
        <li @click="handleContextMenuCommand('refresh')">刷新</li>
        <li @click="handleContextMenuCommand('edit')">编辑连接</li>
        <li @click="handleContextMenuCommand('rename')">重命名</li>
        <li @click="handleContextMenuCommand('delete')">删除</li>
      </template>

      <!-- 数据库节点菜单 -->
      <template v-else-if="selectedNode?.type === 'database'">
        <li @click="handleContextMenuCommand('refresh')">刷新</li>
        <li @click="handleContextMenuCommand('export')">导出数据库</li>
        <li @click="handleContextMenuCommand('delete')">删除数据库</li>
      </template>

      <!-- 表节点菜单 -->
      <template v-else-if="selectedNode?.type === 'table'">
        <li @click="handleContextMenuCommand('refresh')">刷新</li>
        <li @click="handleContextMenuCommand('export')">导出表</li>
        <li @click="handleContextMenuCommand('truncate')">清空表</li>
        <li @click="handleContextMenuCommand('delete')">删除表</li>
      </template>

      <!-- 空白区域菜单（根级别） -->
      <template v-else>
        <li @click="handleContextMenuCommand('newGroup')">新建组</li>
        <li @click="handleContextMenuCommand('newConnection')">新建连接</li>
      </template>
    </ul>
  </el-aside>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CreateConnection, GetDatabases, GetSchemas, GetTables, TestConnection } from '../../wailsjs/go/main/App'
import type { DatabaseConfig, ConnectionFormData } from '../types/database'
import type { TreeNodeData, DragNode, TableInfo } from '../types/tree'
import { StorageManager } from '../utils/storage'
// import SidebarToolbar from './SidebarToolbar.vue'
import TreeNodeContent from './TreeNodeContent.vue'

const emit = defineEmits<{
  (e: 'select-table', data: { config: DatabaseConfig; database: string; table: string }): void
  (e: 'new-query', data: { config: DatabaseConfig; database: string }): void
  (e: 'connection-select', connection: TreeNodeData): void
}>()

const treeRef = ref()
const sidebarWidth = ref('200px')
const dialogVisible = ref(false)
const dialogTitle = ref('')
const dialogType = ref<'name' | 'connection'>('name')
const contextMenuVisible = ref(false)
const contextMenuStyle = ref({
  left: '0px',
  top: '0px'
})

const form = ref({
  name: ''
})

const connectionForm = ref<ConnectionFormData>({
  type: 'mysql',
  host: 'localhost',
  port: 3306,
  user: '',
  password: '',
  database: '',
  sslMode: 'disable'
})

const selectedNode = ref<TreeNodeData | null>(null)
const selectedConnection = ref<TreeNodeData | null>(null)
const storageManager = StorageManager.getInstance()
const treeData = ref<TreeNodeData[]>([])

const defaultProps = {
  children: 'children',
  label: 'label'
} as const

// 初始化树数据
onMounted(() => {
  const data = storageManager.loadFromStorage()
  treeData.value = data.groups
})

// 删除节点
const deleteFromNodes = (nodes: TreeNodeData[]): boolean => {
  for (let i = 0; i < nodes.length; i++) {
    if (nodes[i].id === selectedNode.value?.id) {
      nodes.splice(i, 1)
      return true
    }
    if (nodes[i].children?.length) {
      if (deleteFromNodes(nodes[i].children|| [])) {
        return true
      }
    }
  }
  return false
}

// 查找父节点
const findParentNodeById = (nodes: TreeNodeData[], id: string): TreeNodeData | null => {
  for (const node of nodes) {
    if (node.children) {
      for (const child of node.children) {
        if (child.id === id) {
          return node
        }
        const found = findParentNodeById([child], id)
        if (found) {
          return found
        }
      }
    }
  }
  return null
}

// 添加一个保存函数
const saveTreeData = () => {
  storageManager.saveToStorage(treeData.value || [])
}

// 加载数据库列表
const loadDatabases = async (node: TreeNodeData) => {
  if (!node.config) return

  try {
    const databases = await GetDatabases(node.config)
    node.children = databases.map(db => ({
      id: `${node.id}-${db.Name}`,
      label: db.Name,
      type: 'database' as const,
      dbType: node.dbType,
      children: []
    }))
    
    node.cache = {
      databases: databases.map(db => ({
        Name: db.Name,
        Tables: []
      })),
      lastUpdated: Date.now()
    }
    node.isConnected = true
    saveTreeData()
  } catch (error) {
    node.isConnected = false
    ElMessage.error('获取数据库列表失败: ' + error)
  }
}

// 加载表列表
const loadTables = async (node: TreeNodeData, parentNode: TreeNodeData) => {
  if (!parentNode.config) return

  try {
    const tables = await GetTables(parentNode.config, node.label, '')
    node.children = tables.map(table => ({
      id: `${node.id}-${table.Name}`,
      label: table.Name,
      type: 'table' as const
    }))

    if (!parentNode.cache) {
      parentNode.cache = {
        databases: [],
        lastUpdated: Date.now()
      }
    }
    const dbIndex = parentNode.cache.databases.findIndex(db => db.Name === node.label)
    
    const processedTables: TableInfo[] = tables.map(t => {
      const tableInfo: TableInfo = {
        Name: t.Name,
        Comment: t.Comment || ''
      }

      if ('Type' in t && typeof t.Type === 'string') {
        tableInfo.Type = t.Type
      }
      if ('IsPrimary' in t && typeof t.IsPrimary === 'boolean') {
        tableInfo.IsPrimary = t.IsPrimary
      }

      return tableInfo
    })
    
    if (dbIndex === -1) {
      parentNode.cache.databases.push({
        Name: node.label,
        Tables: processedTables
      })
    } else {
      parentNode.cache.databases[dbIndex].Tables = processedTables
    }
    saveTreeData()
  } catch (error) {
    ElMessage.error('获取数据结构失败: ' + error)
  }
}

// 处理节点点击
const handleNodeClick = async (data: TreeNodeData) => {
  selectedNode.value = data
  if (data.type === 'connection') {
    selectedConnection.value = data
    emit('connection-select', data)
    await loadDatabases(data)
  } else if (data.type === 'database') {
    const parentNode = findParentNodeById(treeData.value, data.id)
    if (parentNode) {
      selectedConnection.value = parentNode
      emit('connection-select', parentNode)
      await loadTables(data, parentNode)
      
      if (parentNode.config) {
        emit('new-query', {
          config: parentNode.config,
          database: data.label
        })
      }
    }
  } else if (data.type === 'table') {
    const parts = data.id.split('-')
    const dbNodeId = parts.slice(0, -1).join('-')
    const dbNode = findNodeById(treeData.value, dbNodeId)
    if (!dbNode) return

    const connNodeId = parts[0]
    const connNode = findNodeById(treeData.value, connNodeId)
    if (!connNode?.config) return

    emit('select-table', {
      config: connNode.config,
      database: dbNode.label,
      table: data.label
    })
  }
}

// 添加 findNodeById 函数
const findNodeById = (nodes: TreeNodeData[], id: string): TreeNodeData | null => {
  for (const node of nodes) {
    if (node.id === id) {
      return node
    }
    if (node.children) {
      const found = findNodeById(node.children, id)
      if (found) {
        return found
      }
    }
  }
  return null
}

// 处理右键菜单
const handleContextMenu = (event: MouseEvent, data: TreeNodeData) => {
  event.preventDefault()
  selectedNode.value = data
  contextMenuVisible.value = true
  contextMenuStyle.value = {
    left: event.clientX + 'px',
    top: event.clientY + 'px'
  }
}

// 处理右键菜单命令
const handleContextMenuCommand = async (command: string) => {
  switch (command) {
    case 'newGroup':
      dialogType.value = 'name'
      dialogTitle.value = '新建组'
      form.value.name = ''
      dialogVisible.value = true
      break

    case 'newConnection':
      dialogType.value = 'connection'
      dialogTitle.value = '新建连接'
      connectionForm.value = {
        type: 'mysql',
        host: 'localhost',
        port: 3306,
        user: '',
        password: '',
        database: '',
        sslMode: 'disable'
      }
      dialogVisible.value = true
      break

    case 'refresh':
      if (selectedNode.value?.type === 'connection') {
        await loadDatabases(selectedNode.value)
      } else if (selectedNode.value?.type === 'database') {
        const parentNode = findParentNodeById(treeData.value, selectedNode.value.id)
        if (parentNode) {
          await loadTables(selectedNode.value, parentNode)
        }
      }
      break

    case 'rename':
      if (selectedNode.value) {
        dialogType.value = 'name'
        dialogTitle.value = '重命名'
        form.value.name = selectedNode.value.label
        dialogVisible.value = true
      }
      break

    case 'delete':
      if (selectedNode.value) {
        ElMessageBox.confirm(
          `确定要删除 ${selectedNode.value.label} 吗？`,
          '警告',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
          }
        ).then(() => {
          if (treeData.value) {
            deleteFromNodes(treeData.value)
            saveTreeData()
            ElMessage.success('删除成功')
          }
        }).catch(() => {
          // 取消删除
        })
      }
      break

    // 可以添加其他命令的处理...
    case 'edit':
    case 'export':
    case 'truncate':
      ElMessage.info('功能开发中...')
      break
  }

  // 关闭右键菜单
  contextMenuVisible.value = false
}

// 处理对话框确认
const handleDialogConfirm = async () => {
  if (dialogType.value === 'name') {
    if (!form.value.name) {
      ElMessage.warning('名称不能为空')
      return
    }

    if (dialogTitle.value === '重命名' && selectedNode.value) {
      selectedNode.value.label = form.value.name
    } else {
      const newNode: TreeNodeData = {
        id: Date.now().toString(),
        label: form.value.name,
        type: dialogTitle.value === '新建组' ? 'group' : 'connection',
        children: dialogTitle.value === '新建组' ? [] : undefined
      }

      if (selectedNode.value?.type === 'group') {
        if (!selectedNode.value.children) {
          selectedNode.value.children = []
        }
        selectedNode.value.children.push(newNode)
      } else {
        if (!treeData.value) {
          treeData.value = []
        }
        treeData.value.push(newNode)
      }
    }
    saveTreeData()
  } else {
    try {
      const config: DatabaseConfig = {
        Type: connectionForm.value.type,
        Host: connectionForm.value.host,
        Port: connectionForm.value.port,
        User: connectionForm.value.user,
        Password: connectionForm.value.password,
        Database: connectionForm.value.database,
        SSLMode: connectionForm.value.type === 'postgres' ? connectionForm.value.sslMode : ''
      }

      await TestConnection(config)
      const result = await CreateConnection(config)
      
      const newNode: TreeNodeData = {
        id: Date.now().toString(),
        label: connectionForm.value.database || '新连接',
        type: 'connection',
        dbType: connectionForm.value.type,
        children: [],
        config: {
          ...config,
          Password: config.Password
        }
      }

      if (selectedNode.value?.type === 'group') {
        if (!selectedNode.value.children) {
          selectedNode.value.children = []
        }
        selectedNode.value.children.push(newNode)
      } else if (treeData.value) {
        treeData.value.push(newNode)
      }

      ElMessage.success(result)
    } catch (error) {
      ElMessage.error('连接失败: ' + error)
      return
    }
  }

  if (treeData.value) {
    storageManager.saveToStorage(treeData.value)
  }
  dialogVisible.value = false
}

// 处理拖拽
const handleDrop = (draggingNode: DragNode, dropNode: { data: TreeNodeData }, dropType: string) => {
  if (dropNode.data.type !== 'group' || dropType !== 'inner') {
    return false
  }

  dropNode.data.children = dropNode.data.children || []
  const parent = draggingNode.parent
  parent.data.children = parent.data.children || []
  
  const index = parent.data.children.findIndex(d => d.id === draggingNode.data.id)
  if (index !== -1) {
    parent.data.children.splice(index, 1)
  }

  dropNode.data.children.push(draggingNode.data)
  saveTreeData()
}

const handleDragStart = (node: { data: TreeNodeData }) => {
  return node.data.type === 'connection'
}

const handleDragEnter = (_: any, dropNode: { data: TreeNodeData }) => {
  return dropNode.data.type === 'group'
}

// 处理树容器右键菜单
const handleTreeContainerContextMenu = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.el-tree-node')) {
    event.preventDefault()
    selectedNode.value = null
    contextMenuVisible.value = true
    contextMenuStyle.value = {
      left: event.clientX + 'px',
      top: event.clientY + 'px'
    }
  }
}

// 处理点击外部
const handleClickOutside = () => {
  contextMenuVisible.value = false
}

// 处理鼠标按下事件（用于调整侧边栏宽度）
const handleMousedown = (e: MouseEvent) => {
  const startX = e.clientX
  const startWidth = parseInt(sidebarWidth.value)

  const handleMousemove = (e: MouseEvent) => {
    const deltaX = e.clientX - startX
    const newWidth = Math.max(200, Math.min(500, startWidth + deltaX))
    sidebarWidth.value = `${newWidth}px`
  }

  const handleMouseup = () => {
    document.removeEventListener('mousemove', handleMousemove)
    document.removeEventListener('mouseup', handleMouseup)
  }

  document.addEventListener('mousemove', handleMousemove)
  document.addEventListener('mouseup', handleMouseup)
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

// 处理刷新
const handleRefresh = async () => {
  if (selectedNode.value?.type === 'connection') {
    await loadDatabases(selectedNode.value)
  } else if (selectedNode.value?.type === 'database') {
    const parentNode = findParentNodeById(treeData.value, selectedNode.value.id)
    if (parentNode) {
      await loadTables(selectedNode.value, parentNode)
    }
  }
}

// 处理新建连接
const handleNewConnection = () => {
  dialogType.value = 'connection'
  dialogTitle.value = '新建连接'
  connectionForm.value = {
    type: 'mysql',
    host: 'localhost',
    port: 3306,
    user: '',
    password: '',
    database: '',
    sslMode: 'disable'
  }
  dialogVisible.value = true
}

// 处理新建查询
const handleNewQuery = () => {
  if (selectedConnection.value?.config) {
    emit('new-query', {
      config: selectedConnection.value.config,
      database: selectedConnection.value.label
    })
  }
}

// 暴露方法给父组件
defineExpose({
  handleContextMenuCommand
})
</script>

<style scoped>
.sidebar {
  position: relative;
  background-color: #f5f7fa;
  border-right: 1px solid #dcdfe6;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.tree-container {
  flex: 1;
  overflow: auto;
  padding: 8px;
  padding-top: 0;
}

.resize-bar {
  position: absolute;
  right: 0;
  top: 0;
  width: 5px;
  height: 100%;
  cursor: col-resize;
  background-color: transparent;
  z-index: 1;
}

.resize-bar:hover {
  background-color: #409EFF;
}

:deep(.el-tree) {
  background-color: transparent;
}

:deep(.el-tree-node__content) {
  height: 28px;
  padding-right: 8px;
}

:deep(.el-tree-node__content:hover) {
  background-color: #e6f1fc;
}

:deep(.el-tree-node.is-current > .el-tree-node__content) {
  background-color: #e6f1fc;
  color: #409EFF;
}

.context-menu {
  position: fixed;
  z-index: 9999;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 5px 0;
  min-width: 150px;
  box-shadow: 0 2px 12px 0 rgba(0,0,0,0.1);
}

.context-menu li {
  list-style: none;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.context-menu li:hover {
  background-color: #f5f7fa;
  color: #409EFF;
}

:deep(.el-tree-node__expand-icon) {
  color: #909399;
}

:deep(.el-tree-node__expand-icon.expanded) {
  transform: rotate(90deg);
}

:deep(.el-tree-node__expand-icon.is-leaf) {
  color: transparent;
}

.menu-bar {
  display: none;
}
</style> 