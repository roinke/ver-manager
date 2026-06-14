package main

import (
	"embed"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"ver-manager/db"
	"ver-manager/handler"
)

//go:embed frontend/dist
var distFS embed.FS

// spaFS 是去掉了 frontend/dist 前缀后的前端文件系统
var spaFS fs.FS

func main() {
	if err := db.Init(""); err != nil {
		fmt.Fprintf(os.Stderr, "数据库初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	initDefaultData()

	// 准备前端文件系统
	var err error
	spaFS, err = fs.Sub(distFS, "frontend/dist")
	if err != nil {
		fmt.Fprintf(os.Stderr, "前端资源加载失败: %v\n", err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ============ API ============
	api := r.Group("/api")
	{
		api.GET("/dashboard", handler.GetDashboard)
		api.GET("/branches", handler.ListBranches)
		api.POST("/branches", handler.CreateBranch)
		api.GET("/branches/:id", handler.GetBranch)
		api.PUT("/branches/:id", handler.UpdateBranch)
		api.DELETE("/branches/:id", handler.DeleteBranch)
		api.GET("/versions", handler.ListVersions)
		api.GET("/versions/latest", handler.GetLatestVersions)
		api.POST("/versions", handler.CreateVersion)
		api.GET("/versions/:id", handler.GetVersion)
		api.PUT("/versions/:id", handler.UpdateVersion)
	}

	// ============ SPA 前端 ============
	r.Use(serveSPA)

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	fmt.Println("🚀 VerMan 已启动: http://localhost:" + port)
	if err := r.Run(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "服务启动失败: %v\n", err)
		os.Exit(1)
	}
}

// serveSPA 中间件：处理所有非 API 请求，返回前端静态文件
func serveSPA(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/api") {
		c.Next()
		return
	}
	if spaFS == nil {
		c.Next()
		return
	}

	// 规范化路径
	urlPath := c.Request.URL.Path
	if urlPath == "/" {
		urlPath = "/index.html"
	}
	filePath := strings.TrimPrefix(urlPath, "/")

	// 尝试读取文件
	data, err := fs.ReadFile(spaFS, filePath)
	if err != nil {
		// SPA fallback：所有路由返回 index.html
		data, err = fs.ReadFile(spaFS, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "前端资源加载失败")
			c.Abort()
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		c.Abort()
		return
	}

	// 根据扩展名设置 Content-Type
	ext := filepath.Ext(filePath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Data(http.StatusOK, contentType, data)
	c.Abort()
}

func initDefaultData() {
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM branches").Scan(&count)
	if count == 0 {
		db.DB.Exec(`INSERT INTO branches (name, branch_type, description, is_active)
		             VALUES ('master', 'main', '默认主分支', 1)`)
		fmt.Println("✅ 已创建默认 master 分支")
	}
}
