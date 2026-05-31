import { useState } from 'react'
import { Button, Card, Col, Descriptions, Drawer, Form, Input, InputNumber, Modal, Popconfirm, Row, Select, Space, Statistic, Steps, Table, Tag, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CheckCircleOutlined, ClockCircleOutlined, EyeOutlined, FileDoneOutlined, PlusOutlined, ReloadOutlined, SearchOutlined, SendOutlined, ShopOutlined, StopOutlined, TeamOutlined } from '@ant-design/icons'
import { groupProcurementApi, GroupParticipantItem, GroupProcurementItem, GroupProcurementPageParams, GroupProcurementRequestItem, PageParams, productApi, ProductItem } from '../../services/api'
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

const orderStatusMeta: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '未生成' },
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
  12: { color: 'purple', text: '待选目的仓' },
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
  23: { color: 'gold', text: '待融资合同' },
  24: { color: 'red', text: '待重新选仓' },
  25: { color: 'red', text: '物流无运力' },
  26: { color: 'purple', text: '待选择仓库' },
  27: { color: 'gold', text: '待物流合同' },
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

function orderStatusTag(status?: number) {
  const meta = orderStatusMeta[Number(status || 0)] || { color: 'default', text: `状态${status}` }
  return <Tag color={meta.color}>{meta.text}</Tag>
}

function money(value: number) {
  return `¥${Number(value || 0).toLocaleString()}`
}

function quantity(value: number, unit?: string) {
  return `${Number(value || 0).toLocaleString()} ${unit || '吨'}`
}

function availableStock(product?: ProductItem) {
  return product?.availableQty ?? product?.stock ?? 0
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
  const [sourceRequest, setSourceRequest] = useState<GroupProcurementRequestItem | null>(null)
  const [selectedProduct, setSelectedProduct] = useState<ProductItem | null>(null)
  const [productKeyword, setProductKeyword] = useState('')
  const [productParams, setProductParams] = useState<PageParams>({ page: 1, pageSize: 5, status: 2 })
  const [createForm] = Form.useForm<{ requestId?: number; productId?: number; title?: string; targetQuantity?: number; durationHours?: number; adminRemark?: string }>()

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
  const { data: productData, isLoading: productLoading } = useQuery({
    queryKey: ['group-procurement-products', productParams],
    queryFn: () => productApi.list(productParams),
    enabled: createOpen,
  })

  const createMutation = useMutation({
    mutationFn: groupProcurementApi.create,
    onSuccess: () => {
      message.success('集采申请已发送给卖家')
      setCreateOpen(false)
      setSourceRequest(null)
      setSelectedProduct(null)
      setProductKeyword('')
      createForm.resetFields()
      void queryClient.invalidateQueries({ queryKey: ['group-procurements'] })
      void queryClient.invalidateQueries({ queryKey: ['group-procurement-requests'] })
    },
    onError: (error) => message.error(getErrorMessage(error, '发起集采失败')),
  })

  const rejectRequestMutation = useMutation({
    mutationFn: ({ id }: { id: number }) => groupProcurementApi.rejectRequest(id),
    onSuccess: () => {
      message.success('集采申请已驳回')
      void queryClient.invalidateQueries({ queryKey: ['group-procurement-requests'] })
    },
    onError: (error) => message.error(getErrorMessage(error, '驳回集采申请失败')),
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
  const products = productData?.data.data.list ?? []
  const activeCount = rows.filter((item) => item.status === 'active').length
  const pendingSellerCount = rows.filter((item) => item.status === 'pending_seller').length
  const orderedCount = rows.filter((item) => item.status === 'order_created').length
  const demandQuantity = rows.reduce((sum, item) => sum + Number(item.demandQuantity || 0), 0)

  const openCreateModal = (request?: GroupProcurementRequestItem) => {
    const nextParams: PageParams = {
      page: 1,
      pageSize: 5,
      status: 2,
      keyword: request ? request.productName || request.category : undefined,
    }
    setProductParams(nextParams)
    setProductKeyword(nextParams.keyword || '')
    setSourceRequest(request || null)
    setSelectedProduct(null)
    createForm.setFieldsValue({
      requestId: request?.id,
      productId: undefined,
      title: request ? `平台集采：${request.productName}` : undefined,
      targetQuantity: request?.quantity,
      durationHours: 24,
      adminRemark: request?.remark,
    })
    setCreateOpen(true)
  }

  const closeCreateModal = () => {
    setCreateOpen(false)
    setSourceRequest(null)
    setSelectedProduct(null)
    setProductKeyword('')
    createForm.resetFields()
  }

  const selectProduct = (product: ProductItem) => {
    setSelectedProduct(product)
    createForm.setFieldsValue({
      productId: product.id,
      title: createForm.getFieldValue('title') || `平台集采：${product.name}`,
      targetQuantity: createForm.getFieldValue('targetQuantity') || product.minOrderQty || 1,
    })
  }

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
    { title: '订单进度', key: 'childOrderStatus', width: 140, render: (_, row) => row.childOrderId ? orderStatusTag(row.childOrderStatus) : '-' },
    { title: '订单更新', dataIndex: 'childOrderUpdatedAt', width: 170, render: formatDateTime },
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 110,
      render: (_, row) => row.childOrderNo ? (
        <Button type="link" size="small" href={`/product/order?keyword=${encodeURIComponent(row.childOrderNo)}`} target="_blank">
          查看订单
        </Button>
      ) : '-',
    },
  ]

  const productColumns: ColumnsType<ProductItem> = [
    {
      title: '商品',
      key: 'product',
      width: 260,
      render: (_, row) => (
        <div>
          <Text strong>{row.name}</Text>
          <div><Text type="secondary">#{row.id} {row.productNo || '-'}</Text></div>
        </div>
      ),
    },
    { title: '卖家企业', dataIndex: 'enterpriseId', width: 100, render: (value) => value ? `#${value}` : '-' },
    { title: '品类', key: 'category', width: 160, render: (_, row) => row.subCategory || row.category || '-' },
    { title: '价格', dataIndex: 'price', width: 110, render: money },
    { title: '可用库存', key: 'availableQty', width: 120, render: (_, row) => quantity(availableStock(row), row.unit) },
    { title: '最小起订', key: 'minOrderQty', width: 120, render: (_, row) => quantity(row.minOrderQty || 0, row.unit) },
    { title: '仓库', key: 'warehouse', width: 180, render: (_, row) => row.warehouseName || [row.province, row.city, row.district].filter(Boolean).join('/') || '-' },
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 90,
      render: (_, row) => (
        <Button type={selectedProduct?.id === row.id ? 'primary' : 'link'} size="small" onClick={() => selectProduct(row)}>
          {selectedProduct?.id === row.id ? '已选择' : '选择'}
        </Button>
      ),
    },
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
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 180,
      render: (_, row) => row.status === 'pending' ? (
        <Space>
          <Button type="link" size="small" icon={<PlusOutlined />} onClick={() => openCreateModal(row)}>采纳发起</Button>
          <Popconfirm title="驳回该集采申请？" okText="确认驳回" cancelText="取消" onConfirm={() => rejectRequestMutation.mutate({ id: row.id })}>
            <Button type="link" danger size="small" loading={rejectRequestMutation.isPending}>驳回</Button>
          </Popconfirm>
        </Space>
      ) : row.activityId ? (
        <Button type="link" size="small" onClick={() => setSelectedId(row.activityId)}>查看活动</Button>
      ) : '-',
    },
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

      <Card title="集采管理" extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => openCreateModal()}>发起集采</Button>}>
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
              <Col span={6}><Card size="small"><Statistic title="参团企业" value={detail.participatingCount} suffix="家" /></Card></Col>
              <Col span={6}><Card size="small"><Statistic title="归集需求" value={detail.demandQuantity} suffix={detail.unit || '吨'} precision={2} /></Card></Col>
              <Col span={6}><Card size="small"><Statistic title="成团门槛" value={detail.minGroupQuantity} suffix={detail.unit || '吨'} precision={2} /></Card></Col>
              <Col span={6}><Card size="small"><Statistic title="已成单子订单" value={(detail.participants || []).filter((item) => item.childOrderId > 0).length} suffix="笔" /></Card></Col>
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
              <Descriptions.Item label="母订单状态">{detail.masterOrderId ? orderStatusTag(detail.masterOrderStatus) : '-'}</Descriptions.Item>
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
        onCancel={closeCreateModal}
        okText="发送给卖家"
        width={1080}
        confirmLoading={createMutation.isPending}
        onOk={() => {
          createForm.validateFields().then((values) => {
            if (!values.productId) {
              message.warning('请先从商品列表选择商品')
              return
            }
            createMutation.mutate({
              requestId: values.requestId,
              productId: values.productId,
              title: values.title,
              targetQuantity: values.targetQuantity,
              durationHours: values.durationHours,
              adminRemark: values.adminRemark,
            })
          })
        }}
      >
        <Form form={createForm} layout="vertical" initialValues={{ durationHours: 24 }}>
          <Form.Item name="requestId" hidden><InputNumber /></Form.Item>
          <Form.Item name="productId" hidden rules={[{ required: true, message: '请从商品列表选择商品' }]}><InputNumber /></Form.Item>
          {sourceRequest ? (
            <Card size="small" className="mb-4" title="来源集采申请">
              <Descriptions column={3} size="small">
                <Descriptions.Item label="申请编号">{sourceRequest.requestNo}</Descriptions.Item>
                <Descriptions.Item label="申请企业">{sourceRequest.buyerName}</Descriptions.Item>
                <Descriptions.Item label="来源">{sourceRequest.source}</Descriptions.Item>
                <Descriptions.Item label="品类">{sourceRequest.category}</Descriptions.Item>
                <Descriptions.Item label="采购物品">{sourceRequest.productName}</Descriptions.Item>
                <Descriptions.Item label="需求量">{quantity(sourceRequest.quantity, sourceRequest.unit)}</Descriptions.Item>
                <Descriptions.Item label="收货地址" span={3}>{sourceRequest.deliveryAddress || '-'}</Descriptions.Item>
              </Descriptions>
            </Card>
          ) : null}

          <Card
            size="small"
            title="选择卖家商品"
            extra={
              <Input.Search
                allowClear
                placeholder="搜索商品名、品类、钢厂、编号"
                style={{ width: 280 }}
                value={productKeyword}
                onChange={(event) => setProductKeyword(event.target.value)}
                onSearch={(keyword) => setProductParams({ ...productParams, page: 1, keyword: keyword.trim() || undefined })}
              />
            }
            className="mb-4"
          >
            <Table
              rowKey="id"
              size="small"
              loading={productLoading}
              columns={productColumns}
              dataSource={products}
              scroll={{ x: 1200 }}
              pagination={{
                current: productParams.page,
                pageSize: productParams.pageSize,
                total: productData?.data.data.total || 0,
                onChange: (page, pageSize) => setProductParams({ ...productParams, page, pageSize }),
              }}
              rowClassName={(record) => record.id === selectedProduct?.id ? 'ant-table-row-selected' : ''}
              onRow={(record) => ({ onClick: () => selectProduct(record) })}
            />
          </Card>

          {selectedProduct ? (
            <Descriptions bordered size="small" column={3} className="mb-4">
              <Descriptions.Item label="已选商品">{selectedProduct.name}</Descriptions.Item>
              <Descriptions.Item label="商品编号">{selectedProduct.productNo || '-'}</Descriptions.Item>
              <Descriptions.Item label="卖家企业">#{selectedProduct.enterpriseId}</Descriptions.Item>
              <Descriptions.Item label="集采品类">{selectedProduct.subCategory || selectedProduct.category || '-'}</Descriptions.Item>
              <Descriptions.Item label="当前价格">{money(selectedProduct.price)}</Descriptions.Item>
              <Descriptions.Item label="可用库存">{quantity(availableStock(selectedProduct), selectedProduct.unit)}</Descriptions.Item>
            </Descriptions>
          ) : null}

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="title" label="集采标题">
                <Input placeholder="不填则按商品名称生成" />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="targetQuantity" label="目标数量（吨）">
                <InputNumber min={0.01} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="durationHours" label="默认时限（小时）" rules={[{ required: true }]}>
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={24}>
              <Form.Item name="adminRemark" label="管理备注">
                <Input.TextArea rows={3} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </Space>
  )
}
