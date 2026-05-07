import { useState } from 'react'
import { Card, Table, Tag, Input, Select, Space, Button, Descriptions, Drawer, Typography, Timeline } from 'antd'
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
  { label: '已支付', value: '4' },
  { label: '物流竞价中', value: '15' },
  { label: '待确认物流', value: '16' },
  { label: '运输中', value: '17' },
  { label: '已送达', value: '7' },
  { label: '已收货', value: '18' },
  { label: '已对账', value: '19' },
  { label: '已归档', value: '20' },
  { label: '已完成', value: '8' },
  { label: '已取消', value: '9' },
  { label: '异常', value: '11' },
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

            {selectedOrder.status >= 12 && (
              <div style={{ marginTop: 24 }}>
                <Typography.Text strong style={{ fontSize: 14 }}>交易履约说明</Typography.Text>
                <div style={{ marginTop: 8, color: '#666', fontSize: 13, lineHeight: 1.8 }}>
                  {selectedOrder.status === 12 && '订单已进入仓库确认阶段，仓储方需核验仓库条件并反馈。'}
                  {selectedOrder.status === 13 && '仓储方已确认仓库条件，等待卖方确认价格、数量和交付条款。'}
                  {selectedOrder.status === 14 && '卖方已确认条款，等待买方最终确认订单。'}
                  {(selectedOrder.status === 21) && '买方已申请托盘融资，资金方正在进行风控审批。'}
                  {(selectedOrder.status === 22) && '融资已审批通过，等待买方签署债务合同并支付手续费。'}
                  {selectedOrder.status === 15 && '卖方已发起物流竞价，等待物流方报价。'}
                  {selectedOrder.status === 16 && '物流方已报价，等待买卖双方确认同一物流资源。'}
                  {selectedOrder.status === 17 && '物流资源已锁定，运输任务进行中。'}
                  {selectedOrder.status === 18 && '买方已确认收货，等待卖方对账确认。'}
                  {selectedOrder.status === 19 && '卖方已完成对账确认，等待归档。'}
                  {selectedOrder.status === 20 && '订单已归档，交易履约流程完成。'}
                </div>
              </div>
            )}
          </>
        )}
      </Drawer>
    </>
  )
}
