package database

import (
	"fmt"
	"regexp"
	"strconv"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteAdapter SQLite适配器
type SQLiteAdapter struct {
	BaseAdapter
}

// NewSQLiteAdapter 创建SQLite适配器
func NewSQLiteAdapter(config DatabaseConfig) *SQLiteAdapter {
	adapter := &SQLiteAdapter{
		BaseAdapter: BaseAdapter{
			config: config,
		},
	}
	adapter.SetConnectFunc(adapter.Connect)
	return adapter
}

// Connect 连接数据库
func (a *SQLiteAdapter) Connect() error {
	// SQLite 直接使用文件路径
	db, err := sqlx.Connect("sqlite3", a.config.Database)
	if err != nil {
		return err
	}

	a.db = db
	// SQLite 特殊设置：只允许一个连接
	a.db.SetMaxOpenConns(1)
	a.SetupConnPool()
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
	db, err := a.DB()
	if err != nil {
		return nil, err
	}

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
	rows, err := db.Queryx(query)
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
func (a *SQLiteAdapter) GetTableColumns(dbName, tableName string) ([]ColumnInfo, error) {
	db, err := a.DB()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Queryx(query)
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

// GetTableRowCount 获取表行数
func (a *SQLiteAdapter) GetTableRowCount(dbName, tableName string) (int64, error) {
	db, err := a.DB()
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	var count int64
	err = db.QueryRowx(query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// QueryTableData 查询表数据
func (a *SQLiteAdapter) QueryTableData(dbName, tableName string, offset, limit int) ([]map[string]string, error) {
	db, err := a.DB()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", tableName)
	rows, err := db.Queryx(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]string
	for rows.Next() {
		row := make(map[string]interface{})
		err := rows.MapScan(row)
		if err != nil {
			return nil, err
		}

		strRow := make(map[string]string)
		for k, v := range row {
			if v == nil {
				strRow[k] = ""
			} else {
				switch v := v.(type) {
				case []byte:
					strRow[k] = string(v)
				default:
					strRow[k] = fmt.Sprintf("%v", v)
				}
			}
		}
		result = append(result, strRow)
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

// CreateDatabase SQLite不需要创建数据库
func (a *SQLiteAdapter) CreateDatabase(name string, charset string, collation string) error {
	return nil
}

// GetCharsets SQLite只支持UTF-8
func (a *SQLiteAdapter) GetCharsets() ([]CharsetInfo, error) {
	return []CharsetInfo{
		{
			Name:        "UTF-8",
			Description: "Unicode (UTF-8)",
			Collations:  []string{"BINARY", "NOCASE", "RTRIM"},
		},
	}, nil
}

// ExecuteQuery 执行SQL查询
func (a *SQLiteAdapter) ExecuteQuery(dbName, sql string) ([]map[string]string, error) {
	db, err := a.DB()
	if err != nil {
		return nil, err
	}

	// 执行查询
	rows, err := db.Queryx(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]string
	for rows.Next() {
		row := make(map[string]interface{})
		err := rows.MapScan(row)
		if err != nil {
			return nil, err
		}

		// 转换为字符串map
		strRow := make(map[string]string)
		for k, v := range row {
			if v == nil {
				strRow[k] = ""
			} else {
				switch v := v.(type) {
				case []byte:
					strRow[k] = string(v)
				default:
					strRow[k] = fmt.Sprintf("%v", v)
				}
			}
		}
		result = append(result, strRow)
	}

	return result, nil
}
