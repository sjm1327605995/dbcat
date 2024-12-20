package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLAdapter MySQL适配器
type MySQLAdapter struct {
	config DatabaseConfig
	db     *sql.DB
}

// NewMySQLAdapter 创建MySQL适配器
func NewMySQLAdapter(config DatabaseConfig) *MySQLAdapter {
	return &MySQLAdapter{
		config: config,
	}
}

// Connect 连接数据库
func (a *MySQLAdapter) Connect() error {
	// 如果已经有连接且连接有效，直接返回
	if a.db != nil {
		if err := a.db.Ping(); err == nil {
			return nil
		}
		// 如果连接无效，关闭它
		a.db.Close()
		a.db = nil
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		a.config.User,
		a.config.Password,
		a.config.Host,
		a.config.Port,
		a.config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)                  // 最大连接数
	db.SetMaxIdleConns(5)                   // 最大空闲连接数
	db.SetConnMaxLifetime(time.Hour)        // 连接最大生命周期
	db.SetConnMaxIdleTime(30 * time.Minute) // 空闲连接最大生命周期

	if err := db.Ping(); err != nil {
		db.Close()
		return err
	}

	a.db = db
	return nil
}

// Close 关闭连接
func (a *MySQLAdapter) Close() error {
	if a.db != nil {
		err := a.db.Close()
		a.db = nil
		return err
	}
	return nil
}

// GetDatabases 获取所有数据库
func (a *MySQLAdapter) GetDatabases() ([]DatabaseInfo, error) {
	rows, err := a.db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []DatabaseInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		databases = append(databases, DatabaseInfo{Name: name})
	}

	return databases, nil
}

// GetSchemas 获取指定数据库的所有schema（MySQL中schema等同于database）
func (a *MySQLAdapter) GetSchemas(dbName string) ([]SchemaInfo, error) {
	// MySQL doesn't have schemas in the same way as PostgreSQL
	return nil, nil
}

// GetTables 获取指定schema的所有表
func (a *MySQLAdapter) GetTables(dbName, schema string) ([]TableInfo, error) {
	query := `
		SELECT 
			TABLE_NAME,
			IFNULL(TABLE_COMMENT, '') as TABLE_COMMENT
		FROM 
			INFORMATION_SCHEMA.TABLES 
		WHERE 
			TABLE_SCHEMA = ?
	`
	rows, err := a.db.Query(query, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.Name, &table.Comment); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// GetTableColumns 获取表结构
func (a *MySQLAdapter) GetTableColumns(dbName, tableName string) ([]ColumnInfo, error) {
	if a.db == nil {
		if err := a.Connect(); err != nil {
			return nil, err
		}
	}

	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			CHARACTER_MAXIMUM_LENGTH,
			IS_NULLABLE,
			COLUMN_KEY
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := a.db.Query(query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var length sql.NullInt64
		var nullable, key string

		err := rows.Scan(&col.Name, &col.Type, &length, &nullable, &key)
		if err != nil {
			return nil, err
		}

		col.Length = int(length.Int64)
		col.Nullable = nullable == "YES"
		col.IsPrimary = key == "PRI"

		columns = append(columns, col)
	}

	return columns, nil
}

// GetTableRowCount 获取表行数（使用近似值以提高性能）
func (a *MySQLAdapter) GetTableRowCount(dbName, tableName string) (int64, error) {
	if a.db == nil {
		if err := a.Connect(); err != nil {
			return 0, err
		}
	}

	// 先尝试从 information_schema 获取近似行数
	query := `
		SELECT table_rows 
		FROM information_schema.tables 
		WHERE table_schema = ? AND table_name = ?
	`
	var count int64
	err := a.db.QueryRow(query, dbName, tableName).Scan(&count)
	if err == nil && count > 0 {
		return count, nil
	}

	// 如果获取失败或行数为0，再使用 COUNT(*)
	query = fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", dbName, tableName)
	err = a.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// QueryTableData 查询表数据
func (a *MySQLAdapter) QueryTableData(dbName, tableName string, offset, limit int) ([]map[string]string, error) {
	if a.db == nil {
		if err := a.Connect(); err != nil {
			return nil, err
		}
	}

	// 使用 SQL_CALC_FOUND_ROWS 来优化分页查询
	query := fmt.Sprintf("SELECT * FROM `%s`.`%s` LIMIT ? OFFSET ?", dbName, tableName)

	rows, err := a.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 获取列名
	colNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 预分配内存
	result := make([]map[string]string, 0, limit)
	values := make([]interface{}, len(colNames))
	valuePtrs := make([]interface{}, len(colNames))

	// 创建一个可重用的缓冲区
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// 读取数据
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		// 转换为 map
		entry := make(map[string]string, len(colNames))
		for i, col := range colNames {
			var v string
			val := values[i]
			if val == nil {
				v = ""
			} else {
				switch val := val.(type) {
				case []byte:
					v = string(val)
				default:
					v = fmt.Sprintf("%v", val)
				}
			}
			entry[col] = v
		}
		result = append(result, entry)
	}

	// 检查是否有错误发生
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// Ping 测试连接是否有效
func (a *MySQLAdapter) Ping() error {
	if a.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return a.db.Ping()
}
