import { useMemo, useState, type Key } from 'react'
import { Button, Card, Col, Descriptions, Drawer, Form, Input, InputNumber, Modal, Popconfirm, Row, Space, Statistic, Table, Tag, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { CheckCircleOutlined, EyeOutlined, FileDoneOutlined, PlusOutlined, ReloadOutlined, SearchOutlined, ShopOutlined, TeamOutlined } from '@ant-design/icons'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { formatDateTime } from '../../utils/format'
import {
  GroupProcurementPageParams,
  GroupProcurementProjectItem,
  GroupProcurementProjectQuoteItem,
  GroupProcurementProjectSellerItem,
  GroupProcurementRequestItem,
  PageParams,
  ProductItem,
  groupProcurementProjectApi,
  groupProcurementRequestApi,
  productApi,
} from '../../services/api'

const { Text } = Typography

function getErrorMessage(error: unknown, fallback: string) {
  const maybe = error as { response?: { data?: { message?: string } }; message?: string }
  return maybe.response?.data?.message || maybe.message || fallback
}

function money(value?: number) {
  return `¥${Number(value || 0).toLocaleString()}`
}

function quantity(value?: number, unit?: string) {
  return `${Number(value || 0).toLocaleString()} ${unit || '吨'}`
}

function projectStatusTag(status?: string) {
  const map: Record<string, { color: string; text: string }> = {
    collecting: { color: 'gold', text: '需求归集' },
    quoting: { color: 'blue', text: '报价中' },
    awarded: { color: 'green', text: '已定标' },
    order_created: { color: 'purple', text: '已成单' },
    closed: { color: 'default', text: '已关闭' },
  }
  const meta = map[status || ''] || { color: 'default', text: status || '-' }
  return <Tag color={meta.color}>{meta.text}</Tag>
}

function enrollmentStatusTag(status?: string) {
  const map: Record<string, { color: string; text: string }> = {
    invited: { color: 'gold', text: '已邀请' },
    approved: { color: 'green', text: '已参与' },
    rejected: { color: 'red', text: '已拒绝' },
  }
  const meta = map[status || ''] || { color: 'default', text: status || '-' }
  return <Tag color={meta.color}>{meta.text}</Tag>
}

interface FilterValues {
  keyword?: string
  status?: string
  category?: string
}

export default function GroupProcurementPage() {
  const queryClient = useQueryClient()
  const [filterForm] = Form.useForm<FilterValues>()
  const [createForm] = Form.useForm<{ requestIds?: number[]; title?: string; targetQuantity?: number; durationHours?: number; adminRemark?: string }>()
  const [params, setParams] = useState<GroupProcurementPageParams>({ page: 1, pageSize: 10 })
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [inviteOpen, setInviteOpen] = useState(false)
  const [selectedRequestRowKeys, setSelectedRequestRowKeys] = useState<Key[]>([])
  const [selectedRequestRows, setSelectedRequestRows] = useState<GroupProcurementRequestItem[]>([])
  const [selectedProductRowKeys, setSelectedProductRowKeys] = useState<Key[]>([])
  const [productParams, setProductParams] = useState<PageParams>({ page: 1, pageSize: 8, status: 2 })

  const { data, isLoading } = useQuery({
    queryKey: ['group-procurement-projects', params],
    queryFn: () => groupProcurementProjectApi.list(params),
  })
  const { data: detailData, isFetching: detailLoading } = useQuery({
    queryKey: ['group-procurement-project-detail', selectedId],
    queryFn: () => groupProcurementProjectApi.get(selectedId!),
    enabled: selectedId !== null,
  })
  const { data: requestData, isLoading: requestLoading } = useQuery({
    queryKey: ['group-procurement-requests', 'pending'],
    queryFn: () => groupProcurementRequestApi.listRequests({ page: 1, pageSize: 20, status: 'pending' }),
  })
  const { data: productData, isLoading: productLoading } = useQuery({
    queryKey: ['group-procurement-invite-products', productParams],
    queryFn: () => productApi.list(productParams),
    enabled: inviteOpen,
  })

  const createMutation = useMutation({
    mutationFn: groupProcurementProjectApi.create,
    onSuccess: () => {
      message.success('竞价项目已创建')
      setCreateOpen(false)
      setSelectedRequestRowKeys([])
      setSelectedRequestRows([])
      createForm.resetFields()
      void queryClient.invalidateQueries({ queryKey: ['group-procurement-projects'] })
      void queryClient.invalidateQueries({ queryKey: ['group-procurement-requests'] })
    },
    onError: (error) => message.error(getErrorMessage(error, '创建项目失败')),
  })

  const rejectRequestMutation = useMutation({
    mutationFn: ({ id }: { id: number }) => groupProcurementRequestApi.rejectRequest(id),
    onSuccess: () => {
      message.success('集采申请已驳回')
      void queryClient.invalidateQueries({ queryKey: ['group-procurement-requests'] })
    },
    onError: (error) => message.error(getErrorMessage(error, '驳回申请失败')),
  })

  const inviteMutation = useMutation({
    mutationFn: ({ id, productIds }: { id: number; productIds: number[] }) => groupProcurementProjectApi.inviteSellers(id, { productIds }),
    onSuccess: () => {
      message.success('卖家邀请已发送')
      setInviteOpen(false)
      setSelectedProductRowKeys([])
      void queryClient.invalidateQueries({ queryKey: ['group-procurement-projects'] })
      if (selectedId) void queryClient.invalidateQueries({ queryKey: ['group-procurement-project-detail', selectedId] })
    },
    onError: (error) => message.error(getErrorMessage(error, '邀请卖家失败')),
  })

  const awardMutation = useMutation({
    mutationFn: ({ id, quoteId }: { id: number; quoteId: number }) => groupProcurementProjectApi.award(id, { quoteId }),
    onSuccess: () => {
      message.success('项目已定标')
      void queryClient.invalidateQueries({ queryKey: ['group-procurement-projects'] })
      if (selectedId) void queryClient.invalidateQueries({ queryKey: ['group-procurement-project-detail', selectedId] })
    },
    onError: (error) => message.error(getErrorMessage(error, '项目定标失败')),
  })

  const createOrderMutation = useMutation({
    mutationFn: ({ id }: { id: number }) => groupProcurementProjectApi.createOrder(id),
    onSuccess: () => {
      message.success('已基于中标报价生成订单')
      void queryClient.invalidateQueries({ queryKey: ['group-procurement-projects'] })
      if (selectedId) void queryClient.invalidateQueries({ queryKey: ['group-procurement-project-detail', selectedId] })
    },
    onError: (error) => message.error(getErrorMessage(error, '生成订单失败')),
  })

  const rows = data?.data.data.list ?? []
  const detail = detailData?.data.data
  const requests = requestData?.data.data.list ?? []
  const products = productData?.data.data.list ?? []

  const stats = useMemo(() => ({
    collecting: rows.filter((item) => item.status === 'collecting').length,
    quoting: rows.filter((item) => item.status === 'quoting').length,
    awarded: rows.filter((item) => item.status === 'awarded').length,
    demand: rows.reduce((sum, item) => sum + Number(item.demandQuantity || 0), 0),
  }), [rows])

  const projectColumns: ColumnsType<GroupProcurementProjectItem> = [
    { title: '项目编号', dataIndex: 'projectNo', width: 180 },
    {
      title: '项目',
      key: 'project',
      width: 260,
      render: (_, row) => (
        <div>
          <Text strong>{row.title}</Text>
          <div><Text type="secondary">{row.productName}</Text></div>
        </div>
      ),
    },
    { title: '状态', dataIndex: 'status', width: 120, render: (_, row) => projectStatusTag(row.status) },
    { title: '需求', key: 'demand', width: 160, render: (_, row) => quantity(row.demandQuantity, row.unit) },
    { title: '买家', key: 'buyers', width: 120, render: (_, row) => `${row.buyerEnterpriseCount} 家 / ${row.buyerRequestCount} 单` },
    { title: '卖家', key: 'sellers', width: 140, render: (_, row) => `${row.invitedSupplierCount} 邀请 / ${row.quotedSupplierCount} 报价` },
    { title: '目标价', dataIndex: 'targetPrice', width: 120, render: (value) => money(value) },
    { title: '截止时间', dataIndex: 'quoteDeadline', width: 180, render: (value) => formatDateTime(value) },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, row) => <Button type="link" icon={<EyeOutlined />} onClick={() => setSelectedId(row.id)}>详情</Button>,
    },
  ]

  const requestColumns: ColumnsType<GroupProcurementRequestItem> = [
    { title: '申请编号', dataIndex: 'requestNo', width: 170 },
    { title: '申请企业', dataIndex: 'buyerName', width: 180 },
    { title: '品类', dataIndex: 'category', width: 120 },
    { title: '采购物品', dataIndex: 'productName', width: 180 },
    { title: '数量', key: 'quantity', width: 120, render: (_, row) => quantity(row.quantity, row.unit) },
    { title: '目标价', dataIndex: 'targetPrice', width: 120, render: (value) => money(value) },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, row) => (
        <Popconfirm title="驳回该集采申请？" okText="确认" cancelText="取消" onConfirm={() => rejectRequestMutation.mutate({ id: row.id })}>
          <Button type="link" danger>驳回</Button>
        </Popconfirm>
      ),
    },
  ]

  const productColumns: ColumnsType<ProductItem> = [
    { title: '商品', key: 'product', width: 220, render: (_, row) => <div><Text strong>{row.name}</Text><div><Text type="secondary">#{row.id} {row.productNo || '-'}</Text></div></div> },
    { title: '卖家企业', key: 'seller', width: 140, render: (_, row) => row.enterpriseId ? `#${row.enterpriseId}` : '-' },
    { title: '品类', key: 'category', width: 140, render: (_, row) => row.subCategory || row.category || '-' },
    { title: '价格', dataIndex: 'price', width: 100, render: (value) => money(value) },
    { title: '库存', key: 'stock', width: 100, render: (_, row) => quantity(row.availableQty || row.stock || 0, '吨') },
  ]

  const quoteColumns: ColumnsType<GroupProcurementProjectQuoteItem> = [
    { title: '卖家', dataIndex: 'sellerName', width: 180 },
    { title: '商品', key: 'product', width: 220, render: (_, row) => `${row.productName || '-'} ${row.productNo ? `(#${row.productNo})` : ''}` },
    { title: '报价', dataIndex: 'unitPrice', width: 120, render: (value) => money(value) },
    { title: '可供量', key: 'supply', width: 120, render: (_, row) => quantity(row.supplyQuantity, row.unit) },
    { title: '交期', key: 'deliveryDays', width: 100, render: (_, row) => `${row.deliveryDays} 天` },
    { title: '仓库', dataIndex: 'deliveryWarehouse', width: 160, render: (value) => value || '-' },
    { title: '付款条件', dataIndex: 'paymentTerms', width: 160, render: (value) => value || '-' },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, row) => detail?.status !== 'order_created' ? (
        <Popconfirm title="确定将该报价定为中标？" okText="定标" cancelText="取消" onConfirm={() => selectedId && awardMutation.mutate({ id: selectedId, quoteId: row.id })}>
          <Button type="link" disabled={detail?.status === 'order_created'}>定标</Button>
        </Popconfirm>
      ) : '-',
    },
  ]

  const enrollmentColumns: ColumnsType<GroupProcurementProjectSellerItem> = [
    { title: '卖家企业', dataIndex: 'sellerName', width: 180 },
    { title: '推荐商品', key: 'product', width: 220, render: (_, row) => `${row.productName || '-'} ${row.productNo ? `(#${row.productNo})` : ''}` },
    { title: '状态', key: 'status', width: 120, render: (_, row) => enrollmentStatusTag(row.status) },
    { title: '备注', dataIndex: 'remark', render: (value) => value || '-' },
  ]

  const applyFilters = (values: FilterValues) => {
    setParams({ ...params, page: 1, keyword: values.keyword, status: values.status, category: values.category })
  }

  const resetFilters = () => {
    filterForm.resetFields()
    setParams({ page: 1, pageSize: 10 })
  }

  const openCreateModal = () => {
    const totalQuantity = selectedRequestRows.reduce((sum, item) => sum + Number(item.quantity || 0), 0)
    createForm.setFieldsValue({
      requestIds: selectedRequestRows.map((item) => item.id),
      title: selectedRequestRows[0] ? `集采竞价项目：${selectedRequestRows[0].productName}` : undefined,
      targetQuantity: totalQuantity || undefined,
      durationHours: 48,
      adminRemark: selectedRequestRows.map((item) => item.remark).filter(Boolean).join('\n'),
    })
    setCreateOpen(true)
  }

  const selectedProducts = products.filter((item) => selectedProductRowKeys.includes(item.id))

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Row gutter={16}>
        <Col xs={24} md={6}><Card><Statistic title="需求归集中" value={stats.collecting} suffix="个" prefix={<ReloadOutlined />} /></Card></Col>
        <Col xs={24} md={6}><Card><Statistic title="报价中" value={stats.quoting} suffix="个" prefix={<TeamOutlined />} /></Card></Col>
        <Col xs={24} md={6}><Card><Statistic title="已定标" value={stats.awarded} suffix="个" prefix={<CheckCircleOutlined />} /></Card></Col>
        <Col xs={24} md={6}><Card><Statistic title="当前页归集量" value={stats.demand} suffix="吨" prefix={<ShopOutlined />} precision={2} /></Card></Col>
      </Row>

      <Card title="竞价项目">
        <Form form={filterForm} layout="vertical" onFinish={applyFilters} style={{ marginBottom: 16 }}>
          <Row gutter={12}>
            <Col xs={24} md={8}><Form.Item name="keyword" label="搜索"><Input allowClear prefix={<SearchOutlined />} placeholder="项目编号 / 标题 / 采购物品" /></Form.Item></Col>
            <Col xs={12} md={5}><Form.Item name="status" label="状态"><Input allowClear placeholder="collecting / quoting / awarded" /></Form.Item></Col>
            <Col xs={12} md={5}><Form.Item name="category" label="品类"><Input allowClear placeholder="如 热轧板卷" /></Form.Item></Col>
            <Col xs={24} md={6}>
              <Form.Item label="操作">
                <Space>
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
          columns={projectColumns}
          dataSource={rows}
          scroll={{ x: 1280 }}
          pagination={{
            current: params.page,
            pageSize: params.pageSize,
            total: data?.data.data.total || 0,
            onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
          }}
        />
      </Card>

      <Card
        title="待归并采购申请"
        extra={<Button type="primary" icon={<PlusOutlined />} disabled={selectedRequestRows.length === 0} onClick={openCreateModal}>创建竞价项目</Button>}
      >
        <Table
          rowKey="id"
          loading={requestLoading}
          columns={requestColumns}
          dataSource={requests}
          scroll={{ x: 1100 }}
          rowSelection={{
            selectedRowKeys: selectedRequestRowKeys,
            onChange: (keys, rows) => {
              setSelectedRequestRowKeys(keys)
              setSelectedRequestRows(rows)
            },
          }}
          pagination={false}
        />
      </Card>

      <Modal
        title="创建竞价项目"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        confirmLoading={createMutation.isPending}
        onOk={() => {
          createForm.validateFields().then((values) => {
            createMutation.mutate({
              requestIds: values.requestIds || selectedRequestRows.map((item) => item.id),
              title: values.title,
              targetQuantity: values.targetQuantity,
              durationHours: values.durationHours,
              adminRemark: values.adminRemark,
            })
          }).catch(() => undefined)
        }}
      >
        <Form form={createForm} layout="vertical">
          <Form.Item name="requestIds" hidden><Input /></Form.Item>
          <Form.Item name="title" label="项目标题" rules={[{ required: true, message: '请输入项目标题' }]}><Input /></Form.Item>
          <Form.Item name="targetQuantity" label="目标数量（吨）"><InputNumber min={0.01} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="durationHours" label="报价时限（小时）" rules={[{ required: true, message: '请输入报价时限' }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="adminRemark" label="管理备注"><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>

      <Modal
        title="邀请卖家参与报价"
        open={inviteOpen}
        onCancel={() => setInviteOpen(false)}
        width={1000}
        confirmLoading={inviteMutation.isPending}
        onOk={() => {
          if (!selectedId || selectedProducts.length === 0) {
            message.warning('请先选择至少一个卖家商品')
            return
          }
          inviteMutation.mutate({ id: selectedId, productIds: selectedProducts.map((item) => item.id) })
        }}
      >
        <Table
          rowKey="id"
          loading={productLoading}
          columns={productColumns}
          dataSource={products}
          scroll={{ x: 900 }}
          rowSelection={{
            selectedRowKeys: selectedProductRowKeys,
            onChange: setSelectedProductRowKeys,
          }}
          pagination={{
            current: productParams.page,
            pageSize: productParams.pageSize,
            total: productData?.data.data.total || 0,
            onChange: (page, pageSize) => setProductParams({ ...productParams, page, pageSize }),
          }}
        />
      </Modal>

      <Drawer
        title={detail ? `竞价项目详情 ${detail.projectNo}` : '竞价项目详情'}
        open={selectedId !== null}
        onClose={() => setSelectedId(null)}
        width={1100}
      >
        {detail ? (
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            <Row gutter={12}>
              <Col span={6}><Card size="small"><Statistic title="买家申请" value={detail.buyerRequestCount} suffix="单" /></Card></Col>
              <Col span={6}><Card size="small"><Statistic title="参与买家" value={detail.buyerEnterpriseCount} suffix="家" /></Card></Col>
              <Col span={6}><Card size="small"><Statistic title="已邀请卖家" value={detail.invitedSupplierCount} suffix="家" /></Card></Col>
              <Col span={6}><Card size="small"><Statistic title="有效报价" value={detail.quotedSupplierCount} suffix="条" /></Card></Col>
            </Row>

            <Descriptions bordered column={2} size="small">
              <Descriptions.Item label="状态">{projectStatusTag(detail.status)}</Descriptions.Item>
              <Descriptions.Item label="报价截止">{formatDateTime(detail.quoteDeadline)}</Descriptions.Item>
              <Descriptions.Item label="品类">{detail.subCategory || detail.category || '-'}</Descriptions.Item>
              <Descriptions.Item label="采购物品">{detail.productName}</Descriptions.Item>
              <Descriptions.Item label="归集需求">{quantity(detail.demandQuantity, detail.unit)}</Descriptions.Item>
              <Descriptions.Item label="目标价">{money(detail.targetPrice)}</Descriptions.Item>
              <Descriptions.Item label="中标卖家">{detail.awardSupplierName || '-'}</Descriptions.Item>
              <Descriptions.Item label="中标报价">{detail.awardUnitPrice ? money(detail.awardUnitPrice) : '-'}</Descriptions.Item>
              <Descriptions.Item label="订单号">{detail.orderNo || '-'}</Descriptions.Item>
              <Descriptions.Item label="收货地址">{detail.deliveryAddress || '-'}</Descriptions.Item>
              <Descriptions.Item label="管理备注" span={2}>{detail.adminRemark || '-'}</Descriptions.Item>
            </Descriptions>

            <Space wrap>
              <Button type="primary" icon={<PlusOutlined />} onClick={() => {
                const searchKeyword = [detail.subCategory, detail.category, detail.productName].filter(Boolean).join(' ')
                setProductParams({
                  page: 1,
                  pageSize: 8,
                  status: 2,
                  keyword: searchKeyword || undefined,
                  category: detail.category || undefined,
                })
                setSelectedProductRowKeys([])
                setInviteOpen(true)
              }}>
                邀请卖家
              </Button>
              {detail.status === 'awarded' && !detail.orderId ? (
                <Popconfirm title="确认基于中标报价生成订单？" okText="生成" cancelText="取消" onConfirm={() => createOrderMutation.mutate({ id: detail.id })}>
                  <Button icon={<FileDoneOutlined />} loading={createOrderMutation.isPending}>生成订单</Button>
                </Popconfirm>
              ) : null}
            </Space>

            <Card title="归集申请">
              <Table rowKey="id" columns={requestColumns.filter((item) => item.key !== 'action')} dataSource={detail.requests || []} pagination={false} scroll={{ x: 900 }} />
            </Card>

            <Card title="参与卖家">
              <Table rowKey="id" columns={enrollmentColumns} dataSource={detail.enrollments || []} pagination={false} scroll={{ x: 780 }} />
            </Card>

            <Card title="报价对比">
              <Table rowKey="id" loading={detailLoading} columns={quoteColumns} dataSource={detail.quotes || []} pagination={false} scroll={{ x: 1100 }} />
            </Card>
          </Space>
        ) : null}
      </Drawer>
    </Space>
  )
}
