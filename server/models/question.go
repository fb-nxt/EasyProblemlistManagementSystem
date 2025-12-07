package models

import (
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