package student

import "edu-schedule-system/internal/pkg/core"

// RegisterBusinessRoutes exposes plural resource aliases for frontend compatibility.
func RegisterBusinessRoutes(h *Handler, readGroup, writeGroup core.RouterGroup) {
	readGroup.GET("/students/:id", h.GetByID())
	writeGroup.POST("/students", h.Create())
	writeGroup.PUT("/students/:id", h.UpdateByID())
	writeGroup.DELETE("/students/:id", h.DeleteByID())
}
