package reschedule_record

import "edu-schedule-system/internal/pkg/core"

// RegisterBusinessRoutes registers plural record aliases for frontend compatibility.
func RegisterBusinessRoutes(h *Handler, readGroup, writeGroup core.RouterGroup) {
	readGroup.GET("/reschedule-records", h.ListBusiness())
	readGroup.GET("/reschedule-records/:id", h.GetByID())
	writeGroup.POST("/reschedule-records", h.Create())
	writeGroup.PUT("/reschedule-records/:id", h.UpdateByID())
	writeGroup.DELETE("/reschedule-records/:id", h.DeleteByID())
}
