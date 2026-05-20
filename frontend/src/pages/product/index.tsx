import { useState } from 'react'
import {
  Alert,
  Button,
  Card,
  Col,
  Descriptions,
  Drawer,
  Form,
  Image,
  Input,
  InputNumber,
  Modal,
  Row,
  Segmented,
  Select,
  Skeleton,
  Space,
  Table,
  Tag,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { productApi, ProductAuditLogItem, ProductAuditSnapshot, ProductItem, PageParams } from '../../services/api'
import { formatDateTime } from '../../utils/format'
import { canAudit } from '../../utils/permissions'
import { useAuthStore } from '../../store/useAuthStore'

const statusMeta = {
  0: { color: 'default', text: '草稿' },
  1: { color: 'orange', text: '待审核' },
  2: { color: 'green', text: '已上架' },
  3: { color: 'default', text: '已下架' },
  4: { color: 'red', text: '审核拒绝' },
} as const

const marketplaceVisibilityMeta = {
  0: { color: 'default', text: '草稿，商城不可见' },
  1: { color: 'gold', text: '待审核，商城暂不可见' },
  2: { color: 'green', text: '已上架，商城可见' },
  3: { color: 'default', text: '已下架，商城不可见' },
  4: { color: 'red', text: '审核拒绝，商城不可见' },
} as const

const goodsTypeMeta = {
  spot: { color: 'blue', text: '现货' },
  futures: { color: 'purple', text: '期货' },
} as const

interface ProductFilterValues {
  keyword?: string
  status?: number
  enterpriseId?: number
  category?: string
  subCategory?: string
  goodsType?: string
  warehouseServiceId?: number
}

interface CompareRow {
  key: string
  group: string
  label: string
  submitted: string
  baseline: string
  changed: boolean
}

const pendingReviewStatuses = new Set([0, 1, 4])

function textValue(value: unknown) {
  if (value === null || value === undefined || value === '') return '-'
  return String(value)
}

function moneyValue(value: unknown) {
  const amount = Number(value || 0)
  return amount > 0 ? `¥${amount.toFixed(2)}` : '-'
}

function numberValue(value: unknown, suffix = '') {
  const amount = Number(value || 0)
  return amount > 0 ? `${amount}${suffix}` : '-'
}

function snapshotText(snapshot: ProductAuditSnapshot | undefined, key: string) {
  if (!snapshot) return '暂无历史快照'
  if (key === 'goodsType') {
    const value = String(snapshot.goodsType || 'spot') as keyof typeof goodsTypeMeta
    return goodsTypeMeta[value]?.text || textValue(snapshot.goodsType)
  }
  if (key === 'price') return moneyValue(snapshot.price)
  if (key === 'stock') return numberValue(snapshot.stock)
  if (key === 'minOrderQty') return numberValue(snapshot.minOrderQty)
  if (key === 'deliveryDays') return numberValue(snapshot.deliveryDays, ' 天')
  if (key === 'quoteValidity') return numberValue(snapshot.quoteValidity, ' 小时')
  if (key === 'warehouseServiceId') return snapshot.warehouseServiceId ? `#${snapshot.warehouseServiceId}` : '-'
  if (key === 'warehouseRegion') {
    return textValue(snapshot.warehouseRegion || [snapshot.province, snapshot.city, snapshot.district].filter(Boolean).join(' / '))
  }
  return textValue(snapshot[key])
}

function latestAuditBaseline(product?: ProductItem): ProductAuditSnapshot | undefined {
  const latest = product?.auditHistory?.[0]
  return latest?.afterSnapshot || latest?.beforeSnapshot
}

function productCompareRows(product?: ProductItem, baselineSnapshot?: ProductAuditSnapshot): CompareRow[] {
  if (!product) return []
  const rows: Omit<CompareRow, 'baseline' | 'changed'>[] = [
    { key: 'name', group: '基础信息', label: '商品名称', submitted: textValue(product.name) },
    { key: 'productNo', group: '基础信息', label: '商品编号', submitted: textValue(product.productNo) },
    { key: 'goodsType', group: '基础信息', label: '货品类型', submitted: goodsTypeMeta[(product.goodsType || 'spot') as keyof typeof goodsTypeMeta]?.text || textValue(product.goodsType) },
    { key: 'category', group: '基础信息', label: '主品类', submitted: textValue(product.category) },
    { key: 'subCategory', group: '基础信息', label: '子品类', submitted: textValue(product.subCategory) },
    { key: 'steelMill', group: '基础信息', label: '钢厂', submitted: textValue(product.steelMill) },
    { key: 'spec', group: '规格价格', label: '规格', submitted: textValue(product.spec) },
    { key: 'material', group: '规格价格', label: '材质', submitted: textValue(product.material) },
    { key: 'bundleNo', group: '规格价格', label: '捆包号', submitted: textValue(product.bundleNo) },
    { key: 'price', group: '规格价格', label: '单价', submitted: moneyValue(product.price) },
    { key: 'stock', group: '规格价格', label: '库存', submitted: numberValue(product.stock) },
    { key: 'minOrderQty', group: '规格价格', label: '最小起订量', submitted: numberValue(product.minOrderQty) },
    { key: 'deliveryDays', group: '履约信息', label: '交货周期', submitted: numberValue(product.deliveryDays, ' 天') },
    { key: 'quoteValidity', group: '履约信息', label: '报价有效期', submitted: numberValue(product.quoteValidity, ' 小时') },
    { key: 'warehouseServiceId', group: '仓储信息', label: '仓储资源', submitted: product.warehouseServiceId ? `#${product.warehouseServiceId}` : '-' },
    { key: 'warehouseName', group: '仓储信息', label: '仓库名称', submitted: textValue(product.warehouseName) },
    { key: 'warehouseCode', group: '仓储信息', label: '仓库编码', submitted: textValue(product.warehouseCode) },
    { key: 'warehouseRegion', group: '仓储信息', label: '仓库地区', submitted: [product.province, product.city, product.district].filter(Boolean).join(' / ') || '-' },
    { key: 'address', group: '仓储信息', label: '仓库地址', submitted: textValue(product.address) },
    { key: 'description', group: '其他', label: '商品描述', submitted: textValue(product.description) },
  ]
  return rows.map((row) => ({
    ...row,
    baseline: snapshotText(baselineSnapshot, row.key),
    changed: baselineSnapshot ? snapshotText(baselineSnapshot, row.key) !== row.submitted : pendingReviewStatuses.has(product.status),
  }))
}

function auditStatusText(status: string) {
  if (status === 'approved') return '通过'
  if (status === 'rejected') return '驳回'
  return status || '-'
}

function auditStatusColor(status: string) {
  if (status === 'approved') return 'green'
  if (status === 'rejected') return 'red'
  return 'default'
}

export default function ProductList() {
  const [filterForm] = Form.useForm<ProductFilterValues>()
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [auditTarget, setAuditTarget] = useState<ProductItem | null>(null)
  const [auditForm] = Form.useForm<{ status: number; auditRemark?: string }>()
  const queryClient = useQueryClient()
  const roleId = useAuthStore((state) => state.adminInfo?.roleId)
  const auditAllowed = canAudit(roleId)

  const { data, isLoading } = useQuery({
    queryKey: ['products', params],
    queryFn: () => productApi.list(params),
  })

  const { data: detailData, isFetching: detailLoading } = useQuery({
    queryKey: ['product-detail', selectedId],
    queryFn: () => productApi.get(selectedId!),
    enabled: selectedId !== null,
  })

  const auditTargetId = auditTarget?.id
  const { data: auditDetailData, isFetching: auditDetailLoading } = useQuery({
    queryKey: ['product-detail', auditTargetId],
    queryFn: () => productApi.get(auditTargetId!),
    enabled: auditTargetId !== undefined,
  })

  const rows = data?.data.data.list ?? []
  const detail = detailData?.data.data
  const auditProductDetail = auditDetailData?.data.data ?? auditTarget ?? undefined
  const detailBaseline = latestAuditBaseline(detail)
  const auditBaseline = latestAuditBaseline(auditProductDetail)
  const detailCompareRows = productCompareRows(detail, detailBaseline)
  const auditCompareRows = productCompareRows(auditProductDetail, auditBaseline)

  const { mutate: auditProduct, isPending: auditSubmitting } = useMutation({
    mutationFn: ({ id, values }: { id: number; values: { status: number; auditRemark?: string } }) =>
      productApi.audit(id, values),
    onSuccess: (_, variables) => {
      message.success(
        variables.values.status === 1
          ? '商品审核已通过，当前已进入商城展示范围'
          : '商品已驳回，当前不会进入商城展示',
      )
      setAuditTarget(null)
      auditForm.resetFields()
      void queryClient.invalidateQueries({ queryKey: ['products'] })
      void queryClient.invalidateQueries({ queryKey: ['product-detail', variables.id] })
      if (selectedId !== null) {
        void queryClient.invalidateQueries({ queryKey: ['product-detail', selectedId] })
      }
    },
  })

  const columns: ColumnsType<ProductItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '商品名称', dataIndex: 'name', key: 'name', width: 240 },
    {
      title: '货品类型',
      dataIndex: 'goodsType',
      key: 'goodsType',
      width: 100,
      render: (value?: 'spot' | 'futures') => {
        const meta = goodsTypeMeta[value || 'spot']
        return <Tag color={meta.color}>{meta.text}</Tag>
      },
    },
    { title: '规格', dataIndex: 'spec', key: 'spec', width: 180 },
    { title: '单位', dataIndex: 'unit', key: 'unit', width: 90 },
    {
      title: '价格',
      dataIndex: 'price',
      key: 'price',
      width: 120,
      render: (price: number) => `¥${price.toFixed(2)}`,
    },
    { title: '库存', dataIndex: 'stock', key: 'stock', width: 100 },
    {
      title: '审核状态',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (status: number) => {
        const meta = statusMeta[status as keyof typeof statusMeta] ?? statusMeta[1]
        return <Tag color={meta.color}>{meta.text}</Tag>
      },
    },
    {
      title: '审核备注',
      dataIndex: 'auditRemark',
      key: 'auditRemark',
      width: 220,
      render: (value?: string) => value || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (value: string) => formatDateTime(value),
    },
    {
      title: '操作',
      key: 'actions',
      width: 160,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => setSelectedId(record.id)}>
            查看
          </Button>
          {auditAllowed && (
            <Button
              type="link"
              size="small"
              onClick={() => {
                setAuditTarget(record)
                auditForm.setFieldsValue({
                  status: record.status === 4 ? 2 : 1,
                  auditRemark: record.auditRemark,
                })
              }}
            >
              审核
            </Button>
          )}
        </Space>
      ),
    },
  ]

  const compareColumns: ColumnsType<CompareRow> = [
    { title: '分组', dataIndex: 'group', key: 'group', width: 96 },
    { title: '字段', dataIndex: 'label', key: 'label', width: 120 },
    { title: '当前提交值', dataIndex: 'submitted', key: 'submitted', ellipsis: true },
    {
      title: '审核基准',
      dataIndex: 'baseline',
      key: 'baseline',
      width: 130,
      render: (value: string) => <Tag>{value}</Tag>,
    },
    {
      title: '判断',
      key: 'changed',
      width: 100,
      render: (_: unknown, record: CompareRow) => record.changed ? <Tag color="gold">待核对</Tag> : <Tag color="green">已审核</Tag>,
    },
  ]

  const auditHistoryColumns: ColumnsType<ProductAuditLogItem> = [
    {
      title: '审核时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 170,
      render: (value: string) => formatDateTime(value),
    },
    {
      title: '结果',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (value: string) => <Tag color={auditStatusColor(value)}>{auditStatusText(value)}</Tag>,
    },
    {
      title: '状态变化',
      key: 'statusChange',
      width: 150,
      render: (_, record) => {
        const from = statusMeta[record.fromStatus as keyof typeof statusMeta]?.text ?? `状态${record.fromStatus}`
        const to = statusMeta[record.toStatus as keyof typeof statusMeta]?.text ?? `状态${record.toStatus}`
        return record.fromStatus || record.toStatus ? `${from} -> ${to}` : '-'
      },
    },
    {
      title: '审核人',
      key: 'operator',
      width: 120,
      render: (_, record) => record.operatorName || (record.operatorId ? `#${record.operatorId}` : '-'),
    },
    { title: '备注', dataIndex: 'remark', key: 'remark', ellipsis: true, render: (value?: string) => value || '-' },
  ]

  const applyFilters = (values: ProductFilterValues) => {
    setParams({
      page: 1,
      pageSize: params.pageSize,
      ...values,
    })
  }

  const resetFilters = () => {
    filterForm.resetFields()
    setParams({ page: 1, pageSize: params.pageSize })
  }

  return (
    <>
      <Card title="商品管理">
        <Form form={filterForm} layout="vertical" onFinish={applyFilters} initialValues={params} className="mb-4">
          <Row gutter={12}>
            <Col xs={24} md={8} lg={6}>
              <Form.Item name="keyword" label="搜索">
                <Input allowClear placeholder="商品名/品类/钢厂/规格/仓库" />
              </Form.Item>
            </Col>
            <Col xs={12} md={8} lg={4}>
              <Form.Item name="status" label="审核状态">
                <Select
                  allowClear
                  placeholder="全部状态"
                  options={[
                    { label: '草稿', value: 0 },
                    { label: '待审核', value: 1 },
                    { label: '已上架', value: 2 },
                    { label: '已下架', value: 3 },
                    { label: '审核拒绝', value: 4 },
                  ]}
                />
              </Form.Item>
            </Col>
            <Col xs={12} md={8} lg={4}>
              <Form.Item name="goodsType" label="货品类型">
                <Select
                  allowClear
                  placeholder="全部类型"
                  options={[
                    { label: '现货', value: 'spot' },
                    { label: '期货', value: 'futures' },
                  ]}
                />
              </Form.Item>
            </Col>
            <Col xs={12} md={8} lg={4}>
              <Form.Item name="enterpriseId" label="企业 ID">
                <InputNumber min={1} precision={0} style={{ width: '100%' }} placeholder="企业 ID" />
              </Form.Item>
            </Col>
            <Col xs={12} md={8} lg={4}>
              <Form.Item name="warehouseServiceId" label="仓储资源 ID">
                <InputNumber min={1} precision={0} style={{ width: '100%' }} placeholder="仓库 ID" />
              </Form.Item>
            </Col>
            <Col xs={12} md={8} lg={4}>
              <Form.Item name="category" label="主品类">
                <Input allowClear placeholder="如 热轧板卷" />
              </Form.Item>
            </Col>
            <Col xs={12} md={8} lg={4}>
              <Form.Item name="subCategory" label="子品类">
                <Input allowClear placeholder="子品类" />
              </Form.Item>
            </Col>
            <Col xs={24} lg={6}>
              <Form.Item label="操作">
                <Space wrap>
                  <Button type="primary" htmlType="submit">查询</Button>
                  <Button onClick={resetFilters}>重置</Button>
                </Space>
              </Form.Item>
            </Col>
          </Row>
        </Form>
        <Table
          columns={columns}
          dataSource={rows}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1100 }}
          pagination={{
            current: params.page,
            pageSize: params.pageSize,
            total: data?.data.data.total || 0,
            onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
          }}
        />
      </Card>

      <Drawer
        title="商品详情"
        width={640}
        open={selectedId !== null}
        onClose={() => setSelectedId(null)}
        destroyOnHidden
      >
        {detailLoading ? (
          <Skeleton active paragraph={{ rows: 10 }} />
        ) : (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Alert
              type="info"
              showIcon
              message="商品改动核对"
              description={
                detailBaseline
                  ? '下方展示当前提交值与最近一次审核后快照的差异，标记待核对的字段需审核员重点确认。'
                  : '当前商品暂无历史字段快照；首次审核后会自动保存审核前后快照，后续可用于差异核对。'
              }
            />
            <Table
              size="small"
              columns={compareColumns}
              dataSource={detailCompareRows}
              rowKey="key"
              pagination={false}
              scroll={{ x: 760 }}
            />
            {detail?.auditHistory && detail.auditHistory.length > 0 ? (
              <Table
                size="small"
                title={() => '审核历史'}
                columns={auditHistoryColumns}
                dataSource={detail.auditHistory}
                rowKey="id"
                pagination={false}
                scroll={{ x: 760 }}
              />
            ) : (
              <Alert type="info" showIcon message="暂无审核历史快照" />
            )}
            <Descriptions bordered column={1} size="small">
              <Descriptions.Item label="审核状态">
                {detail ? (
                  <Tag color={statusMeta[detail.status as keyof typeof statusMeta]?.color ?? statusMeta[1].color}>
                    {statusMeta[detail.status as keyof typeof statusMeta]?.text ?? statusMeta[1].text}
                  </Tag>
                ) : (
                  '-'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="商城展示状态">
                {detail ? (
                  <Tag color={marketplaceVisibilityMeta[detail.status as keyof typeof marketplaceVisibilityMeta]?.color ?? marketplaceVisibilityMeta[1].color}>
                    {marketplaceVisibilityMeta[detail.status as keyof typeof marketplaceVisibilityMeta]?.text ?? marketplaceVisibilityMeta[1].text}
                  </Tag>
                ) : (
                  '-'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="审核备注">{detail?.auditRemark || '-'}</Descriptions.Item>
              <Descriptions.Item label="商品图片">
                {detail?.images && detail.images.length > 0 ? (
                  <Space wrap>
                    {detail.images.map((image) => (
                      <Image
                        key={image}
                        src={image}
                        width={100}
                        height={100}
                        style={{ objectFit: 'cover' }}
                      />
                    ))}
                  </Space>
                ) : (
                  '-'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="审核时间">{formatDateTime(detail?.auditedAt)}</Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatDateTime(detail?.createdAt)}</Descriptions.Item>
              <Descriptions.Item label="更新时间">{formatDateTime(detail?.updatedAt)}</Descriptions.Item>
            </Descriptions>
          </Space>
        )}
      </Drawer>

      <Modal
        title={`商品审核${auditProductDetail ? ` · ${auditProductDetail.name}` : ''}`}
        open={auditTarget !== null}
        onCancel={() => {
          setAuditTarget(null)
          auditForm.resetFields()
        }}
        onOk={() => {
          auditForm
            .validateFields()
            .then((values) => {
              if (!auditTarget) return
              auditProduct({ id: auditTarget.id, values })
            })
            .catch(() => undefined)
        }}
        confirmLoading={auditSubmitting}
        okText="提交审核"
        cancelText="取消"
        destroyOnHidden
      >
        <Alert
          className="mb-4"
          type="warning"
          showIcon
          message="审核前请核对关键字段"
          description={
            auditBaseline
              ? '下方按商品基础、规格价格、履约仓储等维度对比当前提交值和最近一次审核后快照。'
              : '下方按商品基础、规格价格、履约仓储等维度汇总当前提交值；暂无历史快照时审核基准会显示为空。'
          }
        />
        <Table
          size="small"
          columns={compareColumns}
          dataSource={auditCompareRows}
          rowKey="key"
          loading={auditDetailLoading}
          pagination={false}
          scroll={{ x: 760, y: 360 }}
          style={{ marginBottom: 16 }}
        />
        <Form form={auditForm} layout="vertical" initialValues={{ status: 1 }}>
          <Form.Item
            label="审核结果"
            name="status"
            rules={[{ required: true, message: '请选择审核结果' }]}
          >
            <Segmented
              block
              options={[
                { label: '通过上架', value: 1 },
                { label: '驳回', value: 2 },
              ]}
            />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, current) => prev.status !== current.status}>
            {({ getFieldValue }) => (
              <Form.Item
                label="审核备注"
                name="auditRemark"
                rules={getFieldValue('status') === 2 ? [{ required: true, message: '驳回时请输入审核备注' }] : []}
              >
                <Input.TextArea rows={4} placeholder="请输入审核说明，方便后续追踪与沟通。" />
              </Form.Item>
            )}
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
