package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresAdapter PostgreSQL适配器
type PostgresAdapter struct {
	config DatabaseConfig
	db     *sql.DB
}

// NewPostgresAdapter 创建PostgreSQL适配器
func NewPostgresAdapter(config DatabaseConfig) *PostgresAdapter {
	return &PostgresAdapter{
		config: config,
	}
}

// Connect 连接数据库
func (a *PostgresAdapter) Connect() error {
	// 如果已经有连接且连接有效，直接返回
	if a.db != nil {
		if err := a.db.Ping(); err == nil {
			return nil
		}
		// 如果连接无效，关闭它
		a.db.Close()
		a.db = nil
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		a.config.Host,
		a.config.Port,
		a.config.User,
		a.config.Password,
		a.config.Database,
		a.config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)           // 最大打开连接数
	db.SetMaxIdleConns(5)            // 最大空闲连接数
	db.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	if err = db.Ping(); err != nil {
		db.Close()
		return err
	}

	a.db = db
	return nil
}

// Close 关闭连接
func (a *PostgresAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// GetDatabases 获取所有数据库
func (a *PostgresAdapter) GetDatabases() ([]DatabaseInfo, error) {
	query := `
		SELECT datname 
		FROM pg_database 
		WHERE datistemplate = false 
		AND datname NOT IN ('postgres', 'template0', 'template1')
	`
	rows, err := a.db.Query(query)
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

// GetSchemas 获取指定数据库的所有schema
func (a *PostgresAdapter) GetSchemas(dbName string) ([]SchemaInfo, error) {
	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema')
	`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []SchemaInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		schemas = append(schemas, SchemaInfo{Name: name})
	}
	return schemas, nil
}

// GetTables 获取指定schema的所有表
func (a *PostgresAdapter) GetTables(dbName, schema string) ([]TableInfo, error) {
	query := `
		SELECT 
			table_name,
			obj_description((quote_ident(table_schema) || '.' || quote_ident(table_name))::regclass, 'pg_class') as table_comment
		FROM 
			information_schema.tables 
		WHERE 
			table_schema = $1 
		AND 
			table_type = 'BASE TABLE'
	`
	rows, err := a.db.Query(query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		var comment sql.NullString
		if err := rows.Scan(&table.Name, &comment); err != nil {
			return nil, err
		}
		table.Comment = comment.String
		tables = append(tables, table)
	}
	return tables, nil
}

// GetTableColumns 获取指定表的所有列
func (a *PostgresAdapter) GetTableColumns(dbName, tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			character_maximum_length,
			is_nullable,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT ku.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku
				ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY'
				AND tc.table_name = $1
		) pk ON c.column_name = pk.column_name
		WHERE c.table_name = $1
		ORDER BY ordinal_position;
	`

	rows, err := a.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var length sql.NullInt64
		var nullable string

		err := rows.Scan(&col.Name, &col.Type, &length, &nullable, &col.IsPrimary)
		if err != nil {
			return nil, err
		}

		col.Length = int(length.Int64)
		col.Nullable = nullable == "YES"

		columns = append(columns, col)
	}

	return columns, nil
}

// GetTableRowCount 获取指定表的行数
func (a *PostgresAdapter) GetTableRowCount(dbName, tableName string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)

	var count int64
	err := a.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// QueryTableData 查询指定表的数据
func (a *PostgresAdapter) QueryTableData(dbName, tableName string, offset, limit int) ([]map[string]string, error) {
	// 构建查询语句
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", tableName)

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
func (a *PostgresAdapter) Ping() error {
	if a.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return a.db.Ping()
}
