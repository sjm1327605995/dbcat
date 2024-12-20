package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteAdapter SQLite适配器
type SQLiteAdapter struct {
	config DatabaseConfig
	db     *sql.DB
}

// NewSQLiteAdapter 创建SQLite适配器
func NewSQLiteAdapter(config DatabaseConfig) *SQLiteAdapter {
	return &SQLiteAdapter{
		config: config,
	}
}

// Connect 连接数据库
func (a *SQLiteAdapter) Connect() error {
	// 如果已经有连接且连接有效，直接返回
	if a.db != nil {
		if err := a.db.Ping(); err == nil {
			return nil
		}
		// 如果连接无效，关闭它
		a.db.Close()
		a.db = nil
	}

	db, err := sql.Open("sqlite3", a.config.Database)
	if err != nil {
		return err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(1) // SQLite 只支持一个连接
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err = db.Ping(); err != nil {
		db.Close()
		return err
	}

	a.db = db
	return nil
}

// Close 关闭连接
func (a *SQLiteAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// GetDatabases 获取所有数据库（SQLite只有一个数据库文件）
func (a *SQLiteAdapter) GetDatabases() ([]DatabaseInfo, error) {
	return []DatabaseInfo{{Name: "main"}}, nil
}

// GetSchemas 获取指定数据库的所有schema（SQLite没有schema概念）
func (a *SQLiteAdapter) GetSchemas(dbName string) ([]SchemaInfo, error) {
	return []SchemaInfo{{Name: "main"}}, nil
}

// GetTables 获取指定schema的所有表
func (a *SQLiteAdapter) GetTables(dbName, schema string) ([]TableInfo, error) {
	query := `
		SELECT 
			name,
			'' as comment
		FROM 
			sqlite_master 
		WHERE 
			type='table' 
		AND 
			name NOT LIKE 'sqlite_%'
	`
	rows, err := a.db.Query(query)
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

// 添加获取表结构的方法
func (a *SQLiteAdapter) GetTableColumns(dbName, tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var cid int
		var name, type_ string
		var notnull, pk int
		var dflt_value interface{}

		err := rows.Scan(&cid, &name, &type_, &notnull, &dflt_value, &pk)
		if err != nil {
			return nil, err
		}

		col := ColumnInfo{
			Name:      name,
			Type:      type_,
			Nullable:  notnull == 0,
			IsPrimary: pk == 1,
		}

		// 解析类型中的长度信息
		if matches := regexp.MustCompile(`(\w+)\((\d+)\)`).FindStringSubmatch(type_); len(matches) == 3 {
			col.Type = matches[1]
			if length, err := strconv.Atoi(matches[2]); err == nil {
				col.Length = length
			}
		}

		columns = append(columns, col)
	}

	return columns, nil
}

// 添加获取表行数的方法
func (a *SQLiteAdapter) GetTableRowCount(dbName, tableName string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)

	var count int64
	err := a.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// 修改查询表数据的方法
func (a *SQLiteAdapter) QueryTableData(dbName, tableName string, offset, limit int) ([]map[string]string, error) {
	// 构建查询语句
	query := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", tableName)

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

	// 准备扫描目标
	values := make([]interface{}, len(colNames))
	valuePtrs := make([]interface{}, len(colNames))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// 读取数据
	var result []map[string]string
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		// 转换为 map
		entry := make(map[string]string)
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

	return result, nil
}

// Ping 测试连接是否有效
func (a *SQLiteAdapter) Ping() error {
	if a.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return a.db.Ping()
}
