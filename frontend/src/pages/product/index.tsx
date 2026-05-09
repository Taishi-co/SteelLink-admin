import { useState } from 'react'
import {
  Button,
  Card,
  Descriptions,
  Drawer,
  Form,
  Image,
  Input,
  Modal,
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
import { productApi, ProductItem, PageParams } from '../../services/api'
import { formatDateTime } from '../../utils/format'

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

export default function ProductList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [auditTarget, setAuditTarget] = useState<ProductItem | null>(null)
  const [auditForm] = Form.useForm<{ status: number; auditRemark?: string }>()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['products', params],
    queryFn: () => productApi.list(params),
  })

  const { data: detailData, isFetching: detailLoading } = useQuery({
    queryKey: ['product-detail', selectedId],
    queryFn: () => productApi.get(selectedId!),
    enabled: selectedId !== null,
  })

  const rows = data?.data.data.list ?? []
  const detail = detailData?.data.data

  const { mutate: auditProduct, isPending: auditSubmitting } = useMutation({
    mutationFn: ({ id, values }: { id: number; values: { status: number; auditRemark?: string } }) =>
      productApi.audit(id, values),
    onSuccess: (_, variables) => {
      message.success(
        variables.values.status === 2
          ? '商品审核已通过，当前已进入商城展示范围'
          : '商品已驳回，当前不会进入商城展示',
      )
      setAuditTarget(null)
      auditForm.resetFields()
      void queryClient.invalidateQueries({ queryKey: ['products'] })
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
          <Button
            type="link"
            size="small"
            onClick={() => {
              setAuditTarget(record)
              auditForm.setFieldsValue({
                status: record.status === 4 ? 4 : 2,
                auditRemark: record.auditRemark,
              })
            }}
          >
            审核
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card title="商品管理">
        <Space className="mb-4" wrap>
          <Select
            value={params.status}
            placeholder="筛选审核状态"
            allowClear
            style={{ width: 180 }}
            onChange={(status) => setParams({ ...params, page: 1, status })}
            options={[
              { label: '草稿', value: 0 },
              { label: '待审核', value: 1 },
              { label: '已上架', value: 2 },
              { label: '已下架', value: 3 },
              { label: '审核拒绝', value: 4 },
            ]}
          />
        </Space>
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
          <Descriptions bordered column={1} size="small">
            <Descriptions.Item label="商品名称">{detail?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="货品类型">
              {detail ? (
                <Tag color={goodsTypeMeta[detail.goodsType || 'spot'].color}>
                  {goodsTypeMeta[detail.goodsType || 'spot'].text}
                </Tag>
              ) : (
                '-'
              )}
            </Descriptions.Item>
            <Descriptions.Item label="主品类">{detail?.category || '-'}</Descriptions.Item>
            <Descriptions.Item label="子品类">{detail?.subCategory || '-'}</Descriptions.Item>
            <Descriptions.Item label="钢厂">{detail?.steelMill || '-'}</Descriptions.Item>
            <Descriptions.Item label="规格">{detail?.spec || '-'}</Descriptions.Item>
            <Descriptions.Item label="材质">{detail?.material || '-'}</Descriptions.Item>
            <Descriptions.Item label="单位">{detail?.unit || '-'}</Descriptions.Item>
            <Descriptions.Item label="价格">
              {detail ? `¥${Number(detail.price ?? 0).toFixed(2)}` : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="库存">{detail?.stock ?? '-'}</Descriptions.Item>
            <Descriptions.Item label="最小起订量">{detail?.minOrderQty ?? '-'}</Descriptions.Item>
            <Descriptions.Item label="交货周期">{detail?.deliveryDays ? `${detail.deliveryDays} 天` : '-'}</Descriptions.Item>
            <Descriptions.Item label="报价有效期">{detail?.quoteValidity ? `${detail.quoteValidity} 小时` : '-'}</Descriptions.Item>
            <Descriptions.Item label="捆包号">{detail?.bundleNo || '-'}</Descriptions.Item>
            <Descriptions.Item label="绑定仓储资源">
              {detail?.warehouseServiceId ? `#${detail.warehouseServiceId}` : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="仓储方企业">
              {detail?.warehouseEnterpriseId ? `#${detail.warehouseEnterpriseId}` : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="仓库名称">{detail?.warehouseName || '-'}</Descriptions.Item>
            <Descriptions.Item label="仓库编码">{detail?.warehouseCode || '-'}</Descriptions.Item>
            <Descriptions.Item label="仓库地区">
              {[detail?.province, detail?.city, detail?.district].filter(Boolean).join(' / ') || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="仓库地址">{detail?.address || '-'}</Descriptions.Item>
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
            <Descriptions.Item label="商品描述">{detail?.description || '-'}</Descriptions.Item>
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
        )}
      </Drawer>

      <Modal
        title={`商品审核${auditTarget ? ` · ${auditTarget.name}` : ''}`}
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
        <Form form={auditForm} layout="vertical" initialValues={{ status: 2 }}>
          <Form.Item
            label="审核结果"
            name="status"
            rules={[{ required: true, message: '请选择审核结果' }]}
          >
            <Segmented
              block
              options={[
                { label: '通过上架', value: 2 },
                { label: '驳回', value: 4 },
              ]}
            />
          </Form.Item>
          <Form.Item label="审核备注" name="auditRemark">
            <Input.TextArea rows={4} placeholder="请输入审核说明，方便后续追踪与沟通。" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
