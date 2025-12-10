package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"time"
)

type QuestionType string
type Difficulty string

const (
	SingleChoice   QuestionType = "single_choice"
	MultipleChoice QuestionType = "multiple_choice"
	Programming    QuestionType = "programming"
)

const (
	Easy   Difficulty = "easy"
	Medium Difficulty = "medium"
	Hard   Difficulty = "hard"
)

type Question struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	Type       QuestionType   `json:"type" gorm:"type:varchar(20)"`
	Content    string         `json:"content" gorm:"type:text"`
	Options    JSON           `json:"options" gorm:"type:text"` // JSON格式存储选项
	Answer     string         `json:"answer" gorm:"type:text"`
	Difficulty Difficulty     `json:"difficulty" gorm:"type:varchar(10)"`
	Language   string         `json:"language" gorm:"type:varchar(20)"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

type JSON map[string]interface{}

func (j JSON) GormDataType() string {
	return "text"
}

// Value 实现 driver.Valuer 接口，用于将 JSON 类型转换为数据库值
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口，用于从数据库读取值到 JSON 类型
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan non-string value into JSON")
	}

	if len(bytes) == 0 {
		*j = make(JSON)
		return nil
	}

	return json.Unmarshal(bytes, j)
}