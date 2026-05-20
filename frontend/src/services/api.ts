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
  enterpriseId?: number
  category?: string
  subCategory?: string
  goodsType?: string
  warehouseServiceId?: number
  intervention?: number
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

export interface SendNotificationPayload {
  recipient_user_ids?: number[]
  recipient_enterprise_ids?: number[]
  recipient_roles?: string[]
  send_to_all?: boolean
  type?: 'admin' | 'order' | 'audit' | 'system'
  title: string
  content: string
  link_url?: string
  metadata?: Record<string, unknown>
}

export const notificationApi = {
  send: (data: SendNotificationPayload) => post<{ created: number }>('/admin/notifications', data),
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

type BackendProductItem = Omit<ProductItem, 'spec'> & {
  spec?: string | Record<string, unknown>
  enterprise_id?: number
  sub_category?: string
  steel_mill?: string
  goods_type?: 'spot' | 'futures'
  warehouse_service_id?: number
  warehouse_enterprise_id?: number
  warehouse_info?: {
    warehouse_name?: string
    warehouse_code?: string
    province?: string
    city?: string
    district?: string
    address?: string
  }
  bundle_no?: string
  unit_price?: number
  available_qty?: number
  quantity?: number
  min_order_qty?: number
  delivery_days?: number
  quote_validity?: number
  audit_remark?: string
  approved_at?: string
  created_at?: string
  updated_at?: string
}

const productParams = (params: PageParams) => ({
  ...params,
  page_size: params.pageSize,
})

const normalizeProduct = (item: BackendProductItem): ProductItem => {
  const warehouse = item.warehouse_info || {}
  return {
    ...item,
    enterpriseId: item.enterpriseId ?? item.enterprise_id ?? 0,
    subCategory: item.subCategory ?? item.sub_category,
    steelMill: item.steelMill ?? item.steel_mill,
    goodsType: item.goodsType ?? item.goods_type,
    warehouseServiceId: item.warehouseServiceId ?? item.warehouse_service_id,
    warehouseEnterpriseId: item.warehouseEnterpriseId ?? item.warehouse_enterprise_id,
    warehouseName: item.warehouseName ?? warehouse.warehouse_name,
    warehouseCode: item.warehouseCode ?? warehouse.warehouse_code,
    province: item.province ?? warehouse.province,
    city: item.city ?? warehouse.city,
    district: item.district ?? warehouse.district,
    address: item.address ?? warehouse.address,
    bundleNo: item.bundleNo ?? item.bundle_no,
    spec: typeof item.spec === 'string' ? item.spec : item.spec ? JSON.stringify(item.spec) : undefined,
    price: item.price ?? item.unit_price ?? 0,
    stock: item.stock ?? item.available_qty ?? item.quantity ?? 0,
    minOrderQty: item.minOrderQty ?? item.min_order_qty,
    deliveryDays: item.deliveryDays ?? item.delivery_days,
    quoteValidity: item.quoteValidity ?? item.quote_validity,
    auditRemark: item.auditRemark ?? item.audit_remark,
    auditedAt: item.auditedAt ?? item.approved_at,
    createdAt: item.createdAt ?? item.created_at ?? '',
    updatedAt: item.updatedAt ?? item.updated_at,
  }
}

export const productApi = {
  list: async (params: PageParams) => {
    const response = await get<PageResult<BackendProductItem>>('/products', { params: productParams(params) })
    response.data.data.list = response.data.data.list.map(normalizeProduct)
    return response as unknown as Awaited<ReturnType<typeof get<PageResult<ProductItem>>>>
  },
  get: async (id: number) => {
    const response = await get<BackendProductItem>(`/products/${id}`)
    response.data.data = normalizeProduct(response.data.data)
    return response as unknown as Awaited<ReturnType<typeof get<ProductItem>>>
  },
  audit: (id: number, data: { status: number; auditRemark?: string }) =>
    put(`/products/${id}/audit`, { status: data.status, auditRemark: data.auditRemark, audit_remark: data.auditRemark }),
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

type ServiceOfferAuditPayload = { status: number; auditAction?: 'approve' | 'reject'; auditRemark?: string }

const serviceOfferAuditBody = (data: ServiceOfferAuditPayload) => ({
  status: data.status,
  audit_action: data.auditAction,
  auditAction: data.auditAction,
  audit_remark: data.auditRemark,
  auditRemark: data.auditRemark,
})

export const serviceOfferApi = {
  listWarehouse: (params: PageParams) => get<PageResult<ServiceOfferItem>>('/warehouse-services', { params: { ...params, page_size: params.pageSize } }),
  auditWarehouse: (id: number, data: ServiceOfferAuditPayload) => put(`/warehouse-services/${id}/audit`, serviceOfferAuditBody(data)),
  listLogistics: (params: PageParams) => get<PageResult<ServiceOfferItem>>('/logistics-services', { params: { ...params, page_size: params.pageSize } }),
  auditLogistics: (id: number, data: ServiceOfferAuditPayload) => put(`/logistics-services/${id}/audit`, serviceOfferAuditBody(data)),
  listFinance: (params: PageParams) => get<PageResult<ServiceOfferItem>>('/finance-services', { params: { ...params, page_size: params.pageSize } }),
  auditFinance: (id: number, data: ServiceOfferAuditPayload) => put(`/finance-services/${id}/audit`, serviceOfferAuditBody(data)),
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
  terminationReason?: OrderTerminationReason
  termination_reason?: OrderTerminationReason
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
  breachRecords?: BreachRecordInfo[]
  breach_records?: BreachRecordInfo[]
  litigationRecords?: LitigationRecordInfo[]
  litigation_records?: LitigationRecordInfo[]
}

export interface OrderTerminationReason {
  code: string
  label: string
  party?: string
  partyLabel?: string
  reason?: string
  source?: string
  eventType?: string
  needsIntervention: boolean
  createdAt?: string
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

export interface BreachRecordInfo {
  id: number
  orderId: number
  orderNo?: string
  reporterEnterpriseId: number
  breachParty: string
  breachEnterpriseId: number
  breachType: string
  description: string
  evidenceUrls?: string[]
  status: string
  mediationResult?: string
  mediatedBy?: number
  mediatedAt?: string
  resolutionResult?: string
  resolvedBy?: number
  resolvedAt?: string
  createdAt: string
  updatedAt: string
}

export interface LitigationRecordInfo {
  id: number
  breachRecordId: number
  orderId: number
  orderNo?: string
  plaintiffEnterpriseId: number
  defendantEnterpriseId: number
  caseNo?: string
  court?: string
  claimAmount: number
  description: string
  evidenceUrls?: string[]
  status: string
  judgmentResult?: string
  judgmentUrl?: string
  judgedAt?: string
  closedAt?: string
  createdAt: string
  updatedAt: string
}

export const orderApi = {
  list: (params: PageParams) => get<PageResult<OrderItem>>('/orders', { params }),
  detail: (id: number) => get<OrderItem>(`/orders/${id}`),
  freeze: (id: number, reason: string) => post(`/orders/${id}/freeze`, { reason }),
  unfreeze: (id: number) => post(`/orders/${id}/unfreeze`),
  mediateBreach: (id: number, result: string) => post(`/orders/breach-records/${id}/mediate`, { result }),
  resolveBreach: (id: number, result: string) => post(`/orders/breach-records/${id}/resolve`, { result }),
  updateLitigation: (id: number, data: { status: string; judgmentResult?: string; judgmentUrl?: string; claimAmount?: number }) => put(`/orders/litigation-records/${id}`, data),
}

// Product Category API
export interface CategoryItem {
  id: number
  parentId: number
  name: string
  displayName?: string
  code?: string
  icon?: string
  imageUrl?: string
  sort: number
  status: number
  productCount?: number
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

export interface UploadFileResult {
  url: string
  key: string
}

export const fileApi = {
  getPrivateUrl: (url: string) => get<PrivateFileURLResult>('/files/private-url', { params: { url } }),
  upload: (file: File, purpose: string) => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('purpose', purpose)
    return post<UploadFileResult>('/files/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
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
    visionProvider: string
    visionBaseUrl: string
    visionApiKeyEnv: string
    visionApiKeyConfigured: boolean
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
