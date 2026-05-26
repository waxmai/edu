package audit

import (
	"context"
	"fmt"
	"strings"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql"

	"gorm.io/gorm"
)

const (
	defaultTargetSearchLimit = 20
	maxTargetSearchLimit     = 50
)

type TargetSearchQuery struct {
	Module  string `form:"module" json:"module"`
	Keyword string `form:"keyword" json:"keyword"`
	Limit   int    `form:"limit" json:"limit"`
}

type TargetOption struct {
	ID     int32  `json:"id"`
	Module string `json:"module"`
	Name   string `json:"name"`
	Label  string `json:"label"`
	Extra  string `json:"extra,omitempty"`
}

type TargetSearchResponse struct {
	Items []TargetOption `json:"items"`
}

type TargetSearchService interface {
	Search(ctx context.Context, actor proposal.SessionUserInfo, query TargetSearchQuery) (*TargetSearchResponse, error)
}

type targetSearchService struct {
	db mysql.Repo
}

func NewTargetSearchService(db mysql.Repo) TargetSearchService {
	return &targetSearchService{db: db}
}

func (s *targetSearchService) Search(ctx context.Context, _ proposal.SessionUserInfo, query TargetSearchQuery) (*TargetSearchResponse, error) {
	if s == nil || s.db == nil || s.db.GetDbR() == nil {
		return &TargetSearchResponse{Items: []TargetOption{}}, nil
	}
	limit := normalizeTargetSearchLimit(query.Limit)
	module := strings.TrimSpace(query.Module)
	keyword := strings.TrimSpace(query.Keyword)
	items := make([]TargetOption, 0, limit)
	db := s.db.GetDbR().WithContext(ctx)

	searchers := []struct {
		module string
		fn     func(*gorm.DB, string, int) ([]TargetOption, error)
	}{
		{module: "auth", fn: searchUserTargets},
		{module: "user", fn: searchUserTargets},
		{module: "tenant", fn: searchTenantTargets},
		{module: "student", fn: searchStudentTargets},
		{module: "course", fn: searchCourseTargets},
		{module: "lesson_package", fn: searchLessonPackageTargets},
		{module: "payment", fn: searchPaymentTargets},
		{module: "schedule", fn: searchScheduleTargets},
		{module: "reschedule", fn: searchRescheduleTargets},
	}

	for _, searcher := range searchers {
		if module != "" && searcher.module != module {
			continue
		}
		remaining := limit - len(items)
		if remaining <= 0 {
			break
		}
		found, err := searcher.fn(db, keyword, remaining)
		if err != nil {
			return nil, err
		}
		items = append(items, found...)
	}

	return &TargetSearchResponse{Items: items}, nil
}

func normalizeTargetSearchLimit(limit int) int {
	if limit <= 0 {
		return defaultTargetSearchLimit
	}
	if limit > maxTargetSearchLimit {
		return maxTargetSearchLimit
	}
	return limit
}

func likeKeyword(keyword string) string {
	return "%" + strings.ReplaceAll(keyword, "%", "\\%") + "%"
}

func applyIDOrLike(db *gorm.DB, keyword string, columns ...string) *gorm.DB {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return db
	}
	clauses := make([]string, 0, len(columns)+1)
	args := make([]interface{}, 0, len(columns)+1)
	clauses = append(clauses, "CAST(id AS CHAR) = ?")
	args = append(args, keyword)
	like := likeKeyword(keyword)
	for _, column := range columns {
		clauses = append(clauses, column+" LIKE ?")
		args = append(args, like)
	}
	return db.Where(strings.Join(clauses, " OR "), args...)
}

func searchUserTargets(db *gorm.DB, keyword string, limit int) ([]TargetOption, error) {
	type row struct {
		ID       int32
		Username string
		RealName string
		RoleCode string
	}
	var rows []row
	err := applyIDOrLike(db.Table("sys_user"), keyword, "username", "real_name", "phone", "email").
		Select("id, username, real_name, role_code").Order("id DESC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]TargetOption, 0, len(rows))
	for _, row := range rows {
		name := firstNonEmpty(row.RealName, row.Username)
		items = append(items, TargetOption{ID: row.ID, Module: "user", Name: name, Label: fmt.Sprintf("%s（用户 #%d）", name, row.ID), Extra: row.Username})
	}
	return items, nil
}

func searchTenantTargets(db *gorm.DB, keyword string, limit int) ([]TargetOption, error) {
	type orgRow struct {
		ID               int32
		OrgName, OrgCode string
	}
	type campusRow struct {
		ID                     int32
		CampusName, CampusCode string
	}
	items := make([]TargetOption, 0, limit)
	var orgs []orgRow
	orgLimit := limit
	if orgLimit > 10 {
		orgLimit = 10
	}
	if err := applyIDOrLike(db.Table("organization"), keyword, "org_name", "org_code").Select("id, org_name, org_code").Order("id DESC").Limit(orgLimit).Scan(&orgs).Error; err != nil {
		return nil, err
	}
	for _, row := range orgs {
		items = append(items, TargetOption{ID: row.ID, Module: "tenant", Name: row.OrgName, Label: fmt.Sprintf("%s（机构 #%d）", row.OrgName, row.ID), Extra: row.OrgCode})
	}
	remaining := limit - len(items)
	if remaining <= 0 {
		return items, nil
	}
	var campuses []campusRow
	if err := applyIDOrLike(db.Table("campus"), keyword, "campus_name", "campus_code").Select("id, campus_name, campus_code").Order("id DESC").Limit(remaining).Scan(&campuses).Error; err != nil {
		return nil, err
	}
	for _, row := range campuses {
		items = append(items, TargetOption{ID: row.ID, Module: "tenant", Name: row.CampusName, Label: fmt.Sprintf("%s（校区 #%d）", row.CampusName, row.ID), Extra: row.CampusCode})
	}
	return items, nil
}

func searchStudentTargets(db *gorm.DB, keyword string, limit int) ([]TargetOption, error) {
	type row struct {
		ID                             int32
		StudentName, Phone, ParentName string
	}
	var rows []row
	err := applyIDOrLike(db.Table("student"), keyword, "student_name", "phone", "parent_name", "parent_phone").Select("id, student_name, phone, parent_name").Order("id DESC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]TargetOption, 0, len(rows))
	for _, row := range rows {
		items = append(items, TargetOption{ID: row.ID, Module: "student", Name: row.StudentName, Label: fmt.Sprintf("%s（学员 #%d）", row.StudentName, row.ID), Extra: firstNonEmpty(row.Phone, row.ParentName)})
	}
	return items, nil
}

func searchCourseTargets(db *gorm.DB, keyword string, limit int) ([]TargetOption, error) {
	type row struct {
		ID                  int32
		CourseName, Subject string
	}
	var rows []row
	err := applyIDOrLike(db.Table("course"), keyword, "course_name", "subject", "course_type").Select("id, course_name, subject").Order("id DESC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]TargetOption, 0, len(rows))
	for _, row := range rows {
		items = append(items, TargetOption{ID: row.ID, Module: "course", Name: row.CourseName, Label: fmt.Sprintf("%s（课程 #%d）", row.CourseName, row.ID), Extra: row.Subject})
	}
	return items, nil
}

func searchLessonPackageTargets(db *gorm.DB, keyword string, limit int) ([]TargetOption, error) {
	type row struct {
		ID          int32
		PackageName string
		StudentID   int32
	}
	var rows []row
	err := applyIDOrLike(db.Table("lesson_package lp").Joins("LEFT JOIN student s ON s.id = lp.student_id").Joins("LEFT JOIN course c ON c.id = lp.course_id"), keyword, "s.student_name", "c.course_name", "lp.remark").Select("lp.id, CONCAT(s.student_name, ' / ', c.course_name) AS package_name, lp.student_id").Order("id DESC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]TargetOption, 0, len(rows))
	for _, row := range rows {
		items = append(items, TargetOption{ID: row.ID, Module: "lesson_package", Name: row.PackageName, Label: fmt.Sprintf("%s（课包 #%d）", row.PackageName, row.ID), Extra: fmt.Sprintf("学员 #%d", row.StudentID)})
	}
	return items, nil
}

func searchPaymentTargets(db *gorm.DB, keyword string, limit int) ([]TargetOption, error) {
	type row struct {
		ID        int32
		PaymentNo string
		StudentID int32
		Amount    float64
	}
	var rows []row
	err := applyIDOrLike(db.Table("payment_record pr").Joins("LEFT JOIN student s ON s.id = pr.student_id"), keyword, "s.student_name", "pr.remark", "pr.payment_method", "pr.payment_type").Select("pr.id, CONCAT(s.student_name, ' ' , pr.payment_type) AS payment_no, pr.student_id, CAST(pr.amount AS DECIMAL(10,2)) AS amount").Order("id DESC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]TargetOption, 0, len(rows))
	for _, row := range rows {
		name := firstNonEmpty(row.PaymentNo, fmt.Sprintf("缴费记录 #%d", row.ID))
		items = append(items, TargetOption{ID: row.ID, Module: "payment", Name: name, Label: fmt.Sprintf("%s（缴费 #%d）", name, row.ID), Extra: fmt.Sprintf("%.2f", row.Amount)})
	}
	return items, nil
}

func searchScheduleTargets(db *gorm.DB, keyword string, limit int) ([]TargetOption, error) {
	type row struct {
		ID                  int32
		StudentID, CourseID int32
		ScheduleDate        string
	}
	var rows []row
	err := applyIDOrLike(db.Table("schedule sc").Joins("LEFT JOIN student s ON s.id = sc.student_id").Joins("LEFT JOIN course c ON c.id = sc.course_id"), keyword, "s.student_name", "c.course_name", "sc.classroom", "sc.remark", "CAST(sc.class_date AS CHAR)").Select("sc.id, sc.student_id, sc.course_id, CAST(sc.class_date AS CHAR) AS schedule_date").Order("id DESC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]TargetOption, 0, len(rows))
	for _, row := range rows {
		name := fmt.Sprintf("排课 %s", row.ScheduleDate)
		items = append(items, TargetOption{ID: row.ID, Module: "schedule", Name: name, Label: fmt.Sprintf("%s（排课 #%d）", name, row.ID), Extra: fmt.Sprintf("学员 #%d / 课程 #%d", row.StudentID, row.CourseID)})
	}
	return items, nil
}

func searchRescheduleTargets(db *gorm.DB, keyword string, limit int) ([]TargetOption, error) {
	type row struct {
		ID         int32
		ScheduleID int32
		Reason     string
	}
	var rows []row
	err := applyIDOrLike(db.Table("reschedule_record"), keyword, "reason", "operation_type").Select("id, old_schedule_id AS schedule_id, reason").Order("id DESC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]TargetOption, 0, len(rows))
	for _, row := range rows {
		name := firstNonEmpty(row.Reason, fmt.Sprintf("调课记录 #%d", row.ID))
		items = append(items, TargetOption{ID: row.ID, Module: "reschedule", Name: name, Label: fmt.Sprintf("%s（调课 #%d）", name, row.ID), Extra: fmt.Sprintf("排课 #%d", row.ScheduleID)})
	}
	return items, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return "-"
}
