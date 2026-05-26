import { ref } from 'vue'
import { fetchCourseList, type CourseItem } from '@/api/course'
import { fetchLessonPackageList } from '@/api/lessonPackage'
import { fetchStudentList } from '@/api/student'
import { fetchSystemUsers, type SystemUserItem } from '@/api/system'
import { fetchScheduleList, type ScheduleListItem } from '@/api/schedule'
import { usePlatformViewScope } from '@/composables/usePlatformViewScope'
import type { LessonPackageItem } from '@/modules/lesson-package/types/lesson-package'
import type { StudentDetail } from '@/modules/student/types/student'
import { normalizeListPayload } from '@/utils/api'
import { cleanQueryParams } from '@/utils/query'

const OPTION_PAGE_SIZE = 100
const TEACHER_ROLE_CODE = 'teacher'

function lessonPackageOptionLabel(item: LessonPackageItem) {
  const student = item.studentName || (item.studentId ? `学员${item.studentId}` : '未关联学员')
  const course = item.courseName || item.subject || (item.courseId ? `课程${item.courseId}` : '未关联课程')
  const remain = typeof item.remainLessons === 'number' ? `剩余${item.remainLessons}课时` : ''
  return [student, course, `课时包#${item.id}`, remain].filter(Boolean).join(' · ')
}

export function useFormOptions() {
  const { withPlatformViewScope } = usePlatformViewScope()
  const loading = ref(false)
  const studentOptions = ref<StudentDetail[]>([])
  const courseOptions = ref<CourseItem[]>([])
  const lessonPackageOptions = ref<LessonPackageItem[]>([])
  const teacherOptions = ref<SystemUserItem[]>([])
  const scheduleOptions = ref<ScheduleListItem[]>([])

  async function loadStudentOptions() {
    loading.value = true
    try {
      const studentsResp = await fetchStudentList(withPlatformViewScope({ pageNum: 1, pageSize: OPTION_PAGE_SIZE }))
      studentOptions.value = normalizeListPayload<StudentDetail>(studentsResp.data?.data).list
    } finally {
      loading.value = false
    }
  }

  async function loadBaseOptions() {
    loading.value = true
    try {
      const [studentsResp, coursesResp, teachersResp] = await Promise.all([
        fetchStudentList(withPlatformViewScope({ pageNum: 1, pageSize: OPTION_PAGE_SIZE })),
        fetchCourseList(withPlatformViewScope({ pageNum: 1, pageSize: OPTION_PAGE_SIZE })),
        fetchSystemUsers(withPlatformViewScope({ pageNum: 1, pageSize: OPTION_PAGE_SIZE, roleCode: TEACHER_ROLE_CODE, status: 'enabled' })),
      ])
      studentOptions.value = normalizeListPayload<StudentDetail>(studentsResp.data?.data).list
      courseOptions.value = normalizeListPayload<CourseItem>(coursesResp.data?.data).list
      teacherOptions.value = teachersResp.data?.data?.list || []
    } finally {
      loading.value = false
    }
  }

  async function loadLessonPackageOptions(studentId?: number | null) {
    const response = await fetchLessonPackageList(cleanQueryParams(withPlatformViewScope({
      pageNum: 1,
      pageSize: OPTION_PAGE_SIZE,
      ...(typeof studentId === 'number' ? { studentId } : {}),
    })))
    lessonPackageOptions.value = normalizeListPayload<LessonPackageItem>(response.data?.data).list
  }

  async function loadScheduleOptions(filters?: { studentId?: number | string; teacherId?: number | string }) {
    const response = await fetchScheduleList(cleanQueryParams(withPlatformViewScope({
      pageNum: 1,
      pageSize: OPTION_PAGE_SIZE,
      studentId: filters?.studentId,
      teacherId: filters?.teacherId,
    })))
    scheduleOptions.value = normalizeListPayload<ScheduleListItem>(response.data?.data).list
  }

  return {
    loading,
    studentOptions,
    courseOptions,
    lessonPackageOptions,
    teacherOptions,
    scheduleOptions,
    loadStudentOptions,
    loadBaseOptions,
    loadLessonPackageOptions,
    loadScheduleOptions,
    lessonPackageOptionLabel,
  }
}
