import { useState } from 'react'
import { Button, Card, Form, Input, Modal, Segmented, Select, Space, Table, Tag, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { PageParams, serviceOfferApi, ServiceOfferItem } from '../../services/api'
import { formatDateTime } from '../../utils/format'

const statusMeta: Record<number, { text: string; color: string }> = {
  1: { text: '待审核', color: 'processing' },
  2: { text: '已通过', color: 'success' },
  3: { text: '已驳回', color: 'error' },
  4: { text: '已下线', color: 'default' },
}

const EMPTY_ROWS: ServiceOfferItem[] = []

export default function ServiceOfferAudit({ type }: { type: 'warehouse' | 'logistics' }) {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10, status: 1 })
  const [target, setTarget] = useState<ServiceOfferItem | null>(null)
  const [form] = Form.useForm<{ status: number; auditRemark?: string }>()
  const queryClient = useQueryClient()
  const isWarehouse = type === 'warehouse'

  const { data, isLoading } = useQuery({
    queryKey: ['service-offers', type, params],
    queryFn: () => isWarehouse ? serviceOfferApi.listWarehouse(params) : serviceOfferApi.listLogistics(params),
  })

  const { mutate: audit, isPending } = useMutation({
    mutationFn: ({ id, values }: { id: number; values: { status: number; auditRemark?: string } }) =>
      isWarehouse ? serviceOfferApi.auditWarehouse(id, values) : serviceOfferApi.auditLogistics(id, values),
    onSuccess: () => {
      message.success('审核完成')
      setTarget(null)
      form.resetFields()
      void queryClient.invalidateQueries({ queryKey: ['service-offers', type] })
    },
  })

  const rows = data?.data.data.list ?? EMPTY_ROWS
  const total = data?.data.data.total ?? 0

  const columns: ColumnsType<ServiceOfferItem> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: isWarehouse ? '仓库名称' : '物流服务', dataIndex: 'name', width: 220 },
    { title: '企业ID', dataIndex: 'enterpriseId', width: 100 },
    ...(isWarehouse ? [
      { title: '区域', width: 180, render: (_: unknown, r: ServiceOfferItem) => [r.province, r.city].filter(Boolean).join(' / ') || '-' },
      { title: '可用库容', dataIndex: 'availableCapacity', width: 120 },
    ] : [
      { title: '线路', width: 220, render: (_: unknown, r: ServiceOfferItem) => `${r.routeStart || '-'} -> ${r.routeEnd || '-'}` },
      { title: '车型', dataIndex: 'vehicleType', width: 120 },
    ]),
    { title: '价格规则', dataIndex: 'priceRule', width: 220, render: (v?: string) => v || '-' },
    { title: '状态', dataIndex: 'status', width: 110, render: (v: number) => <Tag color={statusMeta[v]?.color}>{statusMeta[v]?.text || '未知'}</Tag> },
    { title: '提交时间', dataIndex: 'createdAt', width: 180, render: formatDateTime },
    {
      title: '操作',
      width: 120,
      render: (_, record) => record.status === 1 ? (
        <Button
          type="link"
          size="small"
          onClick={() => {
            setTarget(record)
            form.setFieldsValue({ status: 1, auditRemark: record.auditRemark })
          }}
        >
          审核
        </Button>
      ) : '-',
    },
  ]

  return (
    <>
      <Card title={isWarehouse ? '仓储资源审核' : '物流资源审核'} extra={<span>共 {total} 条</span>}>
        <Space className="mb-4" wrap>
          <Select
            value={params.status}
            allowClear
            placeholder="审核状态"
            style={{ width: 160 }}
            onChange={(status) => setParams({ ...params, page: 1, status })}
            options={[
              { value: 1, label: '待审核' },
              { value: 2, label: '已通过' },
              { value: 3, label: '已驳回' },
              { value: 4, label: '已下线' },
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
            total,
            onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
          }}
        />
      </Card>
      <Modal
        title={`审核${isWarehouse ? '仓储' : '物流'}资源`}
        open={!!target}
        onCancel={() => setTarget(null)}
        onOk={() => form.validateFields().then((values) => target && audit({ id: target.id, values }))}
        confirmLoading={isPending}
        okText="提交审核"
      >
        <Form form={form} layout="vertical" initialValues={{ status: 1 }}>
          <Form.Item label="审核结果" name="status" rules={[{ required: true }]}>
            <Segmented block options={[{ label: '通过', value: 1 }, { label: '驳回', value: 2 }]} />
          </Form.Item>
          <Form.Item label="审核备注" name="auditRemark">
            <Input.TextArea rows={4} placeholder="请输入审核说明" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
