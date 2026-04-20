import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AdminInfo {
  id: number
  username: string
  realName: string
  avatar?: string
  roleId?: number
}

interface AuthState {
  token: string | null
  adminInfo: AdminInfo | null
  setToken: (token: string) => void
  setAdminInfo: (info: AdminInfo) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      adminInfo: null,
      setToken: (token) => set({ token }),
      setAdminInfo: (adminInfo) => set({ adminInfo }),
      logout: () => set({ token: null, adminInfo: null }),
    }),
    {
      name: 'steellink-admin-auth',
    }
  )
)
