package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
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

// 临时结构体用于灵活解析AI返回的数据
type AIGeneratedQuestionTemp struct {
	Type       interface{} `json:"type"`
	Content    interface{} `json:"content"`
	Options    interface{} `json:"options"`
	Answer     interface{} `json:"answer"`
	Difficulty interface{} `json:"difficulty"`
	Language   interface{} `json:"language"`
	Topic      interface{} `json:"topic,omitempty"`
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

	// 构建详细的AI提示词
	prompt := buildAIPrompt(req)

	// 调用AI API
	aiReq := map[string]interface{}{
		"model": "deepseek-ai/DeepSeek-V3",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"stream":     false,
		"max_tokens": 2000,
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

	log.Printf("AI API raw response: %s\n", string(resp.Body()))

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

	// 获取AI返回的内容
	content := aiResp.Choices[0].Message.Content
	log.Printf("Raw AI response content: %s\n", content)

	// 解析AI返回的题目
	generatedQuestions, err := parseAIResponse(content, req)
	if err != nil {
		log.Printf("Failed to parse AI response: %v", err)
		log.Printf("Cleaned content: %s", cleanJSONContent(content))

		// 创建默认题目作为fallback
		generatedQuestions = createFallbackQuestions(req)
		log.Printf("Using fallback questions, count: %d", len(generatedQuestions))
	}

	// 确保返回的题目数量符合要求
	if len(generatedQuestions) > req.Count {
		generatedQuestions = generatedQuestions[:req.Count]
	} else if len(generatedQuestions) < req.Count {
		log.Printf("Warning: AI只生成了%d道题目，请求的是%d道", len(generatedQuestions), req.Count)
		// 补充默认题目
		for i := len(generatedQuestions); i < req.Count; i++ {
			generatedQuestions = append(generatedQuestions, createDefaultQuestion(req, i+1))
		}
	}

	// 最终验证
	for i, q := range generatedQuestions {
		if q.Type == "" || q.Content == "" {
			generatedQuestions[i] = createDefaultQuestion(req, i+1)
		}
	}

	log.Printf("Successfully generated %d questions", len(generatedQuestions))
	c.JSON(http.StatusOK, gin.H{"data": generatedQuestions})
}

// 构建AI提示词
func buildAIPrompt(req AIGenerateRequest) string {
	typeDesc := ""
	switch req.Type {
	case models.SingleChoice:
		typeDesc = "single_choice"
	case models.MultipleChoice:
		typeDesc = "multiple_choice"
	case models.Programming:
		typeDesc = "programming"
	default:
		typeDesc = string(req.Type)
	}

	difficultyDesc := ""
	switch req.Difficulty {
	case models.Easy:
		difficultyDesc = "easy"
	case models.Medium:
		difficultyDesc = "medium"
	case models.Hard:
		difficultyDesc = "hard"
	default:
		difficultyDesc = string(req.Difficulty)
	}

	prompt := fmt.Sprintf(`请生成%d道编程题目，要求如下：
1. 题目类型：%s
2. 难度：%s  
3. 编程语言：%s`,
		req.Count, typeDesc, difficultyDesc, req.Language)

	if req.Topic != "" {
		prompt += fmt.Sprintf("\n4. 主题：%s", req.Topic)
	}

	prompt += fmt.Sprintf(`

请严格按照以下JSON数组格式返回，不要包含任何其他文字：
[
  {
    "type": "%s",
    "content": "题目内容描述",
    "options": {"A": "选项A内容", "B": "选项B内容", "C": "选项C内容", "D": "选项D内容"},
    "answer": "答案",
    "difficulty": "%s", 
    "language": "%s"
  }
]

具体要求：
1. 如果是选择题(type为single_choice或multiple_choice)：
   - options字段必须是一个JSON对象，包含A、B、C、D四个选项
   - answer字段：单选题如"A"，多选题如"A,B"（用逗号分隔，不要有空格）
   
2. 如果是编程题(type为programming)：
   - options字段请返回空字符串
   - answer字段请返回空字符串""
   
3. type字段必须是：single_choice、multiple_choice或programming
4. difficulty字段必须是：easy、medium或hard
5. 只返回JSON数组，不要有其他内容`,
		typeDesc, difficultyDesc, req.Language)

	return prompt
}

// 解析AI返回的题目并转换为QuestionRequest
func parseAIResponse(content string, req AIGenerateRequest) ([]QuestionRequest, error) {
	content = cleanJSONContent(content)

	// 首先尝试直接解析为QuestionRequest数组
	var questions []QuestionRequest
	if err := json.Unmarshal([]byte(content), &questions); err == nil {
		return normalizeQuestions(questions, req), nil
	}

	// 如果失败，使用临时结构体解析
	var tempQuestions []AIGeneratedQuestionTemp
	if err := json.Unmarshal([]byte(content), &tempQuestions); err == nil {
		return convertTempToQuestions(tempQuestions, req), nil
	}

	// 尝试解析为单个对象
	var tempQuestion AIGeneratedQuestionTemp
	if err := json.Unmarshal([]byte(content), &tempQuestion); err == nil {
		question := convertSingleTempToQuestion(tempQuestion, req)
		return []QuestionRequest{question}, nil
	}

	// 如果还是失败，尝试解析为map数组
	var rawArray []json.RawMessage
	if err := json.Unmarshal([]byte(content), &rawArray); err == nil {
		var questions []QuestionRequest
		for _, raw := range rawArray {
			if question, err := parseRawQuestion(raw, req); err == nil {
				questions = append(questions, question)
			}
		}
		if len(questions) > 0 {
			return questions, nil
		}
	}

	return nil, fmt.Errorf("无法解析AI响应")
}

// 解析单个原始JSON题目
func parseRawQuestion(raw json.RawMessage, req AIGenerateRequest) (QuestionRequest, error) {
	var question QuestionRequest

	// 首先尝试直接解析
	if err := json.Unmarshal(raw, &question); err == nil {
		return normalizeQuestion(question, req), nil
	}

	// 如果失败，解析为map然后转换
	var tempMap map[string]interface{}
	if err := json.Unmarshal(raw, &tempMap); err != nil {
		return question, err
	}

	return convertMapToQuestion(tempMap, req), nil
}

// 转换map到QuestionRequest
func convertMapToQuestion(m map[string]interface{}, req AIGenerateRequest) QuestionRequest {
	var question QuestionRequest

	// 设置Type
	if typeVal, ok := m["type"]; ok {
		question.Type = validateAndConvertType(fmt.Sprintf("%v", typeVal))
	} else {
		question.Type = req.Type
	}

	// 设置Content
	if contentVal, ok := m["content"]; ok {
		switch v := contentVal.(type) {
		case string:
			question.Content = v
		default:
			question.Content = fmt.Sprintf("%v", v)
		}
	}

	// 设置Options（选择题）
	if question.Type == models.SingleChoice || question.Type == models.MultipleChoice {
		if optsVal, ok := m["options"]; ok {
			switch v := optsVal.(type) {
			case string:
				// 如果是字符串，尝试解析为JSON
				var optionsMap map[string]string
				if err := json.Unmarshal([]byte(v), &optionsMap); err == nil {
					question.Options = toModelsJSON(optionsMap)
				}
			case map[string]interface{}:
				// 如果是map，转换为models.JSON
				optionsMap := make(map[string]string)
				for key, val := range v {
					if strVal, ok := val.(string); ok {
						optionsMap[key] = strVal
					} else {
						optionsMap[key] = fmt.Sprintf("%v", val)
					}
				}
				question.Options = toModelsJSON(optionsMap)
			default:
				// 使用默认选项
				question.Options = toModelsJSON(map[string]string{
					"A": "选项A",
					"B": "选项B",
					"C": "选项C",
					"D": "选项D",
				})
			}
		} else {
			// 没有options字段，使用默认
			question.Options = toModelsJSON(map[string]string{
				"A": "选项A",
				"B": "选项B",
				"C": "选项C",
				"D": "选项D",
			})
		}
	}

	// 设置Answer
	if ansVal, ok := m["answer"]; ok {
		switch v := ansVal.(type) {
		case string:
			question.Answer = v
		case []interface{}:
			// 多选题答案数组
			var answers []string
			for _, item := range v {
				if str, ok := item.(string); ok {
					answers = append(answers, str)
				} else {
					answers = append(answers, fmt.Sprintf("%v", item))
				}
			}
			question.Answer = strings.Join(answers, ",")
		case []string:
			question.Answer = strings.Join(v, ",")
		case float64:
			question.Answer = fmt.Sprintf("%.0f", v)
		default:
			question.Answer = fmt.Sprintf("%v", v)
		}
	}

	// 设置Difficulty
	if diffVal, ok := m["difficulty"]; ok {
		question.Difficulty = validateAndConvertDifficulty(fmt.Sprintf("%v", diffVal))
	} else {
		question.Difficulty = req.Difficulty
	}

	// 设置Language
	if langVal, ok := m["language"]; ok {
		switch v := langVal.(type) {
		case string:
			question.Language = v
		default:
			question.Language = req.Language
		}
	} else {
		question.Language = req.Language
	}

	return question
}

// 转换临时结构体数组到QuestionRequest数组
func convertTempToQuestions(tempQuestions []AIGeneratedQuestionTemp, req AIGenerateRequest) []QuestionRequest {
	var questions []QuestionRequest
	for _, temp := range tempQuestions {
		question := convertSingleTempToQuestion(temp, req)
		questions = append(questions, question)
	}
	return questions
}

// 转换单个临时结构体到QuestionRequest
func convertSingleTempToQuestion(temp AIGeneratedQuestionTemp, req AIGenerateRequest) QuestionRequest {
	var question QuestionRequest

	// 转换Type
	if temp.Type != nil {
		question.Type = validateAndConvertType(fmt.Sprintf("%v", temp.Type))
	} else {
		question.Type = req.Type
	}

	// 转换Content
	if temp.Content != nil {
		switch v := temp.Content.(type) {
		case string:
			question.Content = v
		default:
			question.Content = fmt.Sprintf("%v", v)
		}
	}

	// 转换Options（选择题）
	if question.Type == models.SingleChoice || question.Type == models.MultipleChoice {
		if temp.Options != nil {
			switch v := temp.Options.(type) {
			case string:
				// 如果是字符串，尝试解析为JSON
				var optionsMap map[string]string
				if err := json.Unmarshal([]byte(v), &optionsMap); err == nil {
					question.Options = toModelsJSON(optionsMap)
				}
			case map[string]interface{}:
				// 如果是map，转换为models.JSON
				optionsMap := make(map[string]string)
				for key, val := range v {
					if strVal, ok := val.(string); ok {
						optionsMap[key] = strVal
					} else {
						optionsMap[key] = fmt.Sprintf("%v", val)
					}
				}
				question.Options = toModelsJSON(optionsMap)
			default:
				// 使用默认选项
				question.Options = toModelsJSON(map[string]string{
					"A": "选项A",
					"B": "选项B",
					"C": "选项C",
					"D": "选项D",
				})
			}
		} else {
			// 没有options字段，使用默认
			question.Options = toModelsJSON(map[string]string{
				"A": "选项A",
				"B": "选项B",
				"C": "选项C",
				"D": "选项D",
			})
		}
	}

	// 转换Answer
	if temp.Answer != nil {
		switch v := temp.Answer.(type) {
		case string:
			question.Answer = v
		case []interface{}:
			// 多选题答案数组
			var answers []string
			for _, item := range v {
				if str, ok := item.(string); ok {
					answers = append(answers, str)
				} else {
					answers = append(answers, fmt.Sprintf("%v", item))
				}
			}
			question.Answer = strings.Join(answers, ",")
		case []string:
			question.Answer = strings.Join(v, ",")
		case float64:
			question.Answer = fmt.Sprintf("%.0f", v)
		default:
			question.Answer = fmt.Sprintf("%v", v)
		}
	}

	// 转换Difficulty
	if temp.Difficulty != nil {
		question.Difficulty = validateAndConvertDifficulty(fmt.Sprintf("%v", temp.Difficulty))
	} else {
		question.Difficulty = req.Difficulty
	}

	// 转换Language
	if temp.Language != nil {
		switch v := temp.Language.(type) {
		case string:
			question.Language = v
		default:
			question.Language = req.Language
		}
	} else {
		question.Language = req.Language
	}

	return question
}

// 规范化题目数组
func normalizeQuestions(questions []QuestionRequest, req AIGenerateRequest) []QuestionRequest {
	for i := range questions {
		questions[i] = normalizeQuestion(questions[i], req)
	}
	return questions
}

// 规范化单个题目
func normalizeQuestion(question QuestionRequest, req AIGenerateRequest) QuestionRequest {
	// 确保Type不为空
	if question.Type == "" {
		question.Type = req.Type
	} else {
		question.Type = validateAndConvertType(string(question.Type))
	}

	// 确保Difficulty不为空
	if question.Difficulty == "" {
		question.Difficulty = req.Difficulty
	} else {
		question.Difficulty = validateAndConvertDifficulty(string(question.Difficulty))
	}

	// 确保Language不为空
	if question.Language == "" {
		question.Language = req.Language
	}

	// 处理选择题的Options
	if question.Type == models.SingleChoice || question.Type == models.MultipleChoice {
		// 确保Options不为nil
		// 注意：如果 models.JSON 不是引用类型，此处的 nil 检查可能需要根据具体类型调整
		if question.Options == nil {
			question.Options = toModelsJSON(map[string]string{
				"A": "选项A",
				"B": "选项B",
				"C": "选项C",
				"D": "选项D",
			})
		}

		// 清理Answer：去除空格，转换为大写
		if question.Answer != "" {
			question.Answer = strings.ToUpper(strings.ReplaceAll(question.Answer, " ", ""))
		}
	} else if question.Type == models.Programming {
		// 编程题的Answer应该为空
		question.Answer = ""
		// 编程题不需要Options
		question.Options = nil
	}

	return question
}

// 验证和转换类型
func validateAndConvertType(typeStr string) models.QuestionType {
	lowerStr := strings.ToLower(typeStr)
	switch {
	case strings.Contains(lowerStr, "single") || strings.Contains(lowerStr, "单选题"):
		return models.SingleChoice
	case strings.Contains(lowerStr, "multiple") || strings.Contains(lowerStr, "多选题"):
		return models.MultipleChoice
	case strings.Contains(lowerStr, "programming") || strings.Contains(lowerStr, "编程题"):
		return models.Programming
	default:
		return models.SingleChoice // 默认值
	}
}

// 验证和转换难度
func validateAndConvertDifficulty(diffStr string) models.Difficulty {
	lowerStr := strings.ToLower(diffStr)
	switch {
	case strings.Contains(lowerStr, "easy") || strings.Contains(lowerStr, "简单"):
		return models.Easy
	case strings.Contains(lowerStr, "medium") || strings.Contains(lowerStr, "中等"):
		return models.Medium
	case strings.Contains(lowerStr, "hard") || strings.Contains(lowerStr, "困难"):
		return models.Hard
	default:
		return models.Medium // 默认值
	}
}

// cleanJSONContent 清理JSON内容，移除可能的markdown代码块标记和修复格式
func cleanJSONContent(content string) string {
	// 移除前后的空白字符
	content = strings.TrimSpace(content)

	// 1. 移除所有可能的markdown代码块标记
	content = strings.ReplaceAll(content, "```json", "")
	content = strings.ReplaceAll(content, "```javascript", "")
	content = strings.ReplaceAll(content, "```", "")

	// 2. 查找并提取JSON部分
	// 查找第一个 [ 或 {
	startIndex := -1
	for i, ch := range content {
		if ch == '[' || ch == '{' {
			startIndex = i
			break
		}
	}

	if startIndex != -1 {
		content = content[startIndex:]

		// 找到匹配的结束符号
		bracketStack := []rune{}
		for i, ch := range content {
			if ch == '[' || ch == '{' {
				bracketStack = append(bracketStack, ch)
			} else if ch == ']' {
				if len(bracketStack) > 0 && bracketStack[len(bracketStack)-1] == '[' {
					bracketStack = bracketStack[:len(bracketStack)-1]
				}
			} else if ch == '}' {
				if len(bracketStack) > 0 && bracketStack[len(bracketStack)-1] == '{' {
					bracketStack = bracketStack[:len(bracketStack)-1]
				}
			}

			// 当栈为空时，表示找到了匹配的结束位置
			if len(bracketStack) == 0 {
				content = content[:i+1]
				break
			}
		}
	}

	// 3. 修复常见的JSON格式问题
	// 3.1 修复未加引号的键名
	re := regexp.MustCompile(`([{,]\s*)(\w+)(\s*:)`)
	content = re.ReplaceAllString(content, `${1}"${2}"${3}`)

	// 3.2 将单引号替换为双引号
	content = strings.ReplaceAll(content, `'`, `"`)

	// 3.3 修复布尔值和null值
	content = strings.ReplaceAll(content, `: true`, `: "true"`)
	content = strings.ReplaceAll(content, `: false`, `: "false"`)
	content = strings.ReplaceAll(content, `: null`, `: ""`)

	// 3.4 移除尾随逗号
	re = regexp.MustCompile(`,\s*([}\]])`)
	for re.MatchString(content) {
		content = re.ReplaceAllString(content, `$1`)
	}

	// 3.5 修复数字键名（options中的"A"等）
	re = regexp.MustCompile(`"(\d+)":`)
	content = re.ReplaceAllString(content, `"$1":`)

	return strings.TrimSpace(content)
}

// helper: 将 map[string]string 转换为 models.JSON（通过 marshal/unmarshal）
// 这样能兼容 models.JSON 是自定义 JSON 类型的情况。
func toModelsJSON(m map[string]string) models.JSON {
	var j models.JSON
	if m == nil {
		return j
	}
	b, err := json.Marshal(m)
	if err != nil {
		return j
	}
	if err := json.Unmarshal(b, &j); err != nil {
		return j
	}
	return j
}

// Fallback函数
func createFallbackQuestions(req AIGenerateRequest) []QuestionRequest {
	var questions []QuestionRequest
	for i := 0; i < req.Count; i++ {
		questions = append(questions, createDefaultQuestion(req, i+1))
	}
	return questions
}

func createDefaultQuestion(req AIGenerateRequest, index int) QuestionRequest {
	question := QuestionRequest{
		Type:       req.Type,
		Content:    fmt.Sprintf("请在此处填写第%d题题目内容", index),
		Difficulty: req.Difficulty,
		Language:   req.Language,
	}

	if req.Type == models.SingleChoice || req.Type == models.MultipleChoice {
		question.Options = toModelsJSON(map[string]string{
			"A": "选项A",
			"B": "选项B",
			"C": "选项C",
			"D": "选项D",
		})
		question.Answer = "A"
	} else if req.Type == models.Programming {
		question.Answer = ""
	}

	return question
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
