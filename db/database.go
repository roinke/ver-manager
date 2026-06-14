package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// DB 全局数据库连接
var DB *sql.DB

// Init 初始化 SQLite 数据库连接并创建表结构
func Init(dbPath string) error {
	if dbPath == "" {
		dbPath = os.Getenv("VERMAN_DB")
		if dbPath == "" {
			dbPath = "verman.db"
		}
	}

	dir := filepath.Dir(dbPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建数据库目录失败: %w", err)
		}
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	if err := migrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func Close() {
	if DB != nil {
		DB.Close()
	}
}

// migrate 创建表结构
func migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS branches (
		id               INTEGER PRIMARY KEY AUTOINCREMENT,
		name             TEXT    NOT NULL UNIQUE,
		parent_branch_id INTEGER REFERENCES branches(id),
		branch_type      TEXT    NOT NULL DEFAULT 'custom'
		                        CHECK(branch_type IN ('main','release','feature','hotfix','custom')),
		description      TEXT    DEFAULT '',
		is_active        INTEGER DEFAULT 1,
		pulled_at        TEXT    DEFAULT NULL,
		created_at       TEXT    DEFAULT (datetime('now','localtime')),
		updated_at       TEXT    DEFAULT (datetime('now','localtime'))
	);

	CREATE TABLE IF NOT EXISTS versions (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		branch_id       INTEGER NOT NULL REFERENCES branches(id),
		product_name    TEXT    NOT NULL,
		version_number  TEXT    NOT NULL,
		description     TEXT    DEFAULT '',
		release_notes   TEXT    DEFAULT '',
		build_time      TEXT    DEFAULT (datetime('now','localtime')),
		commit_hash     TEXT    DEFAULT '',
		artifact_url    TEXT    DEFAULT '',
		status          TEXT    DEFAULT 'draft'
		                        CHECK(status IN ('draft','released','deprecated','revoked')),
		created_at      TEXT    DEFAULT (datetime('now','localtime')),
		UNIQUE(branch_id, version_number)
	);

	CREATE INDEX IF NOT EXISTS idx_versions_product  ON versions(product_name);
	CREATE INDEX IF NOT EXISTS idx_versions_status   ON versions(status);
	CREATE INDEX IF NOT EXISTS idx_versions_time     ON versions(build_time);
	CREATE INDEX IF NOT EXISTS idx_branches_parent   ON branches(parent_branch_id);
	`

	// SQLite 的 DB.Exec 只执行第一条语句，需要逐条拆分执行
	for _, stmt := range strings.Split(schema, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("执行迁移语句失败: %w\n语句: %s", err, stmt)
		}
	}

	// 兼容性迁移：旧表（v1.0）可能缺少 pulled_at 列，忽略"已存在"错误
	DB.Exec(`ALTER TABLE branches ADD COLUMN pulled_at TEXT DEFAULT NULL`)

	return nil
}
