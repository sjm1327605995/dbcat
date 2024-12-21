package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

type MySQLAdapter struct {
	BaseAdapter
}

func NewMySQLAdapter(config DatabaseConfig) *MySQLAdapter {
	adapter := &MySQLAdapter{
		BaseAdapter: BaseAdapter{
			config: config,
		},
	}
	adapter.SetConnectFunc(adapter.Connect)
	return adapter
}

func (a *MySQLAdapter) Connect() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		a.config.User,
		a.config.Password,
		a.config.Host,
		a.config.Port,
	)

	if a.config.Database != "" {
		dsn += a.config.Database
	}

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return err
	}

	a.db = db
	a.SetupConnPool()
	return nil
}

// GetDatabases 获取所有数据库
func (a *MySQLAdapter) GetDatabases() ([]DatabaseInfo, error) {
	db, err := a.DB()
	if err != nil {
		return nil, err
	}

	query := "SHOW DATABASES"
	rows, err := db.Queryx(query)
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
	db, err := a.DB()
	if err != nil {
		return nil, err
	}

	query := `
		SELECT 
			TABLE_NAME,
			IFNULL(TABLE_COMMENT, '') as TABLE_COMMENT
		FROM 
			INFORMATION_SCHEMA.TABLES 
		WHERE 
			TABLE_SCHEMA = ?
	`
	rows, err := db.Queryx(query, dbName)
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

	rows, err := a.db.Queryx(query, dbName, tableName)
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

// GetTableRowCount 获取表行数
func (a *MySQLAdapter) GetTableRowCount(dbName, tableName string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", dbName, tableName)
	var count int64
	err := a.db.Get(&count, query)
	return count, err
}

// QueryTableData 查询表数据
func (a *MySQLAdapter) QueryTableData(dbName, tableName string, offset, limit int) ([]map[string]string, error) {
	query := fmt.Sprintf("SELECT * FROM `%s`.`%s` LIMIT ? OFFSET ?", dbName, tableName)
	rows, err := a.db.Queryx(query, limit, offset)
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

		// 转换字符串map
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

// CreateDatabase 创建数据库
func (a *MySQLAdapter) CreateDatabase(name string, charset string, collation string) error {
	// 确保连接到服务器而不是具体数据库
	oldDB := a.config.Database
	a.config.Database = ""
	defer func() { a.config.Database = oldDB }()

	db, err := a.DB()
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("CREATE DATABASE `%s`", name)
	if charset != "" {
		sql += fmt.Sprintf(" CHARACTER SET %s", charset)
	}
	if collation != "" {
		sql += fmt.Sprintf(" COLLATE %s", collation)
	}

	_, err = db.Exec(sql)
	return err
}

// GetCharsets 获取字符集
func (a *MySQLAdapter) GetCharsets() ([]CharsetInfo, error) {
	db, err := a.DB()
	if err != nil {
		return nil, err
	}

	query := `
		SELECT 
			c.CHARACTER_SET_NAME as name,
			c.DESCRIPTION as description,
			GROUP_CONCAT(co.COLLATION_NAME) as collations
		FROM information_schema.CHARACTER_SETS c
		JOIN information_schema.COLLATIONS co 
			ON co.CHARACTER_SET_NAME = c.CHARACTER_SET_NAME
		GROUP BY c.CHARACTER_SET_NAME, c.DESCRIPTION
		ORDER BY c.CHARACTER_SET_NAME
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
		return []CharsetInfo{
			{
				Name:        "utf8mb4",
				Description: "UTF-8 Unicode",
				Collations:  []string{"utf8mb4_general_ci", "utf8mb4_unicode_ci"},
			},
		}, nil
	}

	return charsets, nil
}

// ExecuteQuery 执行SQL查询
func (a *MySQLAdapter) ExecuteQuery(dbName, sql string) ([]map[string]string, error) {
	db, err := a.DB()
	if err != nil {
		return nil, err
	}

	// 如果指定了数据库，先切换到该数据库
	if dbName != "" {
		if _, err := db.Exec("USE " + dbName); err != nil {
			return nil, err
		}
	}

	// 清理SQL语句
	helper := &SQLHelper{}
	statements := helper.SplitStatements(sql)
	if len(statements) == 0 {
		return nil, fmt.Errorf("no valid SQL statements found")
	}

	// 执行每个语句，返回最后一个 SELECT 语句的结果
	var result []map[string]string
	for i, stmt := range statements {
		// 检查是否是 SELECT 语句
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(stmt)), "SELECT") {
			// 执行查询并获取结果
			rows, err := db.Queryx(stmt)
			if err != nil {
				return nil, fmt.Errorf("error executing statement %d: %v", i+1, err)
			}
			defer rows.Close()

			result = []map[string]string{} // 清空之前的结果
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
		} else {
			// 执行非 SELECT 语句
			if _, err := db.Exec(stmt); err != nil {
				return nil, fmt.Errorf("error executing statement %d: %v", i+1, err)
			}
		}
	}

	return result, nil
}

// ... 其他方法类似修改
