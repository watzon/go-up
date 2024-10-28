package database

import (
	"fmt"
	"time"

	"github.com/watzon/go-up/internal/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}

func (db *DB) Init() error {
	// Auto migrate the schema
	return db.AutoMigrate(&Monitor{}, &MonitorState{}, &Check{})
}

func (db *DB) AddMonitor(name, url string) error {
	monitor := Monitor{
		Name:     name,
		URL:      url,
		IsActive: true,
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&monitor).Error; err != nil {
			return err
		}

		state := MonitorState{
			MonitorID: monitor.ID,
			State:     "active",
			StartedAt: time.Now(),
		}

		return tx.Create(&state).Error
	})
}

func (db *DB) RemoveMonitor(name string) error {
	return db.Where("name = ?", name).Delete(&Monitor{}).Error
}

func (db *DB) PauseMonitor(name string) error {
	return db.Model(&Monitor{}).Where("name = ?", name).Update("is_active", false).Error
}

func (db *DB) ResumeMonitor(name string) error {
	return db.Model(&Monitor{}).Where("name = ?", name).Update("is_active", true).Error
}

func (db *DB) ListMonitors() ([]types.Monitor, error) {
	var dbMonitors []Monitor
	if err := db.Find(&dbMonitors).Error; err != nil {
		return nil, err
	}

	monitors := make([]types.Monitor, len(dbMonitors))
	for i, m := range dbMonitors {
		monitors[i] = types.Monitor{
			ID:       int(m.ID),
			Name:     m.Name,
			URL:      m.URL,
			IsActive: m.IsActive,
		}
	}

	return monitors, nil
}

func (db *DB) AddStats(monitorName string, responseTime int, isUp bool, certExpiry time.Time) error {
	var monitor Monitor
	if err := db.Where("name = ?", monitorName).First(&monitor).Error; err != nil {
		return err
	}

	check := Check{
		MonitorID:    monitor.ID,
		ResponseTime: responseTime,
		IsUp:         isUp,
		CertExpiry:   &certExpiry,
		Timestamp:    time.Now(),
	}

	return db.Create(&check).Error
}

func (db *DB) GetStats(monitorName string, duration time.Duration) (types.ServiceStatus, error) {
	var status types.ServiceStatus
	status.ServiceName = monitorName

	var monitor Monitor
	if err := db.Preload("States", func(db *gorm.DB) *gorm.DB {
		return db.Order("started_at DESC").Limit(1)
	}).Where("name = ?", monitorName).First(&monitor).Error; err != nil {
		return status, err
	}

	var stats struct {
		AvgResponseTime float64
		Uptime24h       float64
		Uptime30d       float64
	}

	dayAgo := time.Now().AddDate(0, 0, -1)
	monthAgo := time.Now().AddDate(0, 0, -30)

	err := db.Model(&Check{}).
		Where("monitor_id = ? AND timestamp >= ?", monitor.ID, monthAgo).
		Select("COALESCE(AVG(response_time), 0) as avg_response_time").
		Scan(&stats).Error

	if err != nil {
		return status, err
	}

	var upCount24h, totalCount24h int64
	err = db.Model(&Check{}).
		Where("monitor_id = ? AND timestamp >= ?", monitor.ID, dayAgo).
		Select("COUNT(CASE WHEN is_up THEN 1 END) as up_count, COUNT(*) as total_count").
		Row().Scan(&upCount24h, &totalCount24h)

	if err != nil {
		return status, err
	}

	if totalCount24h > 0 {
		stats.Uptime24h = float64(upCount24h) * 100.0 / float64(totalCount24h)
	}

	var upCount30d, totalCount30d int64
	err = db.Model(&Check{}).
		Where("monitor_id = ? AND timestamp >= ?", monitor.ID, monthAgo).
		Select("COUNT(CASE WHEN is_up THEN 1 END) as up_count, COUNT(*) as total_count").
		Row().Scan(&upCount30d, &totalCount30d)

	if err != nil {
		return status, err
	}

	if totalCount30d > 0 {
		stats.Uptime30d = float64(upCount30d) * 100.0 / float64(totalCount30d)
	}

	var lastCheck Check
	if err := db.Where("monitor_id = ?", monitor.ID).
		Order("timestamp DESC").
		First(&lastCheck).Error; err != nil && err != gorm.ErrRecordNotFound {
		return status, err
	}

	status.ServiceURL = monitor.URL
	status.IsActive = monitor.IsActive
	status.ResponseTime = lastCheck.ResponseTime
	status.CurrentStatus = lastCheck.IsUp
	status.AvgResponseTime = stats.AvgResponseTime
	status.Uptime24Hours = stats.Uptime24h
	status.Uptime30Days = stats.Uptime30d
	if lastCheck.CertExpiry != nil {
		status.CertificateExpiry = *lastCheck.CertExpiry
	}

	return status, nil
}

func (db *DB) GetHistoricalStats(monitorID int, count int) ([]types.HistoricalStat, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive, got %d", count)
	}

	var checks []Check
	if err := db.Where("monitor_id = ?", monitorID).
		Order("timestamp DESC").
		Limit(count).
		Find(&checks).Error; err != nil {
		return nil, err
	}

	stats := make([]types.HistoricalStat, len(checks))
	for i, check := range checks {
		stats[i] = types.HistoricalStat{
			ResponseTime: check.ResponseTime,
			IsUp:         check.IsUp,
			Timestamp:    check.Timestamp,
		}
	}

	return stats, nil
}
