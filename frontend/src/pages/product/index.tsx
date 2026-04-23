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
  0: { color: 'orange', text: '待审核' },
  1: { color: 'green', text: '已通过' },
  2: { color: 'red', text: '已驳回' },
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
    onSuccess: () => {
      message.success('商品审核结果已更新')
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
        const meta = statusMeta[status as 0 | 1 | 2] ?? statusMeta[0]
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
                status: record.status === 1 ? 1 : 2,
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
              { label: '待审核', value: 0 },
              { label: '已通过', value: 1 },
              { label: '已驳回', value: 2 },
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
            <Descriptions.Item label="规格">{detail?.spec || '-'}</Descriptions.Item>
            <Descriptions.Item label="单位">{detail?.unit || '-'}</Descriptions.Item>
            <Descriptions.Item label="价格">
              {detail ? `¥${Number(detail.price ?? 0).toFixed(2)}` : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="库存">{detail?.stock ?? '-'}</Descriptions.Item>
            <Descriptions.Item label="审核状态">
              {detail ? (
                <Tag color={statusMeta[detail.status as 0 | 1 | 2]?.color ?? statusMeta[0].color}>
                  {statusMeta[detail.status as 0 | 1 | 2]?.text ?? statusMeta[0].text}
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
        <Form form={auditForm} layout="vertical" initialValues={{ status: 1 }}>
          <Form.Item
            label="审核结果"
            name="status"
            rules={[{ required: true, message: '请选择审核结果' }]}
          >
            <Segmented
              block
              options={[
                { label: '通过', value: 1 },
                { label: '驳回', value: 2 },
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
