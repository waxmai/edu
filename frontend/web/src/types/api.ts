export interface ApiListResult<T> {
  list: T[]
  total: number
}

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}
