import type { DatabaseConfig, DatabaseInfo, TableInfo } from './database'

export interface TreeNodeData {
  id: string
  label: string
  type: 'group' | 'connection' | 'database' | 'schema' | 'table'
  dbType?: 'mysql' | 'postgres' | 'sqlite'
  children?: TreeNodeData[]
  config?: DatabaseConfig
  isConnected?: boolean
  cache?: ConnectionCache
}

export interface DragNode {
  data: TreeNodeData
  parent: {
    data: {
      children: TreeNodeData[]
    }
  }
}

export interface CachedDatabase extends DatabaseInfo {
  Tables: TableInfo[]
}

export interface ConnectionCache {
  databases: CachedDatabase[]
  lastUpdated: number
}

export interface TreeProps {
  children: 'children'
  label: 'label'
} 