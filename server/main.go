package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"homework-server/models"
)

var db *gorm.DB
var aiClient *resty.Client

type Pagination struct {
	Page     int `json:"page" form:"page" binding:"min=1"`
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=100"`
}

type QuestionRequest struct {
	Type       models.QuestionType `json:"type" binding:"required"`
	Content    string              `json:"content" binding:"required"`
	Options    models.JSON         `json:"options"`
	Answer     string              `json:"answer"`
	Difficulty models.Difficulty   `json:"difficulty"`
	Language   string              `json:"language"`
}

type AIGenerateRequest struct {
	Type       models.QuestionType `json:"type" binding:"required"`
	Count      int                 `json:"count" binding:"min=1,max=20"`
	Difficulty models.Difficulty   `json:"difficulty"`
	Language   string              `json:"language"`
	Topic      string              `json:"topic"`
}

type AIResponse struct {
	Questions []QuestionRequest `json:"questions"`
}

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 初始化数据库
	var err error
	db, err = gorm.Open(sqlite.Open("questions.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// 自动迁移
	db.AutoMigrate(&models.Question{})

	// 初始化AI客户端
	aiClient = resty.New()
	aiClient.SetBaseURL("https://api.siliconflow.cn/v1/")
	aiClient.SetHeader("Authorization", "Bearer "+os.Getenv("AI_API_KEY"))
	aiClient.SetHeader("Content-Type", "application/json")

	// 初始化Gin
	r := gin.Default()

	// 配置静态文件服务
	r.Static("/assets", "./client/dist/assets")
	r.StaticFile("/", "./client/dist/index.html")
	r.StaticFile("/index.html", "./client/dist/index.html")

	// API路由组
	api := r.Group("/api")
	{
		// 1. 查询接口
		api.GET("/questions", getQuestions)
		
		// 2. 添加接口
		api.POST("/questions", addQuestion)
		
		// 3. 编辑接口
		api.PUT("/questions/:id", updateQuestion)
		
		// 4. 删除接口
		api.DELETE("/questions/:id", deleteQuestion)
		api.DELETE("/questions", batchDeleteQuestions)
		
		// 5. AI生成接口
		api.POST("/ai/generate", generateQuestions)
		
		// 6. 获取学习心得
		api.GET("/learning-note", getLearningNote)
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s...", port)
	r.Run(":" + port)
}

// 1. 查询接口
func getQuestions(c *gin.Context) {
	var pagination Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	offset := (pagination.Page - 1) * pagination.PageSize

	// 构建查询条件
	query := db.Model(&models.Question{})
	
	// 筛选条件
	if typeStr := c.Query("type"); typeStr != "" {
		query = query.Where("type = ?", typeStr)
	}
	if difficulty := c.Query("difficulty"); difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var questions []models.Question
	if err := query.Offset(offset).Limit(pagination.PageSize).Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  questions,
		"total": total,
		"page":  pagination.Page,
		"size":  pagination.PageSize,
	})
}

// 2. 添加接口
func addQuestion(c *gin.Context) {
	var req QuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	question := models.Question{
		Type:       req.Type,
		Content:    req.Content,
		Options:    req.Options,
		Answer:     req.Answer,
		Difficulty: req.Difficulty,
		Language:   req.Language,
	}

	if err := db.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": question})
}

// 3. 编辑接口
func updateQuestion(c *gin.Context) {
	id := c.Param("id")
	
	var req QuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var question models.Question
	if err := db.First(&question, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	question.Type = req.Type
	question.Content = req.Content
	question.Options = req.Options
	question.Answer = req.Answer
	question.Difficulty = req.Difficulty
	question.Language = req.Language

	if err := db.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": question})
}

// 4. 删除接口（单个）
func deleteQuestion(c *gin.Context) {
	id := c.Param("id")
	
	if err := db.Delete(&models.Question{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question deleted successfully"})
}

// 4.1 批量删除接口
func batchDeleteQuestions(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Delete(&models.Question{}, ids).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Questions deleted successfully"})
}

// 5. AI生成接口
func generateQuestions(c *gin.Context) {
	var req AIGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建AI请求
	prompt := fmt.Sprintf(`请生成%d道%s类型的题目，难度为%s`, 
		req.Count, req.Type, req.Difficulty)
	
	if req.Language != "" {
		prompt += fmt.Sprintf("，编程语言为%s", req.Language)
	}
	if req.Topic != "" {
		prompt += fmt.Sprintf("，主题为%s", req.Topic)
	}

	// 调用AI API
	aiReq := map[string]interface{}{
		"model": "Qwen/Qwen2.5-7B-Instruct",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt + "。请以JSON格式返回，包含content, options, answer字段",
			},
		},
		"max_tokens": 2000,
	}

	resp, err := aiClient.R().
		SetBody(aiReq).
		Post("chat/completions")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service error: " + err.Error()})
		return
	}

	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(resp.Body(), &aiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse AI response"})
		return
	}

	if len(aiResp.Choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No response from AI"})
		return
	}

	// 解析AI返回的题目
	var generatedQuestions []QuestionRequest
	//content := aiResp.Choices[0].Message.Content
	
	// 这里需要根据实际AI返回的格式进行解析
	// 简化处理：直接创建示例题目
	question := QuestionRequest{
		Type:       req.Type,
		Content:    "AI生成的题目内容示例",
		Difficulty: req.Difficulty,
		Language:   req.Language,
	}
	
	if req.Type == models.SingleChoice {
		question.Options = models.JSON{
			"A": "选项A",
			"B": "选项B",
			"C": "选项C",
			"D": "选项D",
		}
		question.Answer = "A"
	}
	
	generatedQuestions = append(generatedQuestions, question)

	c.JSON(http.StatusOK, gin.H{"data": generatedQuestions})
}

// 6. 获取学习心得接口
func getLearningNote(c *gin.Context) {
	// 读取学习心得文件
	content, err := os.ReadFile("../学习心得.md")
	if err != nil {
		// 如果文件不存在，返回默认内容
		content = []byte("# 学习心得\n\n这是我的学习心得内容...")
	}

	c.JSON(http.StatusOK, gin.H{"content": string(content)})
}