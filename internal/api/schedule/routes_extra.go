package schedule

import "edu-schedule-system/internal/pkg/core"

// RegisterBusinessRoutes registers plural routes and schedule business actions for frontend compatibility.
func RegisterBusinessRoutes(h *Handler, readGroup, writeGroup core.RouterGroup) {
	writeGroup.POST("/schedules", h.Create())
	readGroup.GET("/schedules/:id", h.GetByID())
	writeGroup.PUT("/schedules/:id", h.UpdateByID())
	writeGroup.DELETE("/schedules/:id", h.DeleteByID())
	writeGroup.PATCH("/schedules/:id/cancel", h.Cancel())
	writeGroup.POST("/schedules/:id/leave", h.Leave())
	writeGroup.POST("/schedules/:id/reschedule", h.Reschedule())
	writeGroup.POST("/makeup-schedules", h.CreateMakeup())
}
