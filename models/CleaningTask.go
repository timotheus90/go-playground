package models

import (
	"gorm.io/gorm"
	"time"
)

type CleaningTask struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Description string         `json:"description"`
	DueDate     time.Time      `json:"dueDate"`
	Assignee    string         `json:"assignee"`
	Completed   bool           `json:"completed"`
	Category    TaskCategory   `json:"category"`
}

type TaskCategory string

const (
	CategoryKitchen TaskCategory = "kitchen"
	CategoryBaths   TaskCategory = "baths"
	CategoryFloors  TaskCategory = "floors"
	CategoryOther   TaskCategory = "other"
)
