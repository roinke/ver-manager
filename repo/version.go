package repo

import (
	"database/sql"
	"fmt"
	"strings"

	"ver-manager/db"
	"ver-manager/model"
)

// CreateVersion 创建新版本记录
func CreateVersion(v *model.Version) (*model.Version, error) {
	result, err := db.DB.Exec(
		`INSERT INTO versions
		 (branch_id, product_name, version_number, description, release_notes,
		  build_time, commit_hash, artifact_url, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.BranchID, v.ProductName, v.VersionNumber,
		v.Description, v.ReleaseNotes,
		v.BuildTime, v.CommitHash, v.ArtifactURL, v.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("创建版本失败: %w", err)
	}

	id, _ := result.LastInsertId()
	return GetVersionByID(id)
}

// GetVersionByID 根据 ID 获取版本详情
func GetVersionByID(id int64) (*model.Version, error) {
	v := &model.Version{}
	err := db.DB.QueryRow(
		`SELECT v.id, v.branch_id, v.product_name, v.version_number,
		        v.description, v.release_notes, v.build_time,
		        v.commit_hash, v.artifact_url, v.status, v.created_at,
		        b.name
		 FROM versions v
		 JOIN branches b ON b.id = v.branch_id
		 WHERE v.id = ?`, id,
	).Scan(&v.ID, &v.BranchID, &v.ProductName, &v.VersionNumber,
		&v.Description, &v.ReleaseNotes, &v.BuildTime,
		&v.CommitHash, &v.ArtifactURL, &v.Status, &v.CreatedAt,
		&v.BranchName)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("版本不存在: id=%d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("查询版本失败: %w", err)
	}
	return v, nil
}

// ListVersions 根据条件查询版本列表
func ListVersions(q model.VersionQuery) ([]model.Version, error) {
	var conditions []string
	var args []interface{}

	if q.BranchID != nil {
		conditions = append(conditions, "v.branch_id = ?")
		args = append(args, *q.BranchID)
	}
	if q.ProductName != "" {
		conditions = append(conditions, "v.product_name = ?")
		args = append(args, q.ProductName)
	}
	if q.Status != "" {
		conditions = append(conditions, "v.status = ?")
		args = append(args, q.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(
		`SELECT v.id, v.branch_id, v.product_name, v.version_number,
		        v.description, v.release_notes, v.build_time,
		        v.commit_hash, v.artifact_url, v.status, v.created_at,
		        b.name
		 FROM versions v
		 JOIN branches b ON b.id = v.branch_id
		 %s
		 ORDER BY v.build_time DESC`, whereClause)

	if q.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", q.Limit)
	}
	if q.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", q.Offset)
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询版本列表失败: %w", err)
	}
	defer rows.Close()

	return scanVersions(rows)
}

// GetLatestByBranch 获取某个分支的最新版本
func GetLatestByBranch(branchID int64) (*model.Version, error) {
	v := &model.Version{}
	err := db.DB.QueryRow(
		`SELECT v.id, v.branch_id, v.product_name, v.version_number,
		        v.description, v.release_notes, v.build_time,
		        v.commit_hash, v.artifact_url, v.status, v.created_at,
		        b.name
		 FROM versions v
		 JOIN branches b ON b.id = v.branch_id
		 WHERE v.branch_id = ?
		 ORDER BY v.id DESC
		 LIMIT 1`, branchID,
	).Scan(&v.ID, &v.BranchID, &v.ProductName, &v.VersionNumber,
		&v.Description, &v.ReleaseNotes, &v.BuildTime,
		&v.CommitHash, &v.ArtifactURL, &v.Status, &v.CreatedAt,
		&v.BranchName)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询最新版本失败: %w", err)
	}
	return v, nil
}

// GetLatestVersions 获取所有分支的最新版本（每个分支一个）
func GetLatestVersions() ([]model.Version, error) {
	rows, err := db.DB.Query(
		`SELECT v.id, v.branch_id, v.product_name, v.version_number,
		        v.description, v.release_notes, v.build_time,
		        v.commit_hash, v.artifact_url, v.status, v.created_at,
		        b.name
		 FROM versions v
		 JOIN branches b ON b.id = v.branch_id
		 WHERE v.id IN (
		     SELECT MAX(id) FROM versions GROUP BY branch_id
		 )
		 ORDER BY b.name`,
	)
	if err != nil {
		return nil, fmt.Errorf("查询各分支最新版本失败: %w", err)
	}
	defer rows.Close()

	return scanVersions(rows)
}

// GetLatestByProduct 获取某产品各分支的最新版本
func GetLatestByProduct(productName string) ([]model.Version, error) {
	rows, err := db.DB.Query(
		`SELECT v.id, v.branch_id, v.product_name, v.version_number,
		        v.description, v.release_notes, v.build_time,
		        v.commit_hash, v.artifact_url, v.status, v.created_at,
		        b.name
		 FROM versions v
		 JOIN branches b ON b.id = v.branch_id
		 WHERE v.product_name = ?
		   AND v.id IN (
		       SELECT MAX(id) FROM versions WHERE product_name = ? GROUP BY branch_id
		   )
		 ORDER BY b.name`, productName, productName,
	)
	if err != nil {
		return nil, fmt.Errorf("查询产品最新版本失败: %w", err)
	}
	defer rows.Close()

	return scanVersions(rows)
}

// UpdateVersionStatus 更新版本状态
func UpdateVersionStatus(id int64, status string) error {
	_, err := db.DB.Exec(
		`UPDATE versions SET status = ? WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("更新版本状态失败: %w", err)
	}
	return nil
}

// UpdateVersion 全字段更新版本
func UpdateVersion(id int64, v *model.Version) error {
	_, err := db.DB.Exec(
		`UPDATE versions SET
			branch_id = ?, product_name = ?, version_number = ?,
			description = ?, release_notes = ?,
			build_time = ?, commit_hash = ?, artifact_url = ?, status = ?
		 WHERE id = ?`,
		v.BranchID, v.ProductName, v.VersionNumber,
		v.Description, v.ReleaseNotes,
		v.BuildTime, v.CommitHash, v.ArtifactURL, v.Status,
		id,
	)
	if err != nil {
		return fmt.Errorf("更新版本失败: %w", err)
	}
	return nil
}

// CountVersions 统计版本数量（支持 branch_id / product_name / status 筛选）
func CountVersions(q model.VersionQuery) (int, error) {
	var conditions []string
	var args []interface{}

	if q.BranchID != nil {
		conditions = append(conditions, "v.branch_id = ?")
		args = append(args, *q.BranchID)
	}
	if q.ProductName != "" {
		conditions = append(conditions, "v.product_name = ?")
		args = append(args, q.ProductName)
	}
	if q.Status != "" {
		conditions = append(conditions, "v.status = ?")
		args = append(args, q.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM versions v %s", whereClause)
	var count int
	err := db.DB.QueryRow(query, args...).Scan(&count)
	return count, err
}

// scanVersions 扫描多行版本记录
func scanVersions(rows *sql.Rows) ([]model.Version, error) {
	var versions []model.Version
	for rows.Next() {
		var v model.Version
		if err := rows.Scan(&v.ID, &v.BranchID, &v.ProductName, &v.VersionNumber,
			&v.Description, &v.ReleaseNotes, &v.BuildTime,
			&v.CommitHash, &v.ArtifactURL, &v.Status, &v.CreatedAt,
			&v.BranchName); err != nil {
			return nil, fmt.Errorf("扫描版本行失败: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}
