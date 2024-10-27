package database

import (
	"time"
)

type Monitor struct {
	ID        uint           `gorm:"primaryKey"`
	URL       string         `gorm:"uniqueIndex;not null"`
	Name      string         `gorm:"not null"`
	IsActive  bool           `gorm:"default:true"`
	States    []MonitorState `gorm:"foreignKey:MonitorID"`
	Checks    []Check        `gorm:"foreignKey:MonitorID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MonitorState struct {
	ID        uint `gorm:"primaryKey"`
	MonitorID uint
	State     string    `gorm:"not null"`
	StartedAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Check struct {
	ID           uint `gorm:"primaryKey"`
	MonitorID    uint
	Timestamp    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	ResponseTime int
	IsUp         bool
	CertExpiry   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
