import { useState } from 'react'
import { Alert, Card, Table, Tag, Input, Select, Space, Button, Descriptions, Drawer, Typography, Timeline, Divider, Form, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { BreachRecordInfo, LitigationRecordInfo, LogisticsQuoteInfo, orderApi, OrderEventItem, OrderItem, OrderTimelineItem, PageParams } from '../../services/api'
import { formatDateTime } from '../../utils/format'
import { EyeOutlined, SearchOutlined } from '@ant-design/icons'

const { Text } = Typography

const orderStatusMeta: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '草稿' },
  1: { color: 'orange', text: '待确认' },
  2: { color: 'gold', text: '待签合同' },
  3: { color: 'blue', text: '待支付' },
  4: { color: 'cyan', text: '已支付' },
  5: { color: 'purple', text: '待提货' },
  6: { color: 'processing', text: '配送中' },
  7: { color: 'geekblue', text: '已送达' },
  8: { color: 'green', text: '已完成' },
  9: { color: 'red', text: '已取消' },
  10: { color: 'volcano', text: '已冻结' },
  11: { color: 'magenta', text: '异常' },
  12: { color: 'purple', text: '待选择目的地仓储' },
  13: { color: 'purple', text: '待仓库确认' },
  14: { color: 'orange', text: '待卖方确认' },
  15: { color: 'blue', text: '待买方确认' },
  16: { color: 'cyan', text: '物流竞价中' },
  17: { color: 'gold', text: '待确认物流' },
  18: { color: 'processing', text: '运输中' },
  19: { color: 'green', text: '已收货' },
  20: { color: 'success', text: '已对账' },
  21: { color: 'default', text: '已归档' },
  22: { color: 'purple', text: '融资审批中' },
  23: { color: 'gold', text: '待签署融资合同' },
  24: { color: 'red', text: '待重新选仓' },
  25: { color: 'red', text: '待重新发起物流' },
  26: { color: 'purple', text: '待选择仓库' },
  27: { color: 'gold', text: '待物流合同签署' },
}

function timelineColor(status: string) {
  if (status === 'completed') return 'green'
  if (status === 'current') return 'blue'
  if (status === 'skipped') return 'gray'
  return 'gray'
}

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待选择目的地仓储', value: '12' },
  { label: '待仓库确认', value: '13' },
  { label: '待卖方确认', value: '14' },
  { label: '待买方确认', value: '15' },
  { label: '待签合同', value: '2' },
  { label: '待支付', value: '3' },
  { label: '融资审批中', value: '22' },
  { label: '待签署融资合同', value: '23' },
  { label: '待重新选仓', value: '24' },
  { label: '已支付', value: '4' },
  { label: '物流竞价中', value: '16' },
  { label: '待确认物流', value: '17' },
  { label: '待重新发起物流', value: '25' },
  { label: '待物流合同签署', value: '27' },
  { label: '运输中', value: '18' },
  { label: '已送达', value: '7' },
  { label: '已收货', value: '19' },
  { label: '已对账', value: '20' },
  { label: '已归档', value: '21' },
  { label: '已完成', value: '8' },
  { label: '已冻结', value: '10' },
  { label: '已取消', value: '9' },
  { label: '异常', value: '11' },
]

const breachStatusMeta: Record<string, { color: string; text: string }> = {
  submitted: { color: 'orange', text: '已提交' },
  under_review: { color: 'gold', text: '审核中' },
  mediating: { color: 'blue', text: '协调中' },
  litigating: { color: 'purple', text: '诉讼中' },
  resolved: { color: 'green', text: '已处理' },
  dismissed: { color: 'default', text: '已驳回' },
}

const litigationStatusOptions = [
  { label: '已立案', value: 'filed' },
  { label: '已受理', value: 'accepted' },
  { label: '审理中', value: 'in_progress' },
  { label: '已判决', value: 'judged' },
  { label: '已结案', value: 'closed' },
]

const litigationStatusMeta: Record<string, { color: string; text: string }> = {
  filed: { color: 'orange', text: '已立案' },
  accepted: { color: 'blue', text: '已受理' },
  in_progress: { color: 'processing', text: '审理中' },
  judged: { color: 'purple', text: '已判决' },
  closed: { color: 'green', text: '已结案' },
}

const valueOf = <T,>(record: Record<string, unknown> | null | undefined, camel: string, snake: string, fallback?: T): T | undefined => {
  if (!record) return fallback
  return (record[camel] ?? record[snake] ?? fallback) as T | undefined
}

const money = (value: unknown) => Number(value || 0).toLocaleString()

const eventLabel: Record<string, string> = {
  warehouse_rejected: '仓储拒绝',
  warehouse_reselected: '重新选仓',
  logistics_no_capacity: '物流无运力',
  direct_title_control_recorded: '直采货权监管',
  financing_requested: '申请托盘垫资',
  financing_risk_approved: '风控通过',
  financing_risk_rejected: '风控拒绝',
  financing_disbursed: '资金方放款',
  ownership_pledge_recorded: '货权质押记录',
  ownership_pledge_released: '货权释放',
  logistics_bidding_created: '物流竞价发起',
  logistics_quote_submitted: '物流报价',
  logistics_quote_confirmed: '物流报价确认',
}

export default function OrderList() {
  const queryClient = useQueryClient()
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [interventionOnly, setInterventionOnly] = useState(false)
  const [selectedOrder, setSelectedOrder] = useState<OrderItem | null>(null)
  const [freezeReason, setFreezeReason] = useState('')
  const [breachDrafts, setBreachDrafts] = useState<Record<number, string>>({})
  const [litigationDrafts, setLitigationDrafts] = useState<Record<number, { status?: string; judgmentResult?: string; judgmentUrl?: string }>>({})

  const queryParams = { ...params, keyword: keyword || undefined, status: statusFilter ? statusFilter as unknown as number : undefined, intervention: interventionOnly ? 1 : undefined }

  const { data, isLoading } = useQuery({
    queryKey: ['orders', queryParams],
    queryFn: () => orderApi.list(queryParams),
  })

  const selectedOrderId = selectedOrder?.id
  const { data: detailData, isLoading: detailLoading } = useQuery({
    queryKey: ['order-detail', selectedOrderId],
    queryFn: () => orderApi.detail(selectedOrderId!),
    enabled: !!selectedOrderId,
  })
  const orderDetail = (detailData?.data.data || selectedOrder) as OrderItem | null
  const remark = (orderDetail?.remark || {}) as Record<string, unknown>
  const financingInfo = (valueOf(orderDetail, 'financingInfo', 'financing_info') || null) as Record<string, unknown> | null
  const logisticsBidding = (valueOf(orderDetail, 'logisticsBidding', 'logistics_bidding') || null) as Record<string, unknown> | null
  const logisticsQuotes = (valueOf<LogisticsQuoteInfo[]>(orderDetail, 'logisticsQuotes', 'logistics_quotes', []) || []) as LogisticsQuoteInfo[]
  const shipmentTask = (valueOf(orderDetail, 'shipmentTask', 'shipment_task') || null) as Record<string, unknown> | null
  const pledgeEvents = (valueOf<Record<string, unknown>[]>(orderDetail, 'pledgeEvents', 'pledge_events', []) || []) as Record<string, unknown>[]
  const orderEvents = ((orderDetail?.events || []) as OrderEventItem[])
  const breachRecords = (valueOf<BreachRecordInfo[]>(orderDetail, 'breachRecords', 'breach_records', []) || []) as BreachRecordInfo[]
  const litigationRecords = (valueOf<LitigationRecordInfo[]>(orderDetail, 'litigationRecords', 'litigation_records', []) || []) as LitigationRecordInfo[]

  const refreshOrderDetail = () => {
    queryClient.invalidateQueries({ queryKey: ['orders'] })
    if (selectedOrderId) queryClient.invalidateQueries({ queryKey: ['order-detail', selectedOrderId] })
  }

  const freezeMutation = useMutation({
    mutationFn: () => orderApi.freeze(selectedOrderId!, freezeReason.trim()),
    onSuccess: () => { message.success('订单已冻结'); setFreezeReason(''); refreshOrderDetail() },
  })
  const unfreezeMutation = useMutation({
    mutationFn: () => orderApi.unfreeze(selectedOrderId!),
    onSuccess: () => { message.success('订单已解冻'); refreshOrderDetail() },
  })
  const mediateMutation = useMutation({
    mutationFn: ({ id, result }: { id: number; result: string }) => orderApi.mediateBreach(id, result),
    onSuccess: () => { message.success('协调结果已记录'); refreshOrderDetail() },
  })
  const resolveMutation = useMutation({
    mutationFn: ({ id, result }: { id: number; result: string }) => orderApi.resolveBreach(id, result),
    onSuccess: () => { message.success('违约处理已完成'); refreshOrderDetail() },
  })
  const litigationMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: { status: string; judgmentResult?: string; judgmentUrl?: string } }) => orderApi.updateLitigation(id, data),
    onSuccess: () => { message.success('诉讼状态已更新'); refreshOrderDetail() },
  })

  const columns: ColumnsType<OrderItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: '订单号', key: 'orderNo', width: 200, render: (_, r) => valueOf(r, 'orderNo', 'order_no', '-') },
    {
      title: '买方', key: 'buyer', width: 160,
      render: (_, r) => (
        <div>
          <div>{valueOf(r, 'buyerName', 'buyer_name', '-')}</div>
          <Text type="secondary" style={{ fontSize: 12 }}>ID: {valueOf(r, 'buyerId', 'buyer_id', '-')}</Text>
        </div>
      ),
    },
    {
      title: '卖方', key: 'seller', width: 160,
      render: (_, r) => (
        <div>
          <div>{valueOf(r, 'sellerName', 'seller_name', '-')}</div>
          <Text type="secondary" style={{ fontSize: 12 }}>ID: {valueOf(r, 'sellerId', 'seller_id', '-')}</Text>
        </div>
      ),
    },
    {
      title: '数量', key: 'totalQuantity', width: 100,
      render: (_, r) => `${valueOf(r, 'totalQuantity', 'total_quantity', 0)} 吨`,
    },
    {
      title: '金额', key: 'totalAmount', width: 130,
      render: (_, r) => `${r.currency || 'CNY'} ${money(valueOf(r, 'totalAmount', 'total_amount', 0))}`,
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 110,
      render: (s: number) => {
        const meta = orderStatusMeta[s] || { color: 'default', text: `状态${s}` }
        return <Tag color={meta.color}>{meta.text}</Tag>
      },
    },
    {
      title: '创建时间', key: 'createdAt', width: 170,
      render: (_, r) => formatDateTime(valueOf<string>(r, 'createdAt', 'created_at')),
    },
    {
      title: '操作', key: 'actions', width: 80,
      render: (_, record) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setSelectedOrder(record)}>详情</Button>
      ),
    },
  ]

  return (
    <>
      <Card title="订单管理">
        <Space className="mb-4" wrap>
          <Input
            placeholder="搜索订单号 / 买方 / 卖方"
            prefix={<SearchOutlined />}
            allowClear
            style={{ width: 280 }}
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onPressEnter={() => setParams({ ...params, page: 1 })}
          />
          <Select
            value={statusFilter}
            onChange={(v) => { setStatusFilter(v); setParams({ ...params, page: 1 }) }}
            style={{ width: 160 }}
            options={statusOptions}
          />
          <Select
            value={interventionOnly ? '1' : ''}
            onChange={(v) => { setInterventionOnly(v === '1'); setParams({ ...params, page: 1 }) }}
            style={{ width: 160 }}
            options={[
              { label: '全部订单', value: '' },
              { label: '待运营介入', value: '1' },
            ]}
          />
          <Button type="primary" onClick={() => setParams({ ...params, page: 1 })}>搜索</Button>
        </Space>

        <Table
          columns={columns}
          dataSource={data?.data.data.list || []}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1200 }}
          expandable={{
            expandedRowRender: (record) => (
              <Timeline
                items={(record.timeline || []).map((item: OrderTimelineItem) => ({
                  color: timelineColor(item.status),
                  children: (
                    <div>
                      <div style={{ fontWeight: 500 }}>{item.label}</div>
                      {item.time ? <Text type="secondary">{new Date(item.time).toLocaleString()}</Text> : null}
                      {item.description ? <div><Text type="secondary">{item.description}</Text></div> : null}
                    </div>
                  ),
                }))}
              />
            ),
          }}
          pagination={{
            current: params.page,
            pageSize: params.pageSize,
            total: data?.data.data.total || 0,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
          }}
        />
      </Card>

      <Drawer
        title={`订单详情 · ${valueOf(selectedOrder, 'orderNo', 'order_no', '') || ''}`}
        width={900}
        open={selectedOrder !== null}
        onClose={() => setSelectedOrder(null)}
        destroyOnClose
      >
        {orderDetail && (
          <>
            {detailLoading && <Alert type="info" showIcon message="正在加载完整履约详情" style={{ marginBottom: 16 }} />}
            {(remark.admin_attention || orderDetail.status === 24 || orderDetail.status === 25) && (
              <Alert
                type="warning"
                showIcon
                message="平台需处理事项"
                description={
                  orderDetail.status === 24
                    ? `仓储已拒绝，需协助买方重新选择仓储。原因：${remark.warehouse_reject_note || remark.condition_note || '-'}`
                    : orderDetail.status === 25
                      ? `物流反馈无运力，需协助卖方重新发起物流。原因：${remark.logistics_no_capacity_reason || '-'}`
                      : '该订单标记为需要平台关注，请查看下方事件记录。'
                }
                style={{ marginBottom: 16 }}
              />
            )}

            <Card size="small" title="平台处置" style={{ marginBottom: 20 }}>
              <Space direction="vertical" style={{ width: '100%' }} size={12}>
                <Space wrap>
                  {orderDetail.status === 10 ? (
                    <Button loading={unfreezeMutation.isPending} onClick={() => unfreezeMutation.mutate()}>解冻订单</Button>
                  ) : (
                    <Button danger loading={freezeMutation.isPending} disabled={!freezeReason.trim()} onClick={() => freezeMutation.mutate()}>冻结订单</Button>
                  )}
                  <Text type="secondary">冻结会暂停订单推进，并写入订单事件留痕。</Text>
                </Space>
                {orderDetail.status !== 10 && (
                  <Input.TextArea
                    rows={2}
                    placeholder="填写冻结原因"
                    value={freezeReason}
                    onChange={(e) => setFreezeReason(e.target.value)}
                    maxLength={500}
                    showCount
                  />
                )}
              </Space>
            </Card>
            <Descriptions bordered column={2} size="small">
              <Descriptions.Item label="订单号" span={2}>{valueOf(orderDetail, 'orderNo', 'order_no')}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={orderStatusMeta[orderDetail.status]?.color}>{orderStatusMeta[orderDetail.status]?.text}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="币种">{orderDetail.currency || 'CNY'}</Descriptions.Item>
            </Descriptions>

            <Descriptions bordered column={2} size="small" title="买方信息" style={{ marginTop: 20 }}>
              <Descriptions.Item label="企业名称">{valueOf(orderDetail, 'buyerName', 'buyer_name', '-')}</Descriptions.Item>
              <Descriptions.Item label="企业 ID">{valueOf(orderDetail, 'buyerId', 'buyer_id', '-')}</Descriptions.Item>
            </Descriptions>

            <Descriptions bordered column={2} size="small" title="卖方信息" style={{ marginTop: 20 }}>
              <Descriptions.Item label="企业名称">{valueOf(orderDetail, 'sellerName', 'seller_name', '-')}</Descriptions.Item>
              <Descriptions.Item label="企业 ID">{valueOf(orderDetail, 'sellerId', 'seller_id', '-')}</Descriptions.Item>
              <Descriptions.Item label="卖家企业 ID">{valueOf(orderDetail, 'sellerEnterpriseId', 'seller_enterprise_id', '-')}</Descriptions.Item>
            </Descriptions>

            <Descriptions bordered column={2} size="small" title="交易信息" style={{ marginTop: 20 }}>
              <Descriptions.Item label="总数量">{valueOf(orderDetail, 'totalQuantity', 'total_quantity', 0)} 吨</Descriptions.Item>
              <Descriptions.Item label="总金额">{money(valueOf(orderDetail, 'totalAmount', 'total_amount', 0))} 元</Descriptions.Item>
              <Descriptions.Item label="含税金额">{money(valueOf(orderDetail, 'taxTotalAmount', 'tax_total_amount', 0))} 元</Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatDateTime(valueOf<string>(orderDetail, 'createdAt', 'created_at'))}</Descriptions.Item>
              <Descriptions.Item label="过期时间">{formatDateTime(valueOf<string>(orderDetail, 'expiresAt', 'expires_at'))}</Descriptions.Item>
            </Descriptions>

            <Descriptions bordered column={2} size="small" title="履约信息" style={{ marginTop: 20 }}>
              <Descriptions.Item label="出库凭证">{valueOf(orderDetail, 'deliveryVoucher', 'delivery_voucher') ? <Tag color="blue">已上传</Tag> : <Tag>未上传</Tag>}</Descriptions.Item>
              <Descriptions.Item label="签收凭证">{valueOf(orderDetail, 'receiptVoucher', 'receipt_voucher') ? <Tag color="green">已上传</Tag> : <Tag>未上传</Tag>}</Descriptions.Item>
              <Descriptions.Item label="签收时间">{formatDateTime(valueOf<string>(orderDetail, 'receiptAt', 'receipt_at'))}</Descriptions.Item>
              <Descriptions.Item label="直采货权/仓单监管" span={2}>
                {remark.direct_title_voucher_url ? <Tag color="green">已记录</Tag> : <Tag>未记录</Tag>}
                {remark.direct_title_remark ? <Text type="secondary"> {String(remark.direct_title_remark)}</Text> : null}
              </Descriptions.Item>
            </Descriptions>

            {financingInfo && (
              <Descriptions bordered column={2} size="small" title="资金与质押" style={{ marginTop: 20 }}>
                <Descriptions.Item label="融资状态"><Tag color="purple">{String(financingInfo.status || '-')}</Tag></Descriptions.Item>
                <Descriptions.Item label="资金方企业 ID">{valueOf(financingInfo, 'fundingEnterpriseId', 'funding_enterprise_id', '-')}</Descriptions.Item>
                <Descriptions.Item label="申请金额">{money(valueOf(financingInfo, 'requestedAmount', 'requested_amount', 0))} 元</Descriptions.Item>
                <Descriptions.Item label="授信金额">{money(valueOf(financingInfo, 'approvedAmount', 'approved_amount', 0))} 元</Descriptions.Item>
                <Descriptions.Item label="风控结果">{String(financingInfo.risk_result || '-')}</Descriptions.Item>
                <Descriptions.Item label="风控备注">{String(valueOf(financingInfo, 'riskRemark', 'risk_remark', '-') || '-')}</Descriptions.Item>
                <Descriptions.Item label="债务合同">{valueOf(financingInfo, 'debtContractUrl', 'debt_contract_url') ? <Tag color="blue">已生成</Tag> : <Tag>未生成</Tag>}</Descriptions.Item>
                <Descriptions.Item label="分佣凭证">{valueOf(financingInfo, 'commissionVoucher', 'commission_voucher') ? <Tag color="green">已上传</Tag> : <Tag>未上传</Tag>}</Descriptions.Item>
                <Descriptions.Item label="放款凭证">{valueOf(financingInfo, 'disbursementVoucher', 'disbursement_voucher') ? <Tag color="green">已上传</Tag> : <Tag>未上传</Tag>}</Descriptions.Item>
                <Descriptions.Item label="货权质押">{valueOf(financingInfo, 'pledgeVoucherUrl', 'pledge_voucher_url') ? <Tag color="green">已记录</Tag> : <Tag color="red">未记录</Tag>}</Descriptions.Item>
              </Descriptions>
            )}

            {(logisticsBidding || logisticsQuotes.length > 0 || shipmentTask) && (
              <div style={{ marginTop: 20 }}>
                <Typography.Text strong style={{ fontSize: 14 }}>物流履约</Typography.Text>
                {logisticsBidding && (
                  <Descriptions bordered column={3} size="small" style={{ marginTop: 12 }}>
                    <Descriptions.Item label="竞价状态"><Tag color="cyan">{String(logisticsBidding.status || '-')}</Tag></Descriptions.Item>
                    <Descriptions.Item label="运费承担方">{valueOf(logisticsBidding, 'freightPayer', 'freight_payer') === 'seller' ? '卖方' : '买方'}</Descriptions.Item>
                    <Descriptions.Item label="选中报价 ID">{valueOf(logisticsBidding, 'selectedQuoteId', 'selected_quote_id', '-')}</Descriptions.Item>
                  </Descriptions>
                )}
                {logisticsQuotes.length > 0 && (
                  <Table
                    size="small"
                    rowKey="id"
                    style={{ marginTop: 12 }}
                    pagination={false}
                    dataSource={logisticsQuotes}
                    columns={[
                      { title: '物流方', render: (_, r) => valueOf(r, 'logisticsName', 'logistics_name', '-') },
                      { title: '报价', dataIndex: 'amount', render: (v) => `${money(v)} 元` },
                      { title: '车辆', render: (_, r) => valueOf(r, 'vehicleInfo', 'vehicle_info', '-') },
                      { title: '司机', render: (_, r) => valueOf(r, 'driverInfo', 'driver_info', '-') },
                      { title: '预计送达', render: (_, r) => valueOf(r, 'estimatedDeliveryAt', 'estimated_delivery_at', '-') },
                    ]}
                  />
                )}
                {shipmentTask && (
                  <Descriptions bordered column={3} size="small" style={{ marginTop: 12 }}>
                    <Descriptions.Item label="运输状态"><Tag color="processing">{String(shipmentTask.status || '-')}</Tag></Descriptions.Item>
                    <Descriptions.Item label="物流企业 ID">{valueOf(shipmentTask, 'logisticsEnterpriseId', 'logistics_enterprise_id', '-')}</Descriptions.Item>
                    <Descriptions.Item label="报价 ID">{valueOf(shipmentTask, 'logisticsQuoteId', 'logistics_quote_id', '-')}</Descriptions.Item>
                    <Descriptions.Item label="提货凭证">{valueOf(shipmentTask, 'pickupVoucher', 'pickup_voucher') ? <Tag color="green">已上传</Tag> : <Tag>未上传</Tag>}</Descriptions.Item>
                    <Descriptions.Item label="送达凭证">{valueOf(shipmentTask, 'deliveryVoucher', 'delivery_voucher') ? <Tag color="green">已上传</Tag> : <Tag>未上传</Tag>}</Descriptions.Item>
                    <Descriptions.Item label="签收回单">{valueOf(shipmentTask, 'signatureVoucher', 'signature_voucher') ? <Tag color="green">已上传</Tag> : <Tag>未上传</Tag>}</Descriptions.Item>
                  </Descriptions>
                )}
              </div>
            )}

            {pledgeEvents.length > 0 && (
              <div style={{ marginTop: 20 }}>
                <Typography.Text strong style={{ fontSize: 14 }}>货权事件</Typography.Text>
                <Timeline
                  style={{ marginTop: 12 }}
                  items={pledgeEvents.map((event) => ({
                    color: valueOf(event, 'eventType', 'event_type') === 'release' ? 'green' : 'blue',
                    children: (
                      <div>
                        <div style={{ fontWeight: 500 }}>{String(valueOf(event, 'eventType', 'event_type', '-'))}</div>
                        <Text type="secondary">{formatDateTime(valueOf<string>(event, 'createdAt', 'created_at'))}</Text>
                        <div><Text type="secondary">{String(event.reason || event.condition || '')}</Text></div>
                      </div>
                    ),
                  }))}
                />
              </div>
            )}

            {breachRecords.length > 0 && (
              <div style={{ marginTop: 20 }}>
                <Typography.Text strong style={{ fontSize: 14 }}>违约记录</Typography.Text>
                <Table
                  size="small"
                  rowKey="id"
                  style={{ marginTop: 12 }}
                  pagination={false}
                  dataSource={breachRecords}
                  columns={[
                    { title: '违约方', dataIndex: 'breachParty', width: 90, render: (v) => v === 'buyer' ? '买方' : '卖方' },
                    { title: '类型', dataIndex: 'breachType', width: 120 },
                    { title: '描述', dataIndex: 'description', ellipsis: true },
                    {
                      title: '状态',
                      dataIndex: 'status',
                      width: 100,
                      render: (v: string) => {
                        const meta = breachStatusMeta[v] || { color: 'default', text: v }
                        return <Tag color={meta.color}>{meta.text}</Tag>
                      },
                    },
                    { title: '创建时间', dataIndex: 'createdAt', width: 160, render: (v) => formatDateTime(v) },
                  ]}
                  expandable={{
                    expandedRowRender: (record) => (
                      <Space direction="vertical" style={{ width: '100%' }} size={12}>
                        {record.mediationResult ? <Alert type="info" showIcon message="协调结果" description={record.mediationResult} /> : null}
                        {record.resolutionResult ? <Alert type="success" showIcon message="处理结果" description={record.resolutionResult} /> : null}
                        <Input.TextArea
                          rows={2}
                          placeholder="填写平台协调或处理意见"
                          value={breachDrafts[record.id] || ''}
                          onChange={(e) => setBreachDrafts({ ...breachDrafts, [record.id]: e.target.value })}
                          maxLength={1000}
                          showCount
                        />
                        <Space>
                          <Button
                            loading={mediateMutation.isPending}
                            disabled={!(breachDrafts[record.id] || '').trim()}
                            onClick={() => mediateMutation.mutate({ id: record.id, result: (breachDrafts[record.id] || '').trim() })}
                          >
                            记录协调
                          </Button>
                          <Button
                            type="primary"
                            loading={resolveMutation.isPending}
                            disabled={!(breachDrafts[record.id] || '').trim()}
                            onClick={() => resolveMutation.mutate({ id: record.id, result: (breachDrafts[record.id] || '').trim() })}
                          >
                            完成处理
                          </Button>
                        </Space>
                      </Space>
                    ),
                  }}
                />
              </div>
            )}

            {litigationRecords.length > 0 && (
              <div style={{ marginTop: 20 }}>
                <Typography.Text strong style={{ fontSize: 14 }}>诉讼记录</Typography.Text>
                <Table
                  size="small"
                  rowKey="id"
                  style={{ marginTop: 12 }}
                  pagination={false}
                  dataSource={litigationRecords}
                  columns={[
                    { title: '案件号', dataIndex: 'caseNo', width: 140, render: (v) => v || '-' },
                    { title: '法院', dataIndex: 'court', width: 160, render: (v) => v || '-' },
                    { title: '诉讼金额', dataIndex: 'claimAmount', width: 120, render: (v) => `${money(v)} 元` },
                    {
                      title: '状态',
                      dataIndex: 'status',
                      width: 100,
                      render: (v: string) => {
                        const meta = litigationStatusMeta[v] || { color: 'default', text: v }
                        return <Tag color={meta.color}>{meta.text}</Tag>
                      },
                    },
                    { title: '创建时间', dataIndex: 'createdAt', width: 160, render: (v) => formatDateTime(v) },
                  ]}
                  expandable={{
                    expandedRowRender: (record) => {
                      const draft = litigationDrafts[record.id] || { status: record.status }
                      return (
                        <Space direction="vertical" style={{ width: '100%' }} size={12}>
                          <Descriptions bordered size="small" column={2}>
                            <Descriptions.Item label="原告企业">{record.plaintiffEnterpriseId}</Descriptions.Item>
                            <Descriptions.Item label="被告企业">{record.defendantEnterpriseId}</Descriptions.Item>
                            <Descriptions.Item label="诉讼描述" span={2}>{record.description}</Descriptions.Item>
                            <Descriptions.Item label="判决结果" span={2}>{record.judgmentResult || '-'}</Descriptions.Item>
                          </Descriptions>
                          <Form layout="inline">
                            <Form.Item label="状态">
                              <Select
                                style={{ width: 140 }}
                                value={draft.status}
                                options={litigationStatusOptions}
                                onChange={(status) => setLitigationDrafts({ ...litigationDrafts, [record.id]: { ...draft, status } })}
                              />
                            </Form.Item>
                            <Form.Item label="判决结果">
                              <Input
                                style={{ width: 240 }}
                                value={draft.judgmentResult}
                                onChange={(e) => setLitigationDrafts({ ...litigationDrafts, [record.id]: { ...draft, judgmentResult: e.target.value } })}
                              />
                            </Form.Item>
                            <Form.Item label="文书 URL">
                              <Input
                                style={{ width: 240 }}
                                value={draft.judgmentUrl}
                                onChange={(e) => setLitigationDrafts({ ...litigationDrafts, [record.id]: { ...draft, judgmentUrl: e.target.value } })}
                              />
                            </Form.Item>
                            <Form.Item>
                              <Button
                                type="primary"
                                loading={litigationMutation.isPending}
                                onClick={() => litigationMutation.mutate({ id: record.id, data: { status: draft.status || record.status, judgmentResult: draft.judgmentResult, judgmentUrl: draft.judgmentUrl } })}
                              >
                                更新
                              </Button>
                            </Form.Item>
                          </Form>
                        </Space>
                      )
                    },
                  }}
                />
              </div>
            )}

            {orderDetail.timeline && orderDetail.timeline.length > 0 && (
              <div style={{ marginTop: 24 }}>
                <Typography.Text strong style={{ fontSize: 14 }}>订单时间线</Typography.Text>
                <Timeline
                  style={{ marginTop: 12 }}
                  items={orderDetail.timeline.map((item: OrderTimelineItem) => ({
                    color: timelineColor(item.status),
                    children: (
                      <div>
                        <div style={{ fontWeight: 500 }}>{item.label}</div>
                        {item.time ? <Text type="secondary">{new Date(item.time).toLocaleString()}</Text> : null}
                        {item.description ? <div><Text type="secondary">{item.description}</Text></div> : null}
                      </div>
                    ),
                  }))}
                />
              </div>
            )}

            {orderEvents.length > 0 && (
              <>
                <Divider />
                <Typography.Text strong style={{ fontSize: 14 }}>关键事件</Typography.Text>
                <Timeline
                  style={{ marginTop: 12 }}
                  items={orderEvents.map((event) => {
                    const type = String(valueOf(event, 'eventType', 'event_type', ''))
                    const payload = (event.payload || {}) as Record<string, unknown>
                    return {
                      color: type.includes('rejected') || type.includes('no_capacity') ? 'red' : 'blue',
                      children: (
                        <div>
                          <div style={{ fontWeight: 500 }}>{eventLabel[type] || type}</div>
                          <Text type="secondary">{formatDateTime(valueOf<string>(event, 'createdAt', 'created_at'))}</Text>
                          {(payload.note || payload.reason || payload.voucher_url) && (
                            <div><Text type="secondary">{String(payload.note || payload.reason || payload.voucher_url)}</Text></div>
                          )}
                        </div>
                      ),
                    }
                  })}
                />
              </>
            )}

            {orderDetail.status >= 12 && (
              <div style={{ marginTop: 24 }}>
                <Typography.Text strong style={{ fontSize: 14 }}>交易履约说明</Typography.Text>
                <div style={{ marginTop: 8, color: '#666', fontSize: 13, lineHeight: 1.8 }}>
                  {orderDetail.status === 12 && '订单等待买方选择目的地仓储。'}
                  {orderDetail.status === 13 && '订单已进入仓库确认阶段，仓储方需核验仓库条件并反馈。'}
                  {orderDetail.status === 14 && '仓储方已确认仓库条件，等待卖方确认价格、数量和交付条款。'}
                  {orderDetail.status === 15 && '卖方已确认条款，等待买方最终确认订单。'}
                  {orderDetail.status === 22 && '买方已申请托盘融资，资金方正在进行风控审批。'}
                  {orderDetail.status === 23 && '融资已审批通过，等待买方签署债务合同并支付手续费。'}
                  {orderDetail.status === 24 && '仓储方已拒绝当前仓储条件，平台需跟进并引导买方重新选仓。'}
                  {orderDetail.status === 16 && '卖方已发起物流竞价，等待物流方报价。'}
                  {orderDetail.status === 17 && '物流方已报价，等待合同约定的运费承担方确认报价。'}
                  {orderDetail.status === 25 && '物流反馈暂无运力，平台需跟进并引导卖方重新发起物流。'}
                  {orderDetail.status === 18 && '物流资源已锁定，运输任务进行中。'}
                  {orderDetail.status === 19 && '买方已确认收货，等待卖方对账确认。'}
                  {orderDetail.status === 20 && '卖方已完成对账确认，等待归档。'}
                  {orderDetail.status === 21 && '订单已归档，交易履约流程完成。'}
                  {orderDetail.status === 26 && '订单等待买方选择可用仓库。'}
                  {orderDetail.status === 27 && '订单等待物流合同签署完成。'}
                </div>
              </div>
            )}
          </>
        )}
      </Drawer>
    </>
  )
}
