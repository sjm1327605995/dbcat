package database

import (
	"fmt"
	"strings"
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresAdapter PostgreSQL适配器
type PostgresAdapter struct {
	BaseAdapter
}

// NewPostgresAdapter 创建PostgreSQL适配器
func NewPostgresAdapter(config DatabaseConfig) *PostgresAdapter {
	adapter := &PostgresAdapter{
		BaseAdapter: BaseAdapter{
			config: config,
		},
	}
	adapter.SetConnectFunc(adapter.Connect)
	return adapter
}

// Connect 连接数据库
func (a *PostgresAdapter) Connect() error {
	// 创建数据库时连接到默认的 postgres 数据库
	dbname := "postgres"
	if a.config.Database != "" {
		dbname = a.config.Database
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		a.config.Host,
		a.config.Port,
		a.config.User,
		a.config.Password,
		dbname,
		a.config.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return err
	}

	a.db = db
	a.SetupConnPool()
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
	rows, err := a.db.Queryx(query)
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
	rows, err := a.db.Queryx(query)
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
	rows, err := a.db.Queryx(query, schema)
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

	rows, err := a.db.Queryx(query, tableName)
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
	err := a.db.QueryRowx(query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// QueryTableData 查询指定表的数据
func (a *PostgresAdapter) QueryTableData(dbName, tableName string, offset, limit int) ([]map[string]string, error) {
	// 构建查询语句
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", tableName)

	rows, err := a.db.Queryx(query, limit, offset)
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

// CreateDatabase 创建数据库
func (a *PostgresAdapter) CreateDatabase(name string, charset string, collation string) error {
	// 确保连接到 postgres 数据库
	oldDB := a.config.Database
	a.config.Database = "postgres"
	defer func() { a.config.Database = oldDB }()

	db, err := a.DB()
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("CREATE DATABASE \"%s\"", name)
	if charset != "" {
		sql += fmt.Sprintf(" ENCODING '%s'", charset)
	}
	if collation != "" {
		sql += fmt.Sprintf(" LC_COLLATE '%s'", collation)
	}

	_, err = db.Exec(sql)
	return err
}

// GetCharsets 获取PostgreSQL支持的字符集和排序规则
func (a *PostgresAdapter) GetCharsets() ([]CharsetInfo, error) {
	db, err := a.DB()
	if err != nil {
		return []CharsetInfo{}, err
	}

	query := `
		SELECT 
			pg_encoding_to_char(encoding) as name,
			pg_encoding_to_char(encoding) as description,
			string_agg(collname, ',') as collations
		FROM pg_collation
		GROUP BY encoding
		ORDER BY name
	`
	
	rows, err := db.Queryx(query)
	if err != nil {
		return []CharsetInfo{}, err
	}
	defer rows.Close()

	charsets := []CharsetInfo{}
	for rows.Next() {
		var charset CharsetInfo
		var collationsStr string
		if err := rows.Scan(&charset.Name, &charset.Description, &collationsStr); err != nil {
			return []CharsetInfo{}, err
		}
		if collationsStr != "" {
			charset.Collations = strings.Split(collationsStr, ",")
		} else {
			charset.Collations = []string{}
		}
		charsets = append(charsets, charset)
	}

	if len(charsets) == 0 {
		// 如果没有找到字符集，返回默认值
		return []CharsetInfo{
			{
				Name:        "UTF8",
				Description: "Unicode UTF8",
				Collations:  []string{"default"},
			},
		}, nil
	}

	return charsets, nil
}

// ExecuteQuery 执行SQL查询
func (a *PostgresAdapter) ExecuteQuery(dbName, sql string) ([]map[string]string, error) {
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
