package lesson_record

import "edu-schedule-system/internal/pkg/core"

// RegisterBusinessRoutes registers plural lesson-record aliases for frontend compatibility.
func RegisterBusinessRoutes(h *Handler, readGroup, writeGroup core.RouterGroup) {
	readGroup.GET("/lesson-records", h.ListBusiness())
	readGroup.GET("/lesson-records/:id", h.GetByID())
	writeGroup.POST("/lesson-records", h.Create())
	writeGroup.PUT("/lesson-records/:id", h.UpdateByID())
	writeGroup.DELETE("/lesson-records/:id", h.DeleteByID())
}
