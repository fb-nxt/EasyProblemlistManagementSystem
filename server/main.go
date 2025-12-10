package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"homework-server/models"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	Difficulty models.Difficulty   `json:"difficulty" binding:"required"`
	Language   string              `json:"language" binding:"required"`
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

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		log.Printf("Add question error: %v", err)
		return
	}

	// 验证必填字段
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "题目内容不能为空"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误: " + err.Error()})
		log.Printf("Database error: %v", err)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		log.Printf("AI generate request error: %v", err)
		return
	}

	// 验证必填字段
	if req.Language == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "编程语言不能为空"})
		return
	}

	// 构建题型描述
	typeDesc := ""
	switch req.Type {
	case models.SingleChoice:
		typeDesc = "单选题"
	case models.MultipleChoice:
		typeDesc = "多选题"
	case models.Programming:
		typeDesc = "编程题"
	default:
		typeDesc = string(req.Type)
	}

	// 构建难度描述
	difficultyDesc := ""
	switch req.Difficulty {
	case models.Easy:
		difficultyDesc = "简单"
	case models.Medium:
		difficultyDesc = "中等"
	case models.Hard:
		difficultyDesc = "困难"
	default:
		difficultyDesc = string(req.Difficulty)
	}

	// 构建详细的AI提示词
	prompt := fmt.Sprintf(`请生成%d道%s类型的题目，难度为%s，编程语言为%s。`,
		req.Count, typeDesc, difficultyDesc, req.Language)

	if req.Topic != "" {
		prompt += fmt.Sprintf("题目主题或关键词：%s。", req.Topic)
	}

	// 根据题型添加具体要求
	switch req.Type {
case models.SingleChoice, models.MultipleChoice:
		prompt += `每道题目需要包含：
1. content: 题目内容
2. options: 选项对象，格式为 {"A": "选项A内容", "B": "选项B内容", "C": "选项C内容", "D": "选项D内容"}
3. answer: 答案，单选题为单个选项字母（如"A"），多选题为多个选项字母用逗号分隔（如"A,B"）`
	case models.Programming:
		prompt += `每道题目需要包含：
1. content: 编程题目描述
2. answer: 参考答案或解题思路`
	}

	prompt += `。请严格按照JSON数组格式返回，每道题目为一个对象，包含type、content、options（如果是选择题）、answer、difficulty、language字段。只返回JSON数组，不要包含其他文字说明。`

	// 调用AI API - 按照Java示例的格式
	aiReq := map[string]interface{}{
		"model": "deepseek-ai/DeepSeek-V3",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"stream":     false,
		"max_tokens": 2000, // 增加token数量以支持多道题目
		"stop":       []string{"null"},
	}

	resp, err := aiClient.R().
		SetBody(aiReq).
		Post("chat/completions")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI服务调用失败: " + err.Error()})
		log.Printf("AI API call error: %v", err)
		return
	}

	// 检查HTTP状态码
	if resp.StatusCode() != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("AI服务返回错误，状态码: %d, 响应: %s", resp.StatusCode(), string(resp.Body())),
		})
		log.Printf("AI API returned non-200 status: %d, body: %s", resp.StatusCode(), string(resp.Body()))
		return
	}

	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(resp.Body(), &aiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析AI响应失败: " + err.Error()})
		log.Printf("Failed to parse AI response: %v, body: %s", err, string(resp.Body()))
		return
	}

	// 检查是否有错误
	if aiResp.Error.Message != "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI服务错误: " + aiResp.Error.Message})
		return
	}

	if len(aiResp.Choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI未返回任何内容"})
		log.Printf("AI response has no choices, body: %s", string(resp.Body()))
		return
	}

	// 解析AI返回的题目
	var generatedQuestions []QuestionRequest
	content := aiResp.Choices[0].Message.Content

	// 清理内容，移除可能的markdown代码块标记
	content = cleanJSONContent(content)

	// 尝试解析AI返回的JSON内容
	var questionsFromAI []QuestionRequest
	if err := json.Unmarshal([]byte(content), &questionsFromAI); err == nil {
		// 成功解析为数组
		generatedQuestions = questionsFromAI
		log.Printf("Successfully parsed %d questions from AI", len(generatedQuestions))
	} else {
		// 如果不是数组，尝试解析为单个对象
		var singleQuestion QuestionRequest
		if err := json.Unmarshal([]byte(content), &singleQuestion); err == nil {
			generatedQuestions = []QuestionRequest{singleQuestion}
			log.Printf("Successfully parsed single question from AI")
		} else {
			// 如果解析失败，记录错误并返回
			log.Printf("Failed to parse AI response as JSON: %v", err)
			log.Printf("AI response content: %s", content)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "AI返回的内容格式不正确，无法解析为JSON。请重试。",
			})
			return
		}
	}

	// 补充和完善题目信息
	for i := range generatedQuestions {
		// 确保每个题目都有必要的字段
		if generatedQuestions[i].Type == "" {
			generatedQuestions[i].Type = req.Type
		}
		if generatedQuestions[i].Difficulty == "" {
			generatedQuestions[i].Difficulty = req.Difficulty
		}
		if generatedQuestions[i].Language == "" {
			generatedQuestions[i].Language = req.Language
		}

		// 如果是选择题但没有options，添加默认选项
		if (generatedQuestions[i].Type == models.SingleChoice || generatedQuestions[i].Type == models.MultipleChoice) &&
			generatedQuestions[i].Options == nil {
			generatedQuestions[i].Options = models.JSON{
				"A": "选项A",
				"B": "选项B",
				"C": "选项C",
				"D": "选项D",
			}
		}
	}

	// 确保返回的题目数量符合要求
	if len(generatedQuestions) > req.Count {
		generatedQuestions = generatedQuestions[:req.Count]
	} else if len(generatedQuestions) < req.Count {
		log.Printf("Warning: AI只生成了%d道题目，请求的是%d道", len(generatedQuestions), req.Count)
	}

	c.JSON(http.StatusOK, gin.H{"data": generatedQuestions})
}

// cleanJSONContent 清理JSON内容，移除可能的markdown代码块标记
func cleanJSONContent(content string) string {
	// 移除前后的空白字符
	content = strings.TrimSpace(content)

	// 移除可能的markdown代码块标记
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}

	return strings.TrimSpace(content)
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
