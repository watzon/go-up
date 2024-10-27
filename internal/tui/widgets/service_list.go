package widgets

import (
	"fmt"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/watzon/go-up/internal/types"
)

type ServiceList struct {
	*widgets.List
	selectedIndex  int
	pausedMonitors map[string]bool
}

func NewServiceList() *ServiceList {
	list := widgets.NewList()
	list.Title = "Monitors"
	list.TextStyle = termui.NewStyle(termui.ColorWhite)
	list.WrapText = false

	return &ServiceList{
		List:           list,
		selectedIndex:  0,
		pausedMonitors: make(map[string]bool),
	}
}

func (s *ServiceList) Update(services []types.Monitor, currentStatus map[string]types.ServiceStatus) {
	rows := make([]string, len(services))
	for i, service := range services {
		status, exists := currentStatus[service.Name]
		uptime := "Unknown"
		if exists {
			uptime = fmt.Sprintf("%d%%", int(status.Uptime24Hours))
		}

		statusIcon := s.getStatusIcon(service.Name, exists, status)
		title := s.formatTitle(service.Name, uptime)
		if i == s.selectedIndex {
			title = fmt.Sprintf("> %s %s", statusIcon, title)
		} else {
			title = fmt.Sprintf("  %s %s", statusIcon, title)
		}
		rows[i] = title
	}
	s.Rows = rows
}

func (s *ServiceList) getStatusIcon(serviceName string, exists bool, status types.ServiceStatus) string {
	if !exists {
		return "⚫" // Default gray
	}
	if s.pausedMonitors[serviceName] {
		return "⏸️"
	}
	if status.ResponseTime > 0 {
		return "🟢"
	}
	return "🔴"
}

func (s *ServiceList) formatTitle(name, uptime string) string {
	return fmt.Sprintf("%s (%s)", name, uptime)
}

func (s *ServiceList) TogglePause(serviceName string) {
	s.pausedMonitors[serviceName] = !s.pausedMonitors[serviceName]
}

func (s *ServiceList) GetSelectedIndex() int {
	return s.selectedIndex
}

func (s *ServiceList) MoveSelection(delta int, maxIndex int) {
	newIndex := s.selectedIndex + delta
	if newIndex >= 0 && newIndex < maxIndex {
		s.selectedIndex = newIndex
	}
}

// Add these methods to ServiceList

func (s *ServiceList) IsPaused(serviceName string) bool {
	return s.pausedMonitors[serviceName]
}

func (s *ServiceList) GetPausedMonitors() map[string]bool {
	return s.pausedMonitors
}
