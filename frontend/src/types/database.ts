export type DatabaseType = 'mysql' | 'postgres' | 'sqlite'
export type SSLMode = 'disable' | 'require'

export interface DatabaseConfig {
  Type: DatabaseType
  Host: string
  Port: number
  User: string
  Password: string
  password?: string
  Database: string
  SSLMode: string
}

export interface DatabaseInfo {
  Name: string
}

export interface SchemaInfo {
  Name: string
}

export interface TableInfo {
  Name: string
  Comment: string
}

export interface ColumnInfo {
  Name: string
  Type: string
  Length: number
  Nullable: boolean
  IsPrimary: boolean
}

export interface ConnectionFormData {
  type: DatabaseType
  host: string
  port: number
  user: string
  password: string
  database: string
  sslMode: SSLMode
} 