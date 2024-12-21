package main

import (
	"context"
	"fmt"
	"dbcat/database"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// CreateDatabase 创建新数据库
func (a *App) CreateDatabase(config database.DatabaseConfig, options database.CreateDatabaseOptions) error {
	factory := database.NewDBFactory()
	adapter, err := factory.CreateAdapter(config)
	if err != nil {
		return fmt.Errorf("创建数据库适配器失败: %v", err)
	}
	defer adapter.Close()

	if err := adapter.CreateDatabase(options.Name, options.Charset, options.Collation); err != nil {
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	return nil
}

// GetDatabaseCharsets 获取数据库支持的字符集
func (a *App) GetDatabaseCharsets(config database.DatabaseConfig) ([]database.CharsetInfo, error) {
	factory := database.NewDBFactory()
	return factory.GetDatabaseCharsets(config)
}

// TestConnection 测试数据库连接
func (a *App) TestConnection(config database.DatabaseConfig) error {
	factory := database.NewDBFactory()
	adapter, err := factory.CreateAdapter(config)
	if err != nil {
		return fmt.Errorf("创建数据库适配器失败: %v", err)
	}
	defer adapter.Close()

	if err := adapter.Connect(); err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	return adapter.Ping()
}

// CreateConnection 创建数据库连接
func (a *App) CreateConnection(config database.DatabaseConfig) (string, error) {
	factory := database.NewDBFactory()
	adapter, err := factory.CreateAdapter(config)
	if err != nil {
		return "", fmt.Errorf("创建数据库适配器失败: %v", err)
	}
	defer adapter.Close()

	if err := adapter.Connect(); err != nil {
		return "", fmt.Errorf("连接数据库失败: %v", err)
	}

	return "连接成功", nil
}

// GetDatabases 获取数据库列表
func (a *App) GetDatabases(config database.DatabaseConfig) ([]database.DatabaseInfo, error) {
	factory := database.NewDBFactory()
	adapter, err := factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("创建数据库适配器失败: %v", err)
	}
	defer adapter.Close()

	if err := adapter.Connect(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	list,err :=adapter.GetDatabases()
	if err!=nil{
		return nil,err
	}
	if list==nil{
		list=make([]database.DatabaseInfo,0)
	}
	return list,nil
}

// GetSchemas 获取数据库架构
func (a *App) GetSchemas(config database.DatabaseConfig, dbName string) ([]database.SchemaInfo, error) {
	factory := database.NewDBFactory()
	adapter, err := factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("创建数据库适配器失败: %v", err)
	}
	defer adapter.Close()

	if err := adapter.Connect(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	
	list,err :=adapter.GetSchemas(dbName)
	if err!=nil{
		return nil,err
	}
	if list==nil{
		list=make([]database.SchemaInfo,0)
	}
	return list,nil
}

// GetTables 获取表列表
func (a *App) GetTables(config database.DatabaseConfig, dbName, schema string) ([]database.TableInfo, error) {
	factory := database.NewDBFactory()
	adapter, err := factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("创建数据库适配器失败: %v", err)
	}
	defer adapter.Close()

	if err := adapter.Connect(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	list,err :=adapter.GetTables(dbName, schema)
	if err!=nil{
		return nil,err
	}
	if list==nil{
		list=make([]database.TableInfo,0)
	}
	return list,nil
}

// GetTableStructure 获取表结构
func (a *App) GetTableStructure(config database.DatabaseConfig, dbName, tableName string) ([]database.ColumnInfo, error) {
	factory := database.NewDBFactory()
	adapter, err := factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("创建数据库适配器失败: %v", err)
	}
	defer adapter.Close()

	if err := adapter.Connect(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	list,err := adapter.GetTableColumns(dbName, tableName)
	if err!=nil{
		return nil,err
	}
	if list==nil{
		list=make([]database.ColumnInfo,0)
	}
	return list,nil
}

// GetTableData 获取表数据
func (a *App) GetTableData(config database.DatabaseConfig, dbName, tableName string, offset, limit int) ([]map[string]string, error) {
	factory := database.NewDBFactory()
	adapter, err := factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("创建数据库适配器失败: %v", err)
	}
	defer adapter.Close()

	if err := adapter.Connect(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	list,err:= adapter.QueryTableData(dbName, tableName, offset, limit)
	if err!=nil{
		return nil,err
	}
	if list==nil{
		list=make([]map[string]string,0)
	}
	return list,nil
}

// GetTableRowCount 获取表行数
func (a *App) GetTableRowCount(config database.DatabaseConfig, dbName, tableName string) (int64, error) {
	factory := database.NewDBFactory()
	adapter, err := factory.CreateAdapter(config)
	if err != nil {
		return 0, fmt.Errorf("创建数据库适配器失败: %v", err)
	}
	defer adapter.Close()

	if err := adapter.Connect(); err != nil {
		return 0, fmt.Errorf("连接数据库失败: %v", err)
	}

	return adapter.GetTableRowCount(dbName, tableName)
}

// ExecuteQuery 执行SQL查询
func (a *App) ExecuteQuery(config database.DatabaseConfig, dbName, sql string) ([]map[string]string, error) {
	factory := database.NewDBFactory()
	adapter, err := factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("创建数据库适配器失败: %v", err)
	}
	defer adapter.Close()

	if err := adapter.Connect(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	return adapter.ExecuteQuery(dbName, sql)
}
