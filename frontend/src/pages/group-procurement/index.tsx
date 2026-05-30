import { useState } from 'react'
import { Button, Card, Col, Descriptions, Drawer, Form, Input, InputNumber, Modal, Popconfirm, Row, Select, Space, Statistic, Steps, Table, Tag, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CheckCircleOutlined, ClockCircleOutlined, EyeOutlined, FileDoneOutlined, PlusOutlined, ReloadOutlined, SearchOutlined, SendOutlined, ShopOutlined, StopOutlined, TeamOutlined } from '@ant-design/icons'
import { groupProcurementApi, GroupParticipantItem, GroupProcurementItem, GroupProcurementPageParams, GroupProcurementRequestItem } from '../../services/api'
import { formatDateTime } from '../../utils/format'

const { Text } = Typography

const statusMeta: Record<string, { color: string; text: string }> = {
  pending_seller: { color: 'gold', text: '待卖家确认' },
  active: { color: 'green', text: '商城参团中' },
  ended: { color: 'blue', text: '已截止' },
  order_created: { color: 'purple', text: '已生成订单' },
  seller_rejected: { color: 'red', text: '卖家拒绝' },
  closed: { color: 'default', text: '已关闭' },
}

const requestStatusMeta: Record<string, { color: string; text: string }> = {
  pending: { color: 'gold', text: '待处理' },
  accepted: { color: 'green', text: '已采纳' },
  rejected: { color: 'red', text: '已拒绝' },
}

const participantStatusMeta: Record<string, { color: string; text: string }> = {
  joined: { color: 'green', text: '已参团' },
  order_created: { color: 'purple', text: '已成单' },
  cancelled: { color: 'default', text: '已取消' },
}

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待卖家确认', value: 'pending_seller' },
  { label: '商城参团中', value: 'active' },
  { label: '已截止', value: 'ended' },
  { label: '已生成订单', value: 'order_created' },
  { label: '卖家拒绝', value: 'seller_rejected' },
]

const flowSteps = [
  { title: '平台发起', icon: <SendOutlined /> },
  { title: '卖家确认', icon: <CheckCircleOutlined /> },
  { title: '买家参团', icon: <TeamOutlined /> },
  { title: '管理成单', icon: <FileDoneOutlined /> },
  { title: '订单履约', icon: <ShopOutlined /> },
]

interface GroupProcurementFilterValues {
  keyword?: string
  status?: string
  category?: string
}

function statusTag(status: string) {
  const meta = statusMeta[status] || { color: 'default', text: status || '-' }
  return <Tag color={meta.color}>{meta.text}</Tag>
}

function participantStatusTag(status: string) {
  const meta = participantStatusMeta[status] || { color: 'default', text: status || '-' }
  return <Tag color={meta.color}>{meta.text}</Tag>
}

function money(value: number) {
  return `¥${Number(value || 0).toLocaleString()}`
}

function procurementStep(status?: string) {
  if (status === 'pending_seller' || status === 'seller_rejected') return 1
  if (status === 'active' || status === 'ended') return 2
  if (status === 'order_created') return 4
  return 0
}

function getErrorMessage(error: unknown, fallback: string) {
  const maybe = error as { response?: { data?: { message?: string } }; message?: string }
  return maybe.response?.data?.message || maybe.message || fallback
}

export default function GroupProcurementPage() {
  const queryClient = useQueryClient()
  const [filterForm] = Form.useForm<GroupProcurementFilterValues>()
  const [params, setParams] = useState<GroupProcurementPageParams>({ page: 1, pageSize: 10 })
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [createForm] = Form.useForm<{ productId: number; title?: string; targetQuantity?: number; durationHours?: number; adminRemark?: string }>()

  const { data, isLoading } = useQuery({
    queryKey: ['group-procurements', params],
    queryFn: () => groupProcurementApi.list(params),
  })
  const { data: detailData, isFetching: detailLoading } = useQuery({
    queryKey: ['group-procurement-detail', selectedId],
    queryFn: () => groupProcurementApi.get(selectedId!),
    enabled: selectedId !== null,
  })
  const { data: requestData, isLoading: requestLoading } = useQuery({
    queryKey: ['group-procurement-requests'],
    queryFn: () => groupProcurementApi.listRequests({ page: 1, pageSize: 8, status: 'pending' }),
  })

  const createMutation = useMutation({
    mutationFn: groupProcurementApi.create,
    onSuccess: () => {
      message.success('集采申请已发送给卖家')
      setCreateOpen(false)
      createForm.resetFields()
      void queryClient.invalidateQueries({ queryKey: ['group-procurements'] })
    },
    onError: (error) => message.error(getErrorMessage(error, '发起集采失败')),
  })

  const endMutation = useMutation({
    mutationFn: ({ id }: { id: number }) => groupProcurementApi.end(id),
    onSuccess: () => {
      message.success('集采已结束并生成订单')
      void queryClient.invalidateQueries({ queryKey: ['group-procurements'] })
      if (selectedId) void queryClient.invalidateQueries({ queryKey: ['group-procurement-detail', selectedId] })
    },
    onError: (error) => message.error(getErrorMessage(error, '结束集采失败')),
  })

  const rows = data?.data.data.list ?? []
  const detail = detailData?.data.data
  const requests = requestData?.data.data.list ?? []
  const activeCount = rows.filter((item) => item.status === 'active').length
  const pendingSellerCount = rows.filter((item) => item.status === 'pending_seller').length
  const orderedCount = rows.filter((item) => item.status === 'order_created').length
  const demandQuantity = rows.reduce((sum, item) => sum + Number(item.demandQuantity || 0), 0)

  const columns: ColumnsType<GroupProcurementItem> = [
    { title: '编号', dataIndex: 'procurementNo', width: 180 },
    { title: '标题', dataIndex: 'title', width: 260, render: (value, row) => <div><Text strong>{value}</Text><div><Text type="secondary">{row.stage || '-'}</Text></div></div> },
    { title: '商品', key: 'product', width: 220, render: (_, row) => <div><div>{row.productName}</div><Text type="secondary">#{row.productId} {row.productNo}</Text></div> },
    { title: '卖家', dataIndex: 'sellerName', width: 180 },
    { title: '状态', dataIndex: 'status', width: 130, render: statusTag },
    { title: '集采价', dataIndex: 'groupPrice', width: 120, render: money },
    { title: '参团', key: 'join', width: 140, render: (_, row) => `${row.participatingCount} 家 / ${row.demandQuantity}${row.unit || '吨'}` },
    { title: '截止时间', dataIndex: 'endedAt', width: 170, render: formatDateTime },
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 160,
      render: (_, row) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setSelectedId(row.id)}>详情</Button>
          {(row.status === 'active' || row.status === 'ended') && !row.masterOrderId ? (
            <Popconfirm title="结束集采并生成订单？" okText="确认结束" cancelText="取消" onConfirm={() => endMutation.mutate({ id: row.id })}>
              <Button type="link" danger size="small" icon={<StopOutlined />}>结束</Button>
            </Popconfirm>
          ) : null}
        </Space>
      ),
    },
  ]

  const participantColumns: ColumnsType<GroupParticipantItem> = [
    { title: '买家', dataIndex: 'buyerName', width: 180 },
    { title: '数量', key: 'quantity', width: 120, render: (_, row) => `${row.quantity}${row.unit || '吨'}` },
    { title: '金额', dataIndex: 'totalAmount', width: 120, render: money },
    { title: '联系人', key: 'contact', width: 160, render: (_, row) => `${row.contactName} ${row.contactPhone}` },
    { title: '收货地址', dataIndex: 'deliveryAddress', width: 260 },
    { title: '状态', dataIndex: 'status', width: 110, render: participantStatusTag },
    { title: '子订单', dataIndex: 'childOrderNo', width: 170, render: (value) => value || '-' },
  ]

  const requestColumns: ColumnsType<GroupProcurementRequestItem> = [
    { title: '申请编号', dataIndex: 'requestNo', width: 170 },
    { title: '来源', dataIndex: 'source', width: 100 },
    { title: '品类', dataIndex: 'category', width: 120 },
    { title: '采购物品', dataIndex: 'productName', width: 180 },
    { title: '需求量', key: 'quantity', width: 120, render: (_, row) => `${row.quantity}${row.unit || '吨'}` },
    { title: '申请企业', dataIndex: 'buyerName', width: 180 },
    { title: '状态', dataIndex: 'status', width: 110, render: (value) => {
      const meta = requestStatusMeta[value] || { color: 'default', text: value }
      return <Tag color={meta.color}>{meta.text}</Tag>
    } },
    { title: '提交时间', dataIndex: 'createdAt', width: 170, render: formatDateTime },
  ]

  const applyFilters = (values: GroupProcurementFilterValues) => {
    setParams({
      ...params,
      page: 1,
      keyword: values.keyword,
      status: values.status,
      category: values.category,
    })
  }

  const resetFilters = () => {
    filterForm.resetFields()
    setParams({ page: 1, pageSize: params.pageSize })
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Row gutter={16}>
        <Col xs={24} md={6}><Card><Statistic title="待卖家确认" value={pendingSellerCount} suffix="场" prefix={<ClockCircleOutlined />} /></Card></Col>
        <Col xs={24} md={6}><Card><Statistic title="商城参团中" value={activeCount} suffix="场" prefix={<TeamOutlined />} /></Card></Col>
        <Col xs={24} md={6}><Card><Statistic title="已生成订单" value={orderedCount} suffix="场" prefix={<FileDoneOutlined />} /></Card></Col>
        <Col xs={24} md={6}><Card><Statistic title="当前页归集量" value={demandQuantity} suffix="吨" precision={2} prefix={<ShopOutlined />} /></Card></Col>
      </Row>

      <Card title="集采管理" extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>发起集采</Button>}>
        <Form form={filterForm} layout="vertical" onFinish={applyFilters} initialValues={params} style={{ marginBottom: 16 }}>
          <Row gutter={12}>
            <Col xs={24} lg={8}>
              <Form.Item name="keyword" label="搜索">
                <Input allowClear prefix={<SearchOutlined />} placeholder="编号 / 标题 / 商品 / 卖家" />
              </Form.Item>
            </Col>
            <Col xs={12} lg={5}>
              <Form.Item name="status" label="状态">
                <Select allowClear placeholder="全部状态" options={statusOptions.filter((item) => item.value !== '')} />
              </Form.Item>
            </Col>
            <Col xs={12} lg={5}>
              <Form.Item name="category" label="品类">
                <Input allowClear placeholder="如 热轧板卷" />
              </Form.Item>
            </Col>
            <Col xs={24} lg={6}>
              <Form.Item label="操作">
                <Space wrap>
                  <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>查询</Button>
                  <Button icon={<ReloadOutlined />} onClick={resetFilters}>重置</Button>
                </Space>
              </Form.Item>
            </Col>
          </Row>
        </Form>
        <Table
          rowKey="id"
          loading={isLoading}
          columns={columns}
          dataSource={rows}
          scroll={{ x: 1500 }}
          pagination={{
            current: params.page,
            pageSize: params.pageSize,
            total: data?.data.data.total || 0,
            onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
          }}
        />
      </Card>

      <Card title="待处理集采申请">
        <Table rowKey="id" loading={requestLoading} columns={requestColumns} dataSource={requests} scroll={{ x: 1000 }} pagination={false} />
      </Card>

      <Drawer
        title={`集采详情 ${detail?.procurementNo || ''}`}
        open={selectedId !== null}
        onClose={() => setSelectedId(null)}
        width={980}
      >
        {detail ? (
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            <Card size="small">
              <Steps
                current={procurementStep(detail.status)}
                items={flowSteps.map((item) => ({ title: item.title, icon: item.icon }))}
              />
            </Card>
            <Row gutter={12}>
              <Col span={8}><Card size="small"><Statistic title="参团企业" value={detail.participatingCount} suffix="家" /></Card></Col>
              <Col span={8}><Card size="small"><Statistic title="归集需求" value={detail.demandQuantity} suffix={detail.unit || '吨'} precision={2} /></Card></Col>
              <Col span={8}><Card size="small"><Statistic title="成团门槛" value={detail.minGroupQuantity} suffix={detail.unit || '吨'} precision={2} /></Card></Col>
            </Row>
            <Descriptions bordered column={2} size="small">
              <Descriptions.Item label="状态">{statusTag(detail.status)}</Descriptions.Item>
              <Descriptions.Item label="阶段">{detail.stage || '-'}</Descriptions.Item>
              <Descriptions.Item label="商品">{detail.productName}</Descriptions.Item>
              <Descriptions.Item label="卖家">{detail.sellerName}</Descriptions.Item>
              <Descriptions.Item label="集采价">{money(detail.groupPrice)}</Descriptions.Item>
              <Descriptions.Item label="最小成团量">{detail.minGroupQuantity}{detail.unit}</Descriptions.Item>
              <Descriptions.Item label="参团需求">{detail.demandQuantity}{detail.unit}</Descriptions.Item>
              <Descriptions.Item label="参团企业">{detail.participatingCount} 家</Descriptions.Item>
              <Descriptions.Item label="截止时间">{formatDateTime(detail.endedAt)}</Descriptions.Item>
              <Descriptions.Item label="母订单">{detail.masterOrderNo || '-'}</Descriptions.Item>
              <Descriptions.Item label="卖家备注" span={2}>{detail.sellerRemark || '-'}</Descriptions.Item>
              <Descriptions.Item label="管理备注" span={2}>{detail.adminRemark || '-'}</Descriptions.Item>
            </Descriptions>
            {(detail.status === 'active' || detail.status === 'ended') && !detail.masterOrderId ? (
              <Popconfirm title="结束集采并生成订单？" okText="确认结束" cancelText="取消" onConfirm={() => endMutation.mutate({ id: detail.id })}>
                <Button danger icon={<StopOutlined />} loading={endMutation.isPending}>
                  结束集采并生成订单
                </Button>
              </Popconfirm>
            ) : null}
            <Table
              rowKey="id"
              loading={detailLoading}
              title={() => '参团买家'}
              columns={participantColumns}
              dataSource={detail.participants || []}
              scroll={{ x: 1200 }}
              pagination={false}
            />
          </Space>
        ) : null}
      </Drawer>

      <Modal
        title="发起商品集采"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        okText="发送给卖家"
        confirmLoading={createMutation.isPending}
        onOk={() => {
          createForm.validateFields().then((values) => createMutation.mutate(values))
        }}
      >
        <Form form={createForm} layout="vertical" initialValues={{ durationHours: 24 }}>
          <Form.Item name="productId" label="商品 ID" rules={[{ required: true, message: '请输入商品 ID' }]}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="title" label="集采标题">
            <Input placeholder="不填则按商品名称生成" />
          </Form.Item>
          <Form.Item name="targetQuantity" label="目标数量（吨）">
            <InputNumber min={0.01} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="durationHours" label="默认时限（小时）" rules={[{ required: true }]}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="adminRemark" label="管理备注">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  )
}
