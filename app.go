package main

import (
	"context"
	"dbcat/database"
	"fmt"
	"sync"
)

var (
	instance *App
	once     sync.Once
)

// ColumnValue 表示列值
type ColumnValue struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

// RowData 表示一行数据
type RowData struct {
	Columns []ColumnValue `json:"Columns"`
}

// TableStructure 表结构
type TableStructure struct {
	Columns []database.ColumnInfo `json:"Columns"`
}

// TableData 表数据
type TableData struct {
	Data  []RowData `json:"Data"`
	Total int64     `json:"Total"`
}

// App struct
type App struct {
	ctx            context.Context
	factory        *database.DBFactory
	connections    map[string]database.DBAdapter
	connectionLock sync.RWMutex
}

// GetApp 获取 App 单例
func GetApp() *App {
	once.Do(func() {
		instance = &App{
			factory:     database.NewDBFactory(),
			connections: make(map[string]database.DBAdapter),
		}
	})
	return instance
}

// NewApp creates a new App application struct
func NewApp() *App {
	return GetApp()
}

// startup is called when the app starts. The context is saved
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// getConnectionKey 生成连接的唯一键
func (a *App) getConnectionKey(config database.DatabaseConfig) string {
	return fmt.Sprintf("%s://%s@%s:%d/%s", config.Type, config.User, config.Host, config.Port, config.Database)
}

// CreateConnection 创建新的数据库连接
func (a *App) CreateConnection(config database.DatabaseConfig) (string, error) {
	key := a.getConnectionKey(config)

	a.connectionLock.RLock()
	if adapter, exists := a.connections[key]; exists {
		a.connectionLock.RUnlock()
		// 测试连接是否还有效
		if err := adapter.Connect(); err == nil {
			adapter.Close()
			return fmt.Sprintf("Using existing connection to %s database", config.Type), nil
		}
	} else {
		a.connectionLock.RUnlock()
	}

	adapter, err := a.factory.CreateAdapter(config)
	if err != nil {
		return "", fmt.Errorf("创建适配器失败: %v", err)
	}

	if err := adapter.Connect(); err != nil {
		return "", fmt.Errorf("连接数据库失败: %v", err)
	}

	a.connectionLock.Lock()
	a.connections[key] = adapter
	a.connectionLock.Unlock()

	return fmt.Sprintf("Successfully connected to %s database", config.Type), nil
}

// getConnection 获取已存在的连接或创建新连接
func (a *App) getConnection(config database.DatabaseConfig) (database.DBAdapter, error) {
	key := a.getConnectionKey(config)

	a.connectionLock.RLock()
	adapter, exists := a.connections[key]
	a.connectionLock.RUnlock()

	if exists {
		if err := adapter.Ping(); err == nil {
			return adapter, nil
		}
		// 如果连接失效，删除旧连接
		a.connectionLock.Lock()
		adapter.Close() // 确保关闭旧连接
		delete(a.connections, key)
		a.connectionLock.Unlock()
	}

	// 创建新连接
	adapter, err := a.factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("创建适配器失败: %v", err)
	}

	if err := adapter.Connect(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	a.connectionLock.Lock()
	a.connections[key] = adapter
	a.connectionLock.Unlock()

	return adapter, nil
}

// GetDatabases 获取所有数据库
func (a *App) GetDatabases(config database.DatabaseConfig) ([]database.DatabaseInfo, error) {
	adapter, err := a.getConnection(config)
	if err != nil {
		return nil, err
	}
	return adapter.GetDatabases()
}

// GetSchemas 获取指定数据库的所有schema
func (a *App) GetSchemas(config database.DatabaseConfig, dbName string) ([]database.SchemaInfo, error) {
	// MySQL 不支持 schema
	if config.Type == "mysql" {
		return nil, fmt.Errorf("MySQL does not support schemas")
	}

	adapter, err := a.getConnection(config)
	if err != nil {
		return nil, err
	}
	return adapter.GetSchemas(dbName)
}

// GetTables 获取指定schema的所有表
func (a *App) GetTables(config database.DatabaseConfig, dbName, schema string) ([]database.TableInfo, error) {
	adapter, err := a.getConnection(config)
	if err != nil {
		return nil, err
	}

	// MySQL 不需要 schema 参数
	if config.Type == "mysql" {
		return adapter.GetTables(dbName, "")
	}
	return adapter.GetTables(dbName, schema)
}

// TestConnection 测试数据库连接
func (a *App) TestConnection(config database.DatabaseConfig) error {
	adapter, err := a.factory.CreateAdapter(config)
	if err != nil {
		return fmt.Errorf("创建适配器失败: %v", err)
	}

	if err := adapter.Connect(); err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	defer adapter.Close()
	return nil
}

// GetTableStructure 获取表结构
func (a *App) GetTableStructure(config database.DatabaseConfig, dbName, tableName string) (*TableStructure, error) {
	adapter, err := a.getConnection(config)
	if err != nil {
		return nil, err
	}

	columns, err := adapter.GetTableColumns(dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("获取表结构失败: %v", err)
	}

	return &TableStructure{
		Columns: columns,
	}, nil
}

// GetTableData 获取表数据
func (a *App) GetTableData(config database.DatabaseConfig, dbName, tableName string, offset, limit int) (*TableData, error) {
	if limit > 2000 {
		limit = 2000 // 限制最大查询数量
	}

	adapter, err := a.getConnection(config)
	if err != nil {
		return nil, err
	}

	// 获取总数
	total, err := adapter.GetTableRowCount(dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("获取表总行数失败: %v", err)
	}

	// 获取数据
	rawData, err := adapter.QueryTableData(dbName, tableName, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("获取表数据失败: %v", err)
	}

	// 转换数据格式（使用预分配内存）
	data := make([]RowData, 0, len(rawData))
	for _, row := range rawData {
		columns := make([]ColumnValue, 0, len(row))
		for name, value := range row {
			columns = append(columns, ColumnValue{
				Name:  name,
				Value: value,
			})
		}
		data = append(data, RowData{Columns: columns})
	}

	// 清理不再使用的数据
	rawData = nil

	return &TableData{
		Data:  data,
		Total: total,
	}, nil
}
