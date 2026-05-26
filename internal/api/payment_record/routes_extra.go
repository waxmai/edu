package payment_record

import "edu-schedule-system/internal/pkg/core"

// RegisterBusinessRoutes registers plural routes and business extensions for frontend compatibility.
func RegisterBusinessRoutes(h *Handler, readGroup, writeGroup core.RouterGroup) {
	readGroup.GET("/payment-records", h.ListBusiness())
	readGroup.GET("/payment-records/:id", h.GetByID())
	readGroup.GET("/students/:id/payment-records", h.ListBusiness())
	readGroup.GET("/statistics/income", h.IncomeStatistics())

	writeGroup.POST("/payment-records", h.Create())
	writeGroup.PUT("/payment-records/:id", h.UpdateByID())
	writeGroup.DELETE("/payment-records/:id", h.DeleteByID())
}
