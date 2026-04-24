import { useState } from 'react'
import { Card, Table, Tag, Timeline, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useQuery } from '@tanstack/react-query'
import { orderApi, OrderItem, PageParams } from '../../services/api'

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

export default function OrderList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })

  const { data, isLoading } = useQuery({
    queryKey: ['orders', params],
    queryFn: () => orderApi.list(params),
  })

  const columns: ColumnsType<OrderItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '订单号', dataIndex: 'orderNo', key: 'orderNo', width: 220 },
    { title: '买方', dataIndex: 'buyerName', key: 'buyerName', width: 160 },
    { title: '卖方', dataIndex: 'sellerName', key: 'sellerName', width: 160 },
    {
      title: '数量',
      dataIndex: 'totalQuantity',
      key: 'totalQuantity',
      render: (value: number) => `${value} 吨`,
    },
    {
      title: '金额',
      dataIndex: 'totalAmount',
      key: 'totalAmount',
      render: (value: number, record) => `${record.currency || 'CNY'} ${Number(value || 0).toFixed(2)}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: number) => {
        const meta = orderStatusMeta[status] || { color: 'default', text: `状态${status}` }
        return <Tag color={meta.color}>{meta.text}</Tag>
      },
    },
    {
      title: '履约',
      key: 'delivery',
      render: (_, record) => (
        <div>
          <div><Text type="secondary">出库凭证：</Text><Text>{record.deliveryVoucher ? '已上传' : '未上传'}</Text></div>
          <div><Text type="secondary">签收凭证：</Text><Text>{record.receiptVoucher ? '已上传' : '未上传'}</Text></div>
        </div>
      ),
    },
    { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 180 },
  ]

  return (
    <Card title="订单管理">
      <Table
        columns={columns}
        dataSource={data?.data.data.list || []}
        rowKey="id"
        loading={isLoading}
        expandable={{
          expandedRowRender: (record) => (
            <Timeline
              items={(record.timeline || []).map((item) => ({
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
          onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
        }}
      />
    </Card>
  )
}
