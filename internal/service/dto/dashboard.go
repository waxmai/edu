package dto

type DashboardQuery struct {
	OrganizationID int32 `form:"organizationId" json:"organizationId"`
	CampusID       int32 `form:"campusId" json:"campusId"`
}

type DashboardSummaryCard struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Desc  string `json:"desc"`
}

type DashboardQuickAction struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Path  string `json:"path"`
}

type DashboardAlert struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Level string `json:"level"`
	Path  string `json:"path,omitempty"`
}

type DashboardMetricItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type DashboardPlatformRiskItem struct {
	OrganizationName string `json:"organizationName"`
	FollowUpStatus   string `json:"followUpStatus"`
	RemainingDays    int32  `json:"remainingDays"`
	UsedUsers        int64  `json:"usedUsers"`
	MaxUsers         int32  `json:"maxUsers"`
	UsedCampuses     int64  `json:"usedCampuses"`
	MaxCampuses      int32  `json:"maxCampuses"`
	HealthLevel      string `json:"healthLevel"`
}

type DashboardCampusRiskItem struct {
	CampusID         int32  `json:"campusId"`
	CampusName       string `json:"campusName"`
	OrganizationID   int32  `json:"organizationId"`
	OrganizationName string `json:"organizationName"`
	ActiveUsers      int64  `json:"activeUsers"`
	RiskCount        int32  `json:"riskCount"`
	LowLessonCount   int32  `json:"lowLessonCount"`
	ArrearsCount     int32  `json:"arrearsCount"`
	InactiveCount    int32  `json:"inactiveCount"`
	RescheduleCount  int32  `json:"rescheduleCount"`
}

type DashboardPaymentItem struct {
	ID                int32   `json:"id"`
	StudentName       string  `json:"studentName"`
	LessonPackageName string  `json:"lessonPackageName"`
	PaymentType       string  `json:"paymentType"`
	Amount            float64 `json:"amount"`
}

type DashboardScheduleItem struct {
	ID         int32  `json:"id"`
	CourseName string `json:"courseName"`
	ClassDate  string `json:"classDate"`
	StartTime  string `json:"startTime"`
	Status     string `json:"status"`
}

type DashboardLessonRecordItem struct {
	ID               int32  `json:"id"`
	LessonContent    string `json:"lessonContent"`
	AttendanceStatus string `json:"attendanceStatus"`
	RecordedAt       string `json:"recordedAt"`
}

type DashboardResponse struct {
	Role             string                      `json:"role"`
	SummaryCards     []DashboardSummaryCard      `json:"summaryCards"`
	Alerts           []DashboardAlert            `json:"alerts,omitempty"`
	QuickActions     []DashboardQuickAction      `json:"quickActions,omitempty"`
	Payments         []DashboardPaymentItem      `json:"payments,omitempty"`
	Schedules        []DashboardScheduleItem     `json:"schedules,omitempty"`
	LessonRecords    []DashboardLessonRecordItem `json:"lessonRecords,omitempty"`
	PlatformRisks    []DashboardPlatformRiskItem `json:"platformRisks,omitempty"`
	CampusRisks      []DashboardCampusRiskItem   `json:"campusRisks,omitempty"`
	MetricItems      []DashboardMetricItem       `json:"metricItems,omitempty"`
	SubscriptionDays int32                       `json:"subscriptionDays,omitempty"`
}
