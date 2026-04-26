import { get, post, put, del } from './request'

// Types
export interface LoginParams {
  username: string
  password: string
}

export interface LoginResult {
  token: string
  adminInfo: {
    id: number
    username: string
    realName: string
    avatar?: string
    roleId?: number
  }
}

export interface PageParams {
  page: number
  pageSize: number
  keyword?: string
  status?: number
}

export interface PageResult<T> {
  list: T[]
  total: number
}

export interface DashboardStats {
  totalUsers: number
  totalEnterprises: number
  totalProducts: number
  pendingAudits: number
}

// Auth API
export const authApi = {
  login: (params: LoginParams) => post<LoginResult>('/login', params),
  logout: () => post('/logout'),
  getInfo: () => get<LoginResult['adminInfo']>('/admin/info'),
}

export const dashboardApi = {
  stats: () => get<DashboardStats>('/dashboard/stats'),
}

// Admin API
export interface AdminItem {
  id: number
  username: string
  realName: string
  email?: string
  phone?: string
  avatar?: string
  status: number
  roleId?: number
  createdAt: string
}

export const adminApi = {
  list: (params: PageParams) => get<PageResult<AdminItem>>('/admins', { params }),
  get: (id: number) => get<AdminItem>(`/admins/${id}`),
  create: (data: Partial<AdminItem> & { password: string }) => post('/admins', data),
  update: (id: number, data: Partial<AdminItem>) => put(`/admins/${id}`, data),
  delete: (id: number) => del(`/admins/${id}`),
}

// Role API
export interface RoleItem {
  id: number
  name: string
  code: string
  description?: string
  status: number
  sort: number
  createdAt: string
}

export const roleApi = {
  list: (params: PageParams) => get<PageResult<RoleItem>>('/roles', { params }),
  get: (id: number) => get<RoleItem>(`/roles/${id}`),
  create: (data: Partial<RoleItem>) => post('/roles', data),
  update: (id: number, data: Partial<RoleItem>) => put(`/roles/${id}`, data),
  delete: (id: number) => del(`/roles/${id}`),
  getMenus: (id: number) => get<number[]>(`/roles/${id}/menus`),
  updateMenus: (id: number, menuIds: number[]) => put(`/roles/${id}/menus`, { menuIds }),
}

// Menu API
export interface MenuItem {
  id: number
  parentId: number
  name: string
  path: string
  component?: string
  icon?: string
  type: number
  permission?: string
  status: number
  sort: number
  children?: MenuItem[]
}

export const menuApi = {
  list: (params?: PageParams) => get<PageResult<MenuItem>>('/menus', { params }),
  get: (id: number) => get<MenuItem>(`/menus/${id}`),
  create: (data: Partial<MenuItem>) => post('/menus', data),
  update: (id: number, data: Partial<MenuItem>) => put(`/menus/${id}`, data),
  delete: (id: number) => del(`/menus/${id}`),
}

// User API
export interface UserItem {
  id: number
  username: string
  phone?: string
  email?: string
  realName?: string
  avatar?: string
  enterpriseId?: number
  status: number
  createdAt: string
}

export const userApi = {
  list: (params: PageParams) => get<PageResult<UserItem>>('/users', { params }),
  get: (id: number) => get<UserItem>(`/users/${id}`),
  updateStatus: (id: number, status: number) => put(`/users/${id}/status`, { status }),
}

// Enterprise API
export interface EnterpriseItem {
  id: number
  name: string
  shortName?: string
  creditCode?: string
  legalPerson?: string
  contactPerson?: string
  contactPhone?: string
  address?: string
  province?: string
  city?: string
  district?: string
  businessLicense?: string
  status: number
  auditRemark?: string
  auditorId?: number
  auditedAt?: string
  createdAt: string
  updatedAt?: string
}

export const enterpriseApi = {
  list: (params: PageParams) => get<PageResult<EnterpriseItem>>('/enterprises', { params }),
  get: (id: number) => get<EnterpriseItem>(`/enterprises/${id}`),
  audit: (id: number, data: { status: number; auditRemark?: string }) => put(`/enterprises/${id}/audit`, data),
}

// Product API
export interface ProductItem {
  id: number
  name: string
  categoryId?: number
  enterpriseId: number
  category?: string
  subCategory?: string
  steelMill?: string
  spec?: string
  material?: string
  warehouseName?: string
  warehouseCode?: string
  province?: string
  city?: string
  district?: string
  address?: string
  bundleNo?: string
  unit?: string
  price: number
  stock: number
  minOrderQty?: number
  deliveryDays?: number
  quoteValidity?: number
  description?: string
  images?: string[]
  status: number
  auditRemark?: string
  auditorId?: number
  auditedAt?: string
  createdAt: string
  updatedAt?: string
}

export const productApi = {
  list: (params: PageParams) => get<PageResult<ProductItem>>('/products', { params }),
  get: (id: number) => get<ProductItem>(`/products/${id}`),
  audit: (id: number, data: { status: number; auditRemark?: string }) => put(`/products/${id}/audit`, data),
}

// Order API
export interface OrderItem {
  id: number
  orderNo: string
  buyerId: number
  buyerName: string
  sellerId: number
  sellerEnterpriseId: number
  sellerName: string
  totalQuantity: number
  totalAmount: number
  taxTotalAmount: number
  currency: string
  status: number
  deliveryVoucher?: string
  receiptVoucher?: string
  receiptAt?: string
  expiresAt?: string
  createdAt: string
  timeline?: OrderTimelineItem[]
}

export interface OrderTimelineItem {
  key: string
  label: string
  status: string
  time?: string
  description?: string
}

export const orderApi = {
  list: (params: PageParams) => get<PageResult<OrderItem>>('/orders', { params }),
}

// Product Category API
export interface CategoryItem {
  id: number
  parentId: number
  name: string
  code?: string
  icon?: string
  sort: number
  status: number
  children?: CategoryItem[]
}

export const categoryApi = {
  list: (params?: PageParams) => get<PageResult<CategoryItem>>('/product-categories', { params }),
  create: (data: Partial<CategoryItem>) => post('/product-categories', data),
  update: (id: number, data: Partial<CategoryItem>) => put(`/product-categories/${id}`, data),
  delete: (id: number) => del(`/product-categories/${id}`),
}

export interface OperationLogItem {
  id: number
  adminId: number
  adminName: string
  module: string
  action: string
  method: string
  path: string
  ip: string
  requestBody?: string
  responseCode: number
  createdAt: string
}

export interface OperationLogParams extends PageParams {
  module?: string
  action?: string
  adminName?: string
  startDate?: string
  endDate?: string
}

export const operationLogApi = {
  list: (params: OperationLogParams) => get<PageResult<OperationLogItem>>('/operation-logs', { params }),
}
