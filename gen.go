package main

import (
	"fmt"
	"strings"
	"time"
)

// 用于转换mysql数据类型为golang数据类型
var mysqlType = map[string]string{
	"tinyint":    "int",
	"smallint":   "int",
	"mediumint":  "int",
	"int":        "int",
	"integer":    "int",
	"bigint":     "int64",
	"float":      "float64",
	"double":     "float64",
	"decimal":    "float64",
	"char":       "string",
	"varchar":    "string",
	"tinytext":   "string",
	"text":       "string",
	"mediumtext": "string",
	"longtext":   "string",
	"tinyblob":   "string",
	"blob":       "string",
	"mediumblob": "string",
	"longblob":   "string",
	"date":       "time.Time",
	"time":       "time.Time",
	"year":       "time.Time",
	"datetime":   "time.Time",
	"timestamp":  "time.Time",
}

// Table 数据库表结构体
type Table struct {
	TableName    string `json:"tableName"`
	TableComment string `json:"tableComment"`

	Columns []*Column `gorm:"-" json:"columns"`
}

// Column 数据库表字段结构体
type Column struct {
	ColumnName    string `json:"columnName"`
	ColumnKey     string `json:"columnKey"`
	ColumnLen     int    `json:"columnLen"`
	DataType      string `json:"dataType"`
	ColumnComment string `json:"columnComment"`

	ColType string `gorm:"-" json:"colType"` // 未用到，转换成golang的类型
}

// TrackTime 计算函数耗时
func TrackTime(pre time.Time) time.Duration {
	elapsed := time.Since(pre)
	fmt.Println(fmt.Sprintf("耗时: %v", elapsed))
	return elapsed
}

// GetDataBaseName 转换dsn为数据库名
// eg: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
func GetDataBaseName(dsn string) string {
	start := strings.Index(dsn, "/") + 1
	end := strings.Index(dsn, "?")
	if start > 0 && end > 0 && start < end {
		dsn = dsn[start:end]
	} else {
		dsn = ""
	}
	return dsn
}

// 获取数据库表
func getTable(dbName string) ([]*Table, error) {
	rows, err := _DB.Query("select table_name as TableName, table_comment as TableComment from information_schema.tables where table_schema = ?", dbName)
	if err != nil {
		return nil, fmt.Errorf("get tables failed, err: %v", err)
	}
	defer rows.Close()
	tables := make([]*Table, 0)
	for rows.Next() {
		var table Table
		if err = rows.Scan(&table.TableName, &table.TableComment); err != nil {
			return nil, fmt.Errorf("scan table failed, err: %v", err)
		}
		tables = append(tables, &table)
	}
	return tables, nil
}

// 获取数据库表字段
func getColumn(tableName string) ([]*Column, error) {
	rows, err := _DB.Query("select column_name as ColumnName, column_key as ColumnKey, IFNULL(character_maximum_length,0) as ColumnLen, data_type as DataType, column_comment as ColumnComment from information_schema.columns where table_name = ?", tableName)
	if err != nil {
		return nil, fmt.Errorf("get columns failed, err: %v", err)
	}
	defer rows.Close()
	columns := make([]*Column, 0)
	for rows.Next() {
		var column Column
		if err = rows.Scan(&column.ColumnName, &column.ColumnKey, &column.ColumnLen, &column.DataType, &column.ColumnComment); err != nil {
			return nil, fmt.Errorf("scan column failed, err: %v", err)
		}
		columns = append(columns, &column)
	}
	return columns, nil
}

func GetDatabaseData(dsn string) ([]*Table, error) {
	// 获取数据库名
	dbName := GetDataBaseName(dsn)
	if dbName == "" {
		return nil, fmt.Errorf("get database name failed")
	}
	// 获取数据库表
	tables, err := getTable(dbName)
	if err != nil {
		return nil, err
	}
	// 获取数据库表字段
	for i := range tables {
		tables[i].Columns, err = getColumn(tables[i].TableName)
		if err != nil {
			return nil, err
		}
		// 转换字段类型
		for j := range tables[i].Columns {
			tables[i].Columns[j].ColType = mysqlType[tables[i].Columns[j].DataType]
		}
	}
	return tables, nil
}
