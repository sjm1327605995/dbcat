import type { DatabaseConfig } from './database'
import type { database } from '../../wailsjs/go/models'

export interface TableInfo {
  Name: string
  Comment?: string
  Type?: string
  IsPrimary?: boolean
}

export interface DatabaseInfo {
  Name: string
  CharacterSet?: string
  Collation?: string
}

export interface TreeNodeData {
  id: string
  label: string
  type: 'group' | 'connection' | 'database' | 'table'
  dbType?: 'mysql' | 'postgres' | 'sqlite'
  isConnected?: boolean
  config?: DatabaseConfig
  children?: TreeNodeData[]
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