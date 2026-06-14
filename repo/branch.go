package repo

import (
	"database/sql"
	"fmt"

	"ver-manager/db"
	"ver-manager/model"
)

// CreateBranch 创建新分支
func CreateBranch(name string, parentBranchID *int64, branchType, description string, pulledAt *model.DateTime) (*model.Branch, error) {
	result, err := db.DB.Exec(
		`INSERT INTO branches (name, parent_branch_id, branch_type, description, pulled_at)
		 VALUES (?, ?, ?, ?, ?)`,
		name, parentBranchID, branchType, description, pulledAt,
	)
	if err != nil {
		return nil, fmt.Errorf("创建分支失败: %w", err)
	}

	id, _ := result.LastInsertId()
	return GetBranchByID(id)
}

// GetBranchByID 根据 ID 获取分支
func GetBranchByID(id int64) (*model.Branch, error) {
	branch := &model.Branch{}
	var parentID sql.NullInt64
	var isActive int

	err := db.DB.QueryRow(
		`SELECT id, name, parent_branch_id, branch_type, description, is_active,
		        pulled_at, created_at, updated_at
		 FROM branches WHERE id = ?`, id,
	).Scan(&branch.ID, &branch.Name, &parentID, &branch.BranchType,
		&branch.Description, &isActive, &branch.PulledAt, &branch.CreatedAt, &branch.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("分支不存在: id=%d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("查询分支失败: %w", err)
	}

	if parentID.Valid {
		branch.ParentBranchID = &parentID.Int64
	}
	branch.IsActive = isActive == 1
	return branch, nil
}

// GetBranchByName 根据名称获取分支
func GetBranchByName(name string) (*model.Branch, error) {
	branch := &model.Branch{}
	var parentID sql.NullInt64
	var isActive int

	err := db.DB.QueryRow(
		`SELECT id, name, parent_branch_id, branch_type, description, is_active,
		        pulled_at, created_at, updated_at
		 FROM branches WHERE name = ?`, name,
	).Scan(&branch.ID, &branch.Name, &parentID, &branch.BranchType,
		&branch.Description, &isActive, &branch.PulledAt, &branch.CreatedAt, &branch.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("分支不存在: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("查询分支失败: %w", err)
	}

	if parentID.Valid {
		branch.ParentBranchID = &parentID.Int64
	}
	branch.IsActive = isActive == 1
	return branch, nil
}

// ListBranches 列出分支，支持分页（limit=0 表示不限制）
func ListBranches(activeOnly bool, limit, offset int) ([]model.Branch, error) {
	query := `SELECT id, name, parent_branch_id, branch_type, description, is_active,
	                  pulled_at, created_at, updated_at
			FROM branches`
	if activeOnly {
		query += " WHERE is_active = 1"
	}
	query += " ORDER BY id"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询分支列表失败: %w", err)
	}
	defer rows.Close()

	return scanBranches(rows)
}

// GetChildrenBranches 获取某个分支的所有直接子分支
func GetChildrenBranches(parentID int64) ([]model.Branch, error) {
	rows, err := db.DB.Query(
		`SELECT id, name, parent_branch_id, branch_type, description, is_active,
		        pulled_at, created_at, updated_at
		 FROM branches WHERE parent_branch_id = ? ORDER BY id`, parentID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询子分支失败: %w", err)
	}
	defer rows.Close()

	return scanBranches(rows)
}

// UpdateBranch 更新分支信息（全字段）
func UpdateBranch(branch *model.Branch) error {
	activeVal := 0
	if branch.IsActive {
		activeVal = 1
	}

	_, err := db.DB.Exec(
		`UPDATE branches SET name = ?, parent_branch_id = ?, description = ?,
		 is_active = ?, branch_type = ?, pulled_at = ?,
		 updated_at = datetime('now','localtime')
		 WHERE id = ?`,
		branch.Name, branch.ParentBranchID, branch.Description,
		activeVal, branch.BranchType, branch.PulledAt, branch.ID,
	)
	if err != nil {
		return fmt.Errorf("更新分支失败: %w", err)
	}
	return nil
}

// CountBranches 统计分支数量
func CountBranches(activeOnly bool) (int, error) {
	query := "SELECT COUNT(*) FROM branches"
	if activeOnly {
		query += " WHERE is_active = 1"
	}
	var count int
	err := db.DB.QueryRow(query).Scan(&count)
	return count, err
}

// DeactivateBranch 停用分支
func DeactivateBranch(id int64) error {
	_, err := db.DB.Exec(
		`UPDATE branches SET is_active = 0, updated_at = datetime('now','localtime') WHERE id = ?`,
		id,
	)
	return err
}

// scanBranches 扫描多行分支记录
func scanBranches(rows *sql.Rows) ([]model.Branch, error) {
	var branches []model.Branch
	for rows.Next() {
		var b model.Branch
		var parentID sql.NullInt64
		var isActive int

		if err := rows.Scan(&b.ID, &b.Name, &parentID, &b.BranchType,
			&b.Description, &isActive, &b.PulledAt, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描分支行失败: %w", err)
		}

		if parentID.Valid {
			b.ParentBranchID = &parentID.Int64
		}
		b.IsActive = isActive == 1
		branches = append(branches, b)
	}
	return branches, rows.Err()
}
