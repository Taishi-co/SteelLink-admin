export function isSuperAdmin(roleId?: number | null) {
  return Number(roleId || 0) === 1
}

export function canAudit(roleId?: number | null) {
  return [1, 2, 3].includes(Number(roleId || 0))
}

export function canOperate(roleId?: number | null) {
  return [1, 2].includes(Number(roleId || 0))
}
