package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
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
	CreateDatabase(name string, charset string, collation string) error
	GetCharsets() ([]CharsetInfo, error)
	ExecuteQuery(dbName, sql string) ([]map[string]string, error)
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

// CreateDatabaseOptions 创建数据库的选项
type CreateDatabaseOptions struct {
	Name      string `json:"name"`
	Charset   string `json:"charset"`
	Collation string `json:"collation"`
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

// CharsetInfo 字符集信息
type CharsetInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Collations  []string `json:"collations"`
}

// GetDatabaseCharsets 获取数据库支持的字符集
func (f *DBFactory) GetDatabaseCharsets(config DatabaseConfig) ([]CharsetInfo, error) {
	adapter, err := f.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("创建数据库适配器失败: %v", err)
	}
	defer adapter.Close()

	// 先连接数据库
	if err := adapter.Connect(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	return adapter.GetCharsets()
}

// BaseAdapter 基础适配器
type BaseAdapter struct {
	config DatabaseConfig
	db     *sqlx.DB
	// 添加一个字段来存储具体实现类的 Connect 方法
	connectFunc func() error
}

// SetConnectFunc 设置连接函数
func (a *BaseAdapter) SetConnectFunc(fn func() error) {
	a.connectFunc = fn
}

// DB 获取数据库连接，如果未连接则自动连接
func (a *BaseAdapter) DB() (*sqlx.DB, error) {
	if a.db == nil {
		if a.connectFunc == nil {
			return nil, fmt.Errorf("connect function not set")
		}
		if err := a.connectFunc(); err != nil {
			return nil, fmt.Errorf("failed to connect to database: %v", err)
		}
	}
	return a.db, nil
}

// SetupConnPool 设置连接池参数
func (a *BaseAdapter) SetupConnPool() {
	// 最大打开连接数
	a.db.SetMaxOpenConns(10)
	// 最大空闲连接数
	a.db.SetMaxIdleConns(5)
	// 连接最大生命周期
	a.db.SetConnMaxLifetime(time.Hour)
	// 空闲连接最大生命周期
	a.db.SetConnMaxIdleTime(30 * time.Minute)
}

// Close 关闭连接
func (a *BaseAdapter) Close() error {
	if a.db != nil {
		err := a.db.Close()
		a.db = nil
		return err
	}
	return nil
}

// Ping 测试连接是否有效
func (a *BaseAdapter) Ping() error {
	if a.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return a.db.Ping()
}

// SQLHelper SQL辅助函数
type SQLHelper struct{}

// RemoveComments 移除SQL中的注释
func (h *SQLHelper) RemoveComments(sql string) string {
	lines := strings.Split(sql, "\n")
	var result []string

	inMultilineComment := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if inMultilineComment {
			if idx := strings.Index(line, "*/"); idx != -1 {
				inMultilineComment = false
				line = line[idx+2:]
			} else {
				continue
			}
		}

		// 处理多行注释开始
		if idx := strings.Index(line, "/*"); idx != -1 {
			if endIdx := strings.Index(line[idx:], "*/"); endIdx != -1 {
				// 单行内的多行注释
				line = line[:idx] + line[idx+endIdx+2:]
			} else {
				// 跨行的多行注释
				inMultilineComment = true
				line = line[:idx]
			}
		}

		// 处理单行注释
		if idx := strings.Index(line, "--"); idx != -1 {
			line = line[:idx]
		}
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}

		// 添加非空行
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// SplitStatements 分割多个SQL语句
func (h *SQLHelper) SplitStatements(sql string) []string {
	sql = h.RemoveComments(sql)
	statements := strings.Split(sql, ";")
	var result []string

	for _, stmt := range statements {
		if stmt = strings.TrimSpace(stmt); stmt != "" {
			result = append(result, stmt)
		}
	}

	return result
}
