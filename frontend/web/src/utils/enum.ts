export function normalizeEnumValue<T extends string>(value: unknown, allowed: readonly T[], fallback = ''): T | '' {
  if (typeof value !== 'string') {
    return fallback as T | ''
  }
  const trimmed = value.trim() as T
  return allowed.includes(trimmed) ? trimmed : (fallback as T | '')
}
