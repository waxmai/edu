export function cleanQueryParams<T extends Record<string, unknown>>(params: T): Partial<T> {
  const normalizedEntries = Object.entries(params).flatMap(([key, value]) => {
    if (value === '' || value === null || value === undefined) {
      return []
    }

    if (typeof value === 'string') {
      const trimmed = value.trim()
      if (!trimmed) {
        return []
      }
      if (/^\d+$/.test(trimmed)) {
        return [[key, Number(trimmed)]]
      }
      return [[key, trimmed]]
    }

    return [[key, value]]
  })

  return Object.fromEntries(normalizedEntries) as Partial<T>
}
