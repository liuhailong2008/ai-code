package service

import (
	"database/sql"
	"fmt"
	"log"

	"pingmesh-controller/conf"

	_ "github.com/go-sql-driver/mysql"
)

// DB 封装 MySQL 数据库操作
type DB struct {
	conn *sql.DB
}

// NewDB 创建数据库连接
func NewDB(cfg conf.MySQLConfig) (*DB, error) {
	conn, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	log.Printf("MySQL connected: %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)
	return &DB{conn: conn}, nil
}

// Close 关闭数据库连接
func (d *DB) Close() error {
	return d.conn.Close()
}

// ── IDC 机房信息（类型定义，供 handler 层使用）──

// IDC 机房信息
type IDC struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Enable bool   `json:"enable"`
}

// IDCLink 机房间链路
type IDCLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Enable bool   `json:"enable"`
}

// ── Leaf 节点配置查询 ──

// LeafNode Leaf 交换机节点
type LeafNode struct {
	Name string `json:"name"`
	IDC  string `json:"idc"`
}

// GetLeafsByIDC 获取指定机房的 Leaf 节点列表
func (d *DB) GetLeafsByIDC(idcCode string) ([]LeafNode, error) {
	rows, err := d.conn.Query("SELECT name, idc_code FROM leaf_node WHERE idc_code = ? ORDER BY name", idcCode)
	if err != nil {
		return nil, fmt.Errorf("query leafs: %w", err)
	}
	defer rows.Close()

	var leafs []LeafNode
	for rows.Next() {
		var leaf LeafNode
		if err := rows.Scan(&leaf.Name, &leaf.IDC); err != nil {
			return nil, fmt.Errorf("scan leaf: %w", err)
		}
		leafs = append(leafs, leaf)
	}
	return leafs, rows.Err()
}
