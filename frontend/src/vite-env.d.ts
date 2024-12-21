/// <reference types="vite/client" />

declare module '*.vue' {
    import type {DefineComponent} from 'vue'
    const component: DefineComponent<{}, {}, any>
    export default component
}

// 添加 App.custom 模块的声明
declare module '../../wailsjs/go/main/App.custom' {
    import type { DatabaseConfig } from '../types/database'

    export interface CreateDatabaseOptions {
        name: string
        charset: string
        collation: string
    }

    export function CreateDatabase(config: DatabaseConfig, options: CreateDatabaseOptions): Promise<void>
}
