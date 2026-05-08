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
  getMyMenus: () => get<MenuItem[]>('/admin/menus'),
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
  identityStatus?: number
  idCard?: string
  idCardFrontUrl?: string
  idCardBackUrl?: string
  identityReject?: string
  status: number
  createdAt: string
}

export const userApi = {
  list: (params: PageParams) => get<PageResult<UserItem>>('/users', { params }),
  get: (id: number) => get<UserItem>(`/users/${id}`),
  updateStatus: (id: number, status: number) => put(`/users/${id}/status`, { status }),
  auditIdentity: (id: number, data: { status: number; reject?: string }) => put(`/users/${id}/identity`, data),
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
  services: (id: number) => get<ServicePermissionItem[]>(`/enterprises/${id}/services`),
  updateServiceStatus: (id: number, service: string, data: { status: number; reason?: string }) =>
    put(`/enterprises/${id}/services/${service}/status`, data),
}

export interface ServiceDocumentItem {
  id: number
  applicationId: number
  documentType: string
  fileUrl: string
  createdAt: string
}

export interface ServiceApplicationItem {
  id: number
  enterpriseId: number
  enterpriseCode: string
  enterpriseName: string
  service: string
  enterpriseRole: string
  status: number
  auditReason?: string
  submittedBy: number
  auditedBy?: number
  auditedAt?: string
  documents: ServiceDocumentItem[]
  createdAt: string
  updatedAt: string
}

export interface ServicePermissionItem {
  id: number
  enterpriseId: number
  enterpriseCode: string
  service: string
  capabilityKind?: 'trade' | 'provider'
  enterpriseRole: string
  status: number
  auditReason?: string
  approvedAt?: string
  updatedAt: string
}

export const enterpriseServiceApi = {
  listApplications: (params: PageParams & { service?: string }) =>
    get<PageResult<ServiceApplicationItem>>('/enterprise-service-applications', { params }),
  auditApplication: (id: number, data: { status: number; reason?: string }) =>
    put(`/enterprise-service-applications/${id}/audit`, data),
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
  goodsType?: 'spot' | 'futures'
  warehouseServiceId?: number
  warehouseEnterpriseId?: number
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

export interface ServiceOfferItem {
  id: number
  type: 'warehouse' | 'logistics' | 'finance'
  enterpriseId: number
  name: string
  province?: string
  city?: string
  address?: string
  capacity?: number
  availableCapacity?: number
  routeStart?: string
  routeEnd?: string
  serviceScope?: string
  vehicleType?: string
  loadCapacity?: number
  priceRule?: string
  loadingCapacity?: string
  businessHours?: string
  pledgeSupervision?: boolean
  outboundVoucherReq?: string
  quoteRules?: string
  fleetName?: string
  vehicleCerts?: string
  driverCerts?: string
  financeType?: string
  minAmount?: number
  maxAmount?: number
  termDays?: number
  annualRate?: number
  status: number
  auditRemark?: string
  createdAt: string
  updatedAt: string
}

export const serviceOfferApi = {
  listWarehouse: (params: PageParams) => get<PageResult<ServiceOfferItem>>('/warehouse-services', { params }),
  auditWarehouse: (id: number, data: { status: number; auditRemark?: string }) => put(`/warehouse-services/${id}/audit`, data),
  listLogistics: (params: PageParams) => get<PageResult<ServiceOfferItem>>('/logistics-services', { params }),
  auditLogistics: (id: number, data: { status: number; auditRemark?: string }) => put(`/logistics-services/${id}/audit`, data),
  listFinance: (params: PageParams) => get<PageResult<ServiceOfferItem>>('/finance-services', { params }),
  auditFinance: (id: number, data: { status: number; auditRemark?: string }) => put(`/finance-services/${id}/audit`, data),
}

// Order API
export interface OrderItem {
  [key: string]: unknown
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
  remark?: Record<string, unknown>
  financingInfo?: FinancingRequestInfo
  financing_info?: FinancingRequestInfo
  logisticsBidding?: LogisticsBiddingInfo
  logistics_bidding?: LogisticsBiddingInfo
  logisticsQuotes?: LogisticsQuoteInfo[]
  logistics_quotes?: LogisticsQuoteInfo[]
  shipmentTask?: ShipmentTaskInfo
  shipment_task?: ShipmentTaskInfo
  pledgeEvents?: PledgeEventInfo[]
  pledge_events?: PledgeEventInfo[]
  events?: OrderEventItem[]
}

export interface OrderTimelineItem {
  key: string
  label: string
  status: string
  time?: string
  description?: string
}

export interface FinancingRequestInfo {
  [key: string]: unknown
  id: number
  status: string
  risk_result?: string
  riskRemark?: string
  risk_remark?: string
  approved_amount?: number
  approvedAmount?: number
  requested_amount?: number
  requestedAmount?: number
  funding_enterprise_id?: number
  fundingEnterpriseId?: number
  disbursement_voucher?: string
  disbursementVoucher?: string
  pledge_voucher_url?: string
  pledgeVoucherUrl?: string
  commission_voucher?: string
  commissionVoucher?: string
  debt_contract_url?: string
  debtContractUrl?: string
}

export interface LogisticsBiddingInfo {
  [key: string]: unknown
  id: number
  status: string
  freight_payer?: string
  freightPayer?: string
  selected_quote_id?: number
  selectedQuoteId?: number
}

export interface LogisticsQuoteInfo {
  [key: string]: unknown
  id: number
  logistics_name?: string
  logisticsName?: string
  logistics_enterprise_id?: number
  logisticsEnterpriseId?: number
  amount: number
  vehicle_info?: string
  vehicleInfo?: string
  driver_info?: string
  driverInfo?: string
  estimated_delivery_at?: string
  estimatedDeliveryAt?: string
  remark?: string
}

export interface ShipmentTaskInfo {
  [key: string]: unknown
  id: number
  status: string
  logistics_enterprise_id?: number
  logisticsEnterpriseId?: number
  logistics_quote_id?: number
  logisticsQuoteId?: number
  pickup_voucher?: string
  pickupVoucher?: string
  delivery_voucher?: string
  deliveryVoucher?: string
  signature_voucher?: string
  signatureVoucher?: string
}

export interface PledgeEventInfo {
  [key: string]: unknown
  id: number
  event_type?: string
  eventType?: string
  voucher_url?: string
  voucherUrl?: string
  reason?: string
  condition?: string
  created_at?: string
  createdAt?: string
}

export interface OrderEventItem {
  [key: string]: unknown
  id: number
  event_type?: string
  eventType?: string
  actor_service?: string
  actorService?: string
  from_status?: number
  fromStatus?: number
  to_status?: number
  toStatus?: number
  payload?: Record<string, unknown>
  created_at?: string
  createdAt?: string
}

export const orderApi = {
  list: (params: PageParams) => get<PageResult<OrderItem>>('/orders', { params }),
  detail: (id: number) => get<OrderItem>(`/orders/${id}`),
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

export interface PrivateFileURLResult {
  url: string
  expiresIn: number
}

export const fileApi = {
  getPrivateUrl: (url: string) => get<PrivateFileURLResult>('/files/private-url', { params: { url } }),
}

export interface AIProcurementStageConfig {
  code: string
  name: string
  enabled: boolean
  required: boolean
  sort: number
  timeoutSec: number
}

export interface AIProcurementConfig {
  enabled: boolean
  stages: AIProcurementStageConfig[]
  model: {
    parsingEnabled: boolean
    provider: string
    baseUrl: string
    apiKeyEnv: string
    apiKeyConfigured: boolean
    model: string
    temperature: number
    topP: number
    maxTokens: number
    timeoutSec: number
    retryCount: number
    confidenceThreshold: number
  }
  matching: {
    mode: string
    vectorEnabled: boolean
    vectorProvider: string
    vectorEndpoint: string
    vectorDatabase: string
    vectorUsername: string
    vectorPasswordEnv: string
    vectorConnectionEnv: string
    productCollection: string
    supplierCollection: string
    financeProductCollection: string
    embeddingProvider: string
    embeddingBaseUrl: string
    embeddingApiKeyEnv: string
    embeddingModel: string
    embeddingDimension: number
    vectorTopK: number
    vectorMinScore: number
    hybridKeywordWeight: number
    hybridVectorWeight: number
    categoryBoost: number
    supplierPriority: string
    priceTolerancePct: number
    maxRecommendations: number
  }
  finance: {
    enabled: boolean
    minAmount: number
    maxAmount: number
    requiredDocs: string[]
    autoSubmitReview: boolean
  }
  review: {
    reviewerRole: string
    manualReviewThreshold: number
    riskKeywords: string[]
  }
  updatedAt?: string
}

export const aiProcurementApi = {
  getConfig: () => get<AIProcurementConfig>('/ai-procurement/config'),
  updateConfig: (data: AIProcurementConfig) => put<AIProcurementConfig>('/ai-procurement/config', data),
}
