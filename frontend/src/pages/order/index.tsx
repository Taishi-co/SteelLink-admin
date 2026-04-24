import { useState } from 'react'
import { Card, Table, Tag, Input, Select, Space, Button, Modal, Descriptions, Drawer, Spin, Typography, Timeline } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useQuery } from '@tanstack/react-query'
import { orderApi, OrderItem, OrderTimelineItem, PageParams } from '../../services/api'
import { formatDateTime } from '../../utils/format'
import { EyeOutlined, SearchOutlined } from '@ant-design/icons'

const { Text } = Typography

const orderStatusMeta: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '草稿' },
  1: { color: 'orange', text: '待确认' },
  2: { color: 'gold', text: '待签合同' },
  3: { color: 'blue', text: '待支付' },
  4: { color: 'cyan', text: '待收款确认' },
  5: { color: 'purple', text: '待提货' },
  6: { color: 'processing', text: '配送中' },
  7: { color: 'green', text: '已送达' },
  8: { color: 'green', text: '已完成' },
  9: { color: 'red', text: '已取消' },
}

function timelineColor(status: string) {
  if (status === 'completed') return 'green'
  if (status === 'current') return 'blue'
  if (status === 'skipped') return 'gray'
  return 'gray'
}

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待支付', value: 'pending_payment' },
  { label: '待收款确认', value: 'pending_payment_audit' },
  { label: '待提货', value: 'pending_pickup' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' },
]

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

  const columns: ColumnsType<OrderItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: '订单号', dataIndex: 'orderNo', key: 'orderNo', width: 200 },
    {
      title: '买方', key: 'buyer', width: 160,
      render: (_, r) => (
        <div>
          <div>{r.buyerName || '-'}</div>
          <Text type="secondary" style={{ fontSize: 12 }}>ID: {r.buyerId}</Text>
        </div>
      ),
    },
    {
      title: '卖方', key: 'seller', width: 160,
      render: (_, r) => (
        <div>
          <div>{r.sellerName || '-'}</div>
          <Text type="secondary" style={{ fontSize: 12 }}>ID: {r.sellerId}</Text>
        </div>
      ),
    },
    {
      title: '数量', dataIndex: 'totalQuantity', key: 'totalQuantity', width: 100,
      render: (v: number) => `${v} 吨`,
    },
    {
      title: '金额', dataIndex: 'totalAmount', key: 'totalAmount', width: 130,
      render: (v: number, r) => `${r.currency || 'CNY'} ${Number(v || 0).toLocaleString()}`,
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 110,
      render: (s: number) => {
        const meta = orderStatusMeta[s] || { color: 'default', text: `状态${s}` }
        return <Tag color={meta.color}>{meta.text}</Tag>
      },
    },
    {
      title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 170,
      render: (v: string) => formatDateTime(v),
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
        title={`订单详情 · ${selectedOrder?.orderNo || ''}`}
        width={720}
        open={selectedOrder !== null}
        onClose={() => setSelectedOrder(null)}
        destroyOnClose
      >
        {selectedOrder && (
          <>
            <Descriptions bordered column={2} size="small">
              <Descriptions.Item label="订单号" span={2}>{selectedOrder.orderNo}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={orderStatusMeta[selectedOrder.status]?.color}>{orderStatusMeta[selectedOrder.status]?.text}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="币种">{selectedOrder.currency || 'CNY'}</Descriptions.Item>
            </Descriptions>

            <Descriptions bordered column={2} size="small" title="买方信息" style={{ marginTop: 20 }}>
              <Descriptions.Item label="企业名称">{selectedOrder.buyerName || '-'}</Descriptions.Item>
              <Descriptions.Item label="企业 ID">{selectedOrder.buyerId}</Descriptions.Item>
            </Descriptions>

            <Descriptions bordered column={2} size="small" title="卖方信息" style={{ marginTop: 20 }}>
              <Descriptions.Item label="企业名称">{selectedOrder.sellerName || '-'}</Descriptions.Item>
              <Descriptions.Item label="企业 ID">{selectedOrder.sellerId}</Descriptions.Item>
              <Descriptions.Item label="卖家企业 ID">{selectedOrder.sellerEnterpriseId}</Descriptions.Item>
            </Descriptions>

            <Descriptions bordered column={2} size="small" title="交易信息" style={{ marginTop: 20 }}>
              <Descriptions.Item label="总数量">{selectedOrder.totalQuantity} 吨</Descriptions.Item>
              <Descriptions.Item label="总金额">{Number(selectedOrder.totalAmount || 0).toLocaleString()} 元</Descriptions.Item>
              <Descriptions.Item label="含税金额">{Number(selectedOrder.taxTotalAmount || 0).toLocaleString()} 元</Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatDateTime(selectedOrder.createdAt)}</Descriptions.Item>
              <Descriptions.Item label="过期时间">{formatDateTime(selectedOrder.expiresAt)}</Descriptions.Item>
            </Descriptions>

            <Descriptions bordered column={2} size="small" title="履约信息" style={{ marginTop: 20 }}>
              <Descriptions.Item label="出库凭证">{selectedOrder.deliveryVoucher ? <Tag color="blue">已上传</Tag> : <Tag>未上传</Tag>}</Descriptions.Item>
              <Descriptions.Item label="签收凭证">{selectedOrder.receiptVoucher ? <Tag color="green">已上传</Tag> : <Tag>未上传</Tag>}</Descriptions.Item>
              <Descriptions.Item label="签收时间">{formatDateTime(selectedOrder.receiptAt)}</Descriptions.Item>
            </Descriptions>

            {selectedOrder.timeline && selectedOrder.timeline.length > 0 && (
              <div style={{ marginTop: 24 }}>
                <Typography.Text strong style={{ fontSize: 14 }}>订单时间线</Typography.Text>
                <Timeline
                  style={{ marginTop: 12 }}
                  items={selectedOrder.timeline.map((item: OrderTimelineItem) => ({
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
          </>
        )}
      </Drawer>
    </>
  )
}
