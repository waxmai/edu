export function normalizeListPayload<T>(payload: unknown): { list: T[]; total: number } {
  if (Array.isArray(payload)) {
    return {
      list: payload as T[],
      total: payload.length,
    }
  }

  if (payload && typeof payload === 'object') {
    const data = payload as Record<string, unknown>
    const list = Array.isArray(data.list)
      ? (data.list as T[])
      : Array.isArray(data.records)
        ? (data.records as T[])
        : []
    const total = typeof data.total === 'number' ? data.total : list.length
    return { list, total }
  }

  return {
    list: [],
    total: 0,
  }
}
