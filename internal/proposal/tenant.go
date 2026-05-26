package proposal

const (
	OrganizationStatusActive   = "active"
	OrganizationStatusInactive = "inactive"

	SubscriptionStatusTrial     = "trial"
	SubscriptionStatusActive    = "active"
	SubscriptionStatusPastDue   = "past_due"
	SubscriptionStatusSuspended = "suspended"
	SubscriptionStatusExpired   = "expired"
)

func IsActiveSubscriptionStatus(status string) bool {
	switch status {
	case SubscriptionStatusTrial, SubscriptionStatusActive:
		return true
	default:
		return false
	}
}
