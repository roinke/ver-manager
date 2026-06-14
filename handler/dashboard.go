package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ver-manager/model"
	"ver-manager/repo"
)

// GetDashboard 获取仪表盘统计数据
func GetDashboard(c *gin.Context) {
	branchCount, _ := repo.CountBranches(true)
	totalBranches, _ := repo.CountBranches(false)
	versionCount, _ := repo.CountVersions("")

	branches, _ := repo.ListBranches(false, 0, 0)
	recentVersions, _ := repo.ListVersions(model.VersionQuery{Limit: 10})

	c.JSON(http.StatusOK, gin.H{
		"branch_count":    branchCount,
		"total_branches":  totalBranches,
		"version_count":   versionCount,
		"branches":        branches,
		"recent_versions": recentVersions,
	})
}
