package course

import "edu-schedule-system/internal/pkg/core"

// RegisterBusinessRoutes exposes plural resource aliases for frontend compatibility.
func RegisterBusinessRoutes(h *Handler, readGroup, writeGroup core.RouterGroup) {
	writeGroup.POST("/courses", h.Create())
	readGroup.GET("/courses/:id", h.GetByID())
	writeGroup.PUT("/courses/:id", h.UpdateByID())
	writeGroup.DELETE("/courses/:id", h.DeleteByID())
}
