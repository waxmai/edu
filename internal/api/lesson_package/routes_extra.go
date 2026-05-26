package lesson_package

import "edu-schedule-system/internal/pkg/core"

// RegisterBusinessRoutes registers plural lesson-package aliases for frontend compatibility.
func RegisterBusinessRoutes(h *Handler, readGroup, writeGroup core.RouterGroup) {
	readGroup.GET("/lesson-packages", h.List())
	writeGroup.POST("/lesson-packages", h.Create())
	readGroup.GET("/lesson-packages/:id", h.GetByID())
	writeGroup.PUT("/lesson-packages/:id", h.UpdateByID())
	writeGroup.DELETE("/lesson-packages/:id", h.DeleteByID())
}
