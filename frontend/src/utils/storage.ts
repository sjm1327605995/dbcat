import { ElMessage } from 'element-plus'
import type { TreeNodeData } from '../types/tree'

export interface StorageData {
  version: string
  lastUpdated: number
  groups: TreeNodeData[]
}

export class StorageManager {
  private static instance: StorageManager
  private readonly STORAGE_KEY = 'dbcat_connections'
  private readonly ENCRYPTION_KEY = 'dbcat_secret_key'
  private data: StorageData

  private constructor() {
    this.data = this.loadFromStorage()
  }

  public static getInstance(): StorageManager {
    if (!StorageManager.instance) {
      StorageManager.instance = new StorageManager()
    }
    return StorageManager.instance
  }

  // 加密敏感数据
  private encryptData(data: string): string {
    return btoa(data)
  }

  // 解密数据
  private decryptData(data: string): string {
    return atob(data)
  }

  // 保存数据到本地存储
  public saveToStorage(treeData: TreeNodeData[] | undefined): void {
    try {
      const processedData = this.processDataForStorage(treeData || [])
      
      const storageData: StorageData = {
        version: '1.0.0',
        lastUpdated: Date.now(),
        groups: processedData
      }

      const data = JSON.stringify(storageData)
      localStorage.setItem(this.STORAGE_KEY, data)
      
      this.emitStorageEvent('save')
    } catch (error) {
      console.error('保存数据失败:', error)
      ElMessage.error('保存数据失败')
    }
  }

  // 从本地存储加载数据
  public loadFromStorage(): StorageData {
    try {
      const data = localStorage.getItem(this.STORAGE_KEY)
      if (data) {
        const storageData = JSON.parse(data)
        storageData.groups = this.processDataFromStorage(storageData.groups || [])
        return storageData
      }
    } catch (error) {
      console.error('加载数据失败:', error)
      ElMessage.error('加载数据失败')
    }
    return {
      version: '1.0.0',
      lastUpdated: Date.now(),
      groups: []
    }
  }

  // 处理数据用于存储
  private processDataForStorage(nodes: TreeNodeData[]): TreeNodeData[] {
    return nodes.map(node => {
      const processedNode = { ...node }
      if (node.type === 'connection' && node.config) {
        const password = node.config.Password || node.config.password || ''
        processedNode.config = {
          ...node.config,
          Password: this.encryptData(password)
        }
        if (node.cache) {
          processedNode.cache = node.cache
        }
      }
      if (node.children) {
        processedNode.children = this.processDataForStorage(node.children)
      }
      return processedNode
    })
  }

  // 处理从存储加载的数据
  private processDataFromStorage(nodes: TreeNodeData[]): TreeNodeData[] {
    return nodes.map(node => {
      const processedNode = { ...node }
      if (node.type === 'connection' && node.config) {
        const password = node.config.Password || node.config.password || ''
        processedNode.config = {
          ...node.config,
          Password: this.decryptData(password),
          password: this.decryptData(password)
        }
        if (node.cache) {
          processedNode.cache = node.cache
        }
      }
      if (node.children) {
        processedNode.children = this.processDataFromStorage(node.children)
      }
      return processedNode
    })
  }

  // 导出数据
  public exportData(): string {
    const data = this.loadFromStorage()
    return JSON.stringify(data, null, 2)
  }

  // 导入数据
  public importData(jsonData: string): boolean {
    try {
      const data = JSON.parse(jsonData)
      if (this.validateImportData(data)) {
        this.saveToStorage(data.groups || [])
        return true
      }
      return false
    } catch (error) {
      console.error('导入数据失败:', error)
      return false
    }
  }

  // 验证导入的数据
  private validateImportData(data: any): boolean {
    return (
      data &&
      typeof data === 'object' &&
      typeof data.version === 'string' &&
      Array.isArray(data.groups)
    )
  }

  // 触发存储事件
  private emitStorageEvent(type: 'save' | 'load' | 'import' | 'export'): void {
    window.dispatchEvent(new CustomEvent('dbcat-storage', {
      detail: {
        type,
        timestamp: Date.now()
      }
    }))
  }
} 