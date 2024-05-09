package main

import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"strings"
	"time"
)

var (
	_DB     *sql.DB
	_config *Config
)

type Config struct {
	Dsn          string // 数据库连接信息
	ExportMDPath string // 导出的md的文件夹路径
}

func init() {
	initConfig()
	initDatabase()
}

// 读取配置文件
func initConfig() {
	_config = new(Config)
	if _, err := toml.DecodeFile("./config.toml", &_config); err != nil {
		panic(fmt.Errorf("load config file error: %v", err))
	}
}

// 数据库连接
func initDatabase() {
	DB, _ := sql.Open("mysql", _config.Dsn)
	//设置数据库最大连接数
	DB.SetConnMaxLifetime(100)
	//设置上数据库最大闲置连接数
	DB.SetMaxIdleConns(10)
	//设置超时时间
	DB.SetConnMaxIdleTime(10)
	//验证连接
	if err := DB.Ping(); err != nil {
		panic(fmt.Errorf("database connect error: %v", err))
	}
	_DB = DB
}

func main() {
	// 计算耗时
	defer TrackTime(time.Now())

	// 获取数据库表
	tables, err := GetDatabaseData(_config.Dsn)
	if err != nil {
		panic(err)
	}

	// 生成markdown表格
	if err = genMarkDownTable(tables); err != nil {
		panic(err)
	}
}

// 生成带有列的表格的Markdown
func markdownTable(t *Table, md *strings.Builder) {
	// 添加带有表名和可用表注释的表格标题
	md.WriteString(fmt.Sprintf("## %s\n", t.TableName))
	if t.TableComment != "" {
		md.WriteString(fmt.Sprintf("> %s\n\n", t.TableComment))
	}

	// 为列定义定义表头
	md.WriteString("| Column Name | Data Type | Length | Column Key | Column Comment |\n")
	md.WriteString("|-------------|-----------|--------|------------|----------------|\n")

	// 将每个列定义添加到表格中
	for _, column := range t.Columns {
		// 这里很可能描述内容中包含了换行，所以需要替换掉
		column.ColumnComment = strings.ReplaceAll(column.ColumnComment, "\n", " ")
		// 内容中也可能包含了|,需要替换掉
		column.ColumnComment = strings.ReplaceAll(column.ColumnComment, "|", "\\|")
		md.WriteString(fmt.Sprintf("| %s | %s | %d | %s | %s |\n",
			column.ColumnName, column.DataType, column.ColumnLen, column.ColumnKey, column.ColumnComment))
	}
}

// 生成markdown表格
func genMarkDownTable(tables []*Table) error {
	if len(tables) == 0 {
		return fmt.Errorf("no tables found")
	}
	// 获取Markdown的文件夹
	if _config.ExportMDPath == "" {
		var err error
		_config.ExportMDPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %w", err)
		}
	}
	// 确保目录存在
	if err := os.MkdirAll(_config.ExportMDPath, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}
	// 文件命名为yyyymmdd.md
	databaseName := GetDataBaseName(_config.Dsn)
	filePath := fmt.Sprintf("%s/%s-%s.md", _config.ExportMDPath, databaseName, time.Now().Format("20060102"))
	// 判断文件是否存在
	if _, err := os.Stat(filePath); err == nil {
		// 清空文件内容
		if err = os.Truncate(filePath, 0); err != nil {
			return fmt.Errorf("error truncating file: %w", err)
		}
	}
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	mdContent := new(strings.Builder)
	// 为每个表生成Markdown表格
	for _, table := range tables {
		markdownTable(table, mdContent)
		// 增加一行分割线
		mdContent.WriteString("\n---\n\n")
	}
	// 将Markdown内容写入文件
	if _, err = file.WriteString(mdContent.String()); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}
