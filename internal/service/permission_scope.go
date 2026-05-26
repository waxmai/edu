package service

import (
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/service/apperr"
)

func EnsureDataScopeStudent(actor proposal.SessionUserInfo, studentID int32) error {
	if actor.DataScope == proposal.DataScopeAll || actor.DataScope == proposal.DataScopeOrg || actor.DataScope == proposal.DataScopeCampus {
		return nil
	}
	if actor.DataScope == proposal.DataScopeSelf && studentID == actor.Id {
		return nil
	}
	return apperr.Forbidden("当前账号无权访问该学员数据")
}

func EnsureDataScopeTeacher(actor proposal.SessionUserInfo, teacherID int32) error {
	if actor.DataScope == proposal.DataScopeAll || actor.DataScope == proposal.DataScopeOrg || actor.DataScope == proposal.DataScopeCampus {
		return nil
	}
	if actor.DataScope == proposal.DataScopeSelf && teacherID == actor.Id {
		return nil
	}
	return apperr.Forbidden("当前账号无权访问该教师数据")
}

func EnsureDataScopeStudentOrTeacher(actor proposal.SessionUserInfo, studentID, teacherID int32) error {
	if actor.DataScope == proposal.DataScopeAll || actor.DataScope == proposal.DataScopeOrg || actor.DataScope == proposal.DataScopeCampus {
		return nil
	}
	if actor.DataScope == proposal.DataScopeSelf && (studentID == actor.Id || teacherID == actor.Id) {
		return nil
	}
	return apperr.Forbidden("当前账号无权访问该数据")
}
