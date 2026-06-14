package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"ver-manager/model"
	"ver-manager/repo"
)

// ListBranches GET /api/branches?page=1&page_size=20
func ListBranches(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	branches, err := repo.ListBranches(false, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	total, _ := repo.CountBranches(false)

	c.JSON(http.StatusOK, gin.H{
		"data":      branches,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetBranch GET /api/branches/:id
func GetBranch(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "无效的分支ID"})
		return
	}

	branch, err := repo.GetBranchByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "分支不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": branch})
}

// CreateBranch POST /api/branches
func CreateBranch(c *gin.Context) {
	var body struct {
		Name           string `json:"name" binding:"required"`
		ParentBranchID *int64 `json:"parent_branch_id"`
		BranchType     string `json:"branch_type"`
		Description    string `json:"description"`
		PulledAt       string `json:"pulled_at"` // ISO 时间字符串
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误: " + err.Error()})
		return
	}
	if body.BranchType == "" {
		body.BranchType = "custom"
	}
	if body.ParentBranchID != nil && *body.ParentBranchID == 0 {
		body.ParentBranchID = nil
	}

	var pulledAt *model.DateTime
	if body.PulledAt != "" {
		t, err := parseTime(body.PulledAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "pulled_at 时间格式错误"})
			return
		}
		pulledAt = &t
	}

	branch, err := repo.CreateBranch(body.Name, body.ParentBranchID, body.BranchType, body.Description, pulledAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": branch})
}

// UpdateBranch PUT /api/branches/:id
func UpdateBranch(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "无效的分支ID"})
		return
	}

	branch, err := repo.GetBranchByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "分支不存在"})
		return
	}

	var body struct {
		Name           string `json:"name"`
		ParentBranchID *int64 `json:"parent_branch_id"`
		BranchType     string `json:"branch_type"`
		Description    string `json:"description"`
		IsActive       *bool  `json:"is_active"`
		PulledAt       string `json:"pulled_at"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}

	if body.Name != "" {
		branch.Name = body.Name
	}
	if body.ParentBranchID != nil {
		if *body.ParentBranchID == 0 {
			branch.ParentBranchID = nil
		} else {
			branch.ParentBranchID = body.ParentBranchID
		}
	}
	if body.BranchType != "" {
		branch.BranchType = body.BranchType
	}
	if body.Description != "" {
		branch.Description = body.Description
	}
	if body.IsActive != nil {
		branch.IsActive = *body.IsActive
	}
	if body.PulledAt != "" {
		t, err := parseTime(body.PulledAt)
		if err == nil {
			branch.PulledAt = &t
		}
	}

	if err := repo.UpdateBranch(branch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": branch})
}

// DeleteBranch DELETE /api/branches/:id (停用而非物理删除)
func DeleteBranch(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "无效的分支ID"})
		return
	}

	if err := repo.DeactivateBranch(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

// parseTime 尝试多种格式解析时间字符串
func parseTime(s string) (model.DateTime, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return model.DateTime(t), nil
		}
	}
	return model.DateTime{}, fmt.Errorf("无法解析时间: %s", s)
}
