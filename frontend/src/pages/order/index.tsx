import { useState } from 'react'
import { Alert, Card, Table, Tag, Input, Select, Space, Button, Descriptions, Drawer, Typography, Timeline, Divider } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useQuery } from '@tanstack/react-query'
import { LogisticsQuoteInfo, orderApi, OrderEventItem, OrderItem, OrderTimelineItem, PageParams } from '../../services/api'
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
  12: { color: 'purple', text: '待仓库确认' },
  13: { color: 'orange', text: '待卖方确认' },
  14: { color: 'blue', text: '待买方确认' },
  15: { color: 'cyan', text: '物流竞价中' },
  16: { color: 'gold', text: '待确认物流' },
  17: { color: 'processing', text: '运输中' },
  18: { color: 'green', text: '已收货' },
  19: { color: 'success', text: '已对账' },
  20: { color: 'default', text: '已归档' },
  21: { color: 'purple', text: '融资审批中' },
  22: { color: 'gold', text: '待签署融资合同' },
  23: { color: 'red', text: '待重新选仓' },
  24: { color: 'red', text: '待重新发起物流' },
}

function timelineColor(status: string) {
  if (status === 'completed') return 'green'
  if (status === 'current') return 'blue'
  if (status === 'skipped') return 'gray'
  return 'gray'
}

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待仓库确认', value: '12' },
  { label: '待卖方确认', value: '13' },
  { label: '待买方确认', value: '14' },
  { label: '待签合同', value: '2' },
  { label: '待支付', value: '3' },
  { label: '融资审批中', value: '21' },
  { label: '待签署融资合同', value: '22' },
  { label: '待重新选仓', value: '23' },
  { label: '已支付', value: '4' },
  { label: '物流竞价中', value: '15' },
  { label: '待确认物流', value: '16' },
  { label: '待重新发起物流', value: '24' },
  { label: '运输中', value: '17' },
  { label: '已送达', value: '7' },
  { label: '已收货', value: '18' },
  { label: '已对账', value: '19' },
  { label: '已归档', value: '20' },
  { label: '已完成', value: '8' },
  { label: '已取消', value: '9' },
  { label: '异常', value: '11' },
]

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
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [selectedOrder, setSelectedOrder] = useState<OrderItem | null>(null)

  const queryParams = { ...params, keyword: keyword || undefined, status: statusFilter ? statusFilter as unknown as number : undefined }

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
            {(remark.admin_attention || orderDetail.status === 23 || orderDetail.status === 24) && (
              <Alert
                type="warning"
                showIcon
                message="平台需处理事项"
                description={
                  orderDetail.status === 23
                    ? `仓储已拒绝，需协助买方重新选择仓储。原因：${remark.warehouse_reject_note || remark.condition_note || '-'}`
                    : orderDetail.status === 24
                      ? `物流反馈无运力，需协助卖方重新发起物流。原因：${remark.logistics_no_capacity_reason || '-'}`
                      : '该订单标记为需要平台关注，请查看下方事件记录。'
                }
                style={{ marginBottom: 16 }}
              />
            )}
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
                  {orderDetail.status === 12 && '订单已进入仓库确认阶段，仓储方需核验仓库条件并反馈。'}
                  {orderDetail.status === 13 && '仓储方已确认仓库条件，等待卖方确认价格、数量和交付条款。'}
                  {orderDetail.status === 14 && '卖方已确认条款，等待买方最终确认订单。'}
                  {orderDetail.status === 21 && '买方已申请托盘融资，资金方正在进行风控审批。'}
                  {orderDetail.status === 22 && '融资已审批通过，等待买方签署债务合同并支付手续费。'}
                  {orderDetail.status === 23 && '仓储方已拒绝当前仓储条件，平台需跟进并引导买方重新选仓。'}
                  {orderDetail.status === 15 && '卖方已发起物流竞价，等待物流方报价。'}
                  {orderDetail.status === 16 && '物流方已报价，等待合同约定的运费承担方确认报价。'}
                  {orderDetail.status === 24 && '物流反馈暂无运力，平台需跟进并引导卖方重新发起物流。'}
                  {orderDetail.status === 17 && '物流资源已锁定，运输任务进行中。'}
                  {orderDetail.status === 18 && '买方已确认收货，等待卖方对账确认。'}
                  {orderDetail.status === 19 && '卖方已完成对账确认，等待归档。'}
                  {orderDetail.status === 20 && '订单已归档，交易履约流程完成。'}
                </div>
              </div>
            )}
          </>
        )}
      </Drawer>
    </>
  )
}
