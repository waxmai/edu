export interface ScheduleItem {
  id: number
  studentId?: number
  courseId?: number
  teacherId?: number
  lessonPackageId?: number
  studentName?: string
  courseName?: string
  teacherName?: string
  lessonPackageName?: string
  classDate?: string
  startTime: string
  endTime: string
  classroom?: string
  status: string
  scheduleStatus?: string
  remark?: string
}
