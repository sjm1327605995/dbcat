package database

import (
	"fmt"
)

// DBAdapter 数据库适配器接口
type DBAdapter interface {
	Connect() error
	Close() error
	Ping() error
	GetDatabases() ([]DatabaseInfo, error)
	GetSchemas(dbName string) ([]SchemaInfo, error)
	GetTables(dbName, schema string) ([]TableInfo, error)
	GetTableColumns(dbName, tableName string) ([]ColumnInfo, error)
	GetTableRowCount(dbName, tableName string) (int64, error)
	QueryTableData(dbName, tableName string, offset, limit int) ([]map[string]string, error)
}

// DatabaseInfo 数据库信息
type DatabaseInfo struct {
	Name string `json:"Name"`
}

// SchemaInfo schema信息
type SchemaInfo struct {
	Name string `json:"Name"`
}

// TableInfo 表信息
type TableInfo struct {
	Name    string `json:"Name"`
	Comment string `json:"Comment"`
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name      string `json:"Name"`
	Type      string `json:"Type"`
	Length    int    `json:"Length"`
	Nullable  bool   `json:"Nullable"`
	IsPrimary bool   `json:"IsPrimary"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `json:"Type"`
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	User     string `json:"User"`
	Password string `json:"Password"`
	Database string `json:"Database"`
	SSLMode  string `json:"SSLMode"`
}

// DBFactory 数据库工厂
type DBFactory struct{}

// NewDBFactory 创建数据库工厂
func NewDBFactory() *DBFactory {
	return &DBFactory{}
}

// CreateAdapter 创建数据库适配器
func (f *DBFactory) CreateAdapter(config DatabaseConfig) (DBAdapter, error) {
	switch config.Type {
	case "mysql":
		return NewMySQLAdapter(config), nil
	case "postgres":
		return NewPostgresAdapter(config), nil
	case "sqlite":
		return NewSQLiteAdapter(config), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}
