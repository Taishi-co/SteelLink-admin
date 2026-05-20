import { describe, expect, it } from 'vitest'
import { canAudit, canOperate, isSuperAdmin } from './permissions'

describe('permission helpers', () => {
  it('maps admin roles to action permissions', () => {
    expect(isSuperAdmin(1)).toBe(true)
    expect(canOperate(1)).toBe(true)
    expect(canOperate(2)).toBe(true)
    expect(canOperate(3)).toBe(false)
    expect(canAudit(3)).toBe(true)
    expect(canAudit(4)).toBe(false)
  })
})
