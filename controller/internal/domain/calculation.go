package domain

import (
	"gorm.io/gorm"
	"time"
)

type Calculation struct {
	gorm.Model
	Owner       string
	Expression  string
	Result      *float64
	CompletedAt *time.Time
}
