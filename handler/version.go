package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"ver-manager/model"
	"ver-manager/repo"
)

// ListVersions GET /api/versions?branch_id=&product=&status=&page=1&page_size=20 (或 ?limit=999 兼容旧调用)
func ListVersions(c *gin.Context) {
	var limit, offset int

	// 优先用 page/page_size；否则用 limit（向后兼容）
	if _, hasPage := c.GetQuery("page"); hasPage {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}
		limit = pageSize
		offset = (page - 1) * pageSize
	} else {
		limit = 100
		if v := c.Query("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = n
			}
		}
	}

	var branchID *int64
	if v := c.Query("branch_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			branchID = &id
		}
	}

	versions, err := repo.ListVersions(model.VersionQuery{
		BranchID:    branchID,
		ProductName: c.Query("product"),
		Status:      c.Query("status"),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	total, _ := repo.CountVersions(model.VersionQuery{
		BranchID:    branchID,
		ProductName: c.Query("product"),
		Status:      c.Query("status"),
	})

	c.JSON(http.StatusOK, gin.H{
		"data":      versions,
		"total":     total,
		"page":      (offset / limit) + 1,
		"page_size": limit,
	})
}

// GetVersion GET /api/versions/:id
func GetVersion(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "无效的版本ID"})
		return
	}

	version, err := repo.GetVersionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "版本不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": version})
}

// CreateVersion POST /api/versions
func CreateVersion(c *gin.Context) {
	var body struct {
		BranchID      int64  `json:"branch_id" binding:"required"`
		ProductName   string `json:"product_name" binding:"required"`
		VersionNumber string `json:"version_number" binding:"required"`
		Description   string `json:"description"`
		ReleaseNotes  string `json:"release_notes"`
		BuildTime     string `json:"build_time"` // ISO 8601 or "2006-01-02 15:04:05"
		CommitHash    string `json:"commit_hash"`
		ArtifactURL   string `json:"artifact_url"`
		Status        string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误: " + err.Error()})
		return
	}
	if body.Status == "" {
		body.Status = "released"
	}

	buildTime := model.Now()
	if body.BuildTime != "" {
		if t, err := time.Parse("2006-01-02T15:04:05Z07:00", body.BuildTime); err == nil {
			buildTime = model.DateTime(t)
		} else if t, err := time.Parse("2006-01-02 15:04:05", body.BuildTime); err == nil {
			buildTime = model.DateTime(t)
		} else if t, err := time.Parse(time.RFC3339, body.BuildTime); err == nil {
			buildTime = model.DateTime(t)
		}
	}

	version, err := repo.CreateVersion(&model.Version{
		BranchID:      body.BranchID,
		ProductName:   body.ProductName,
		VersionNumber: body.VersionNumber,
		Description:   body.Description,
		ReleaseNotes:  body.ReleaseNotes,
		BuildTime:     buildTime,
		CommitHash:    body.CommitHash,
		ArtifactURL:   body.ArtifactURL,
		Status:        body.Status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": version})
}

// UpdateVersion PUT /api/versions/:id
func UpdateVersion(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "无效的版本ID"})
		return
	}

	version, err := repo.GetVersionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "版本不存在"})
		return
	}

	var body struct {
		BranchID      int64  `json:"branch_id"`
		ProductName   string `json:"product_name"`
		VersionNumber string `json:"version_number"`
		Description   string `json:"description"`
		ReleaseNotes  string `json:"release_notes"`
		BuildTime     string `json:"build_time"`
		CommitHash    string `json:"commit_hash"`
		ArtifactURL   string `json:"artifact_url"`
		Status        string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}

	if body.BranchID != 0 {
		version.BranchID = body.BranchID
	}
	if body.ProductName != "" {
		version.ProductName = body.ProductName
	}
	if body.VersionNumber != "" {
		version.VersionNumber = body.VersionNumber
	}
	if body.Description != "" {
		version.Description = body.Description
	}
	if body.ReleaseNotes != "" {
		version.ReleaseNotes = body.ReleaseNotes
	}
	if body.BuildTime != "" {
		if t, err := time.Parse("2006-01-02T15:04:05Z07:00", body.BuildTime); err == nil {
			version.BuildTime = model.DateTime(t)
		} else if t, err := time.Parse("2006-01-02 15:04:05", body.BuildTime); err == nil {
			version.BuildTime = model.DateTime(t)
		}
	}
	if body.CommitHash != "" {
		version.CommitHash = body.CommitHash
	}
	if body.ArtifactURL != "" {
		version.ArtifactURL = body.ArtifactURL
	}
	if body.Status != "" {
		version.Status = body.Status
	}

	if err := repo.UpdateVersion(id, version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	// 重新查询返回最新数据
	updated, _ := repo.GetVersionByID(id)
	c.JSON(http.StatusOK, gin.H{"data": updated})
}

// GetLatestVersions GET /api/versions/latest?product=
func GetLatestVersions(c *gin.Context) {
	product := c.Query("product")

	var versions []model.Version
	var err error

	if product != "" {
		versions, err = repo.GetLatestByProduct(product)
	} else {
		versions, err = repo.GetLatestVersions()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": versions})
}
