import { useState } from 'react'
import { Button, Card, Descriptions, Form, Input, Modal, Segmented, Select, Space, Table, Tag, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { PageParams, supplyDemandApi, SupplyDemandItem, SupplyDemandType } from '../../services/api'
import { formatDateTime } from '../../utils/format'
import { canAudit } from '../../utils/permissions'
import { useAuthStore } from '../../store/useAuthStore'

const typeMeta: Record<SupplyDemandType, { text: string; color: string }> = {
  supply: { text: '供应', color: 'blue' },
  demand: { text: '求购', color: 'green' },
  warehouse: { text: '仓储', color: 'purple' },
  logistics: { text: '物流', color: 'orange' },
}

const statusMeta: Record<number, { text: string; color: string }> = {
  0: { text: '草稿', color: 'default' },
  1: { text: '已通过，商城可见', color: 'success' },
  2: { text: '已撮合', color: 'cyan' },
  3: { text: '已关闭', color: 'default' },
  4: { text: '待审核，商城不可见', color: 'processing' },
  5: { text: '已驳回，商城不可见', color: 'error' },
}

const EMPTY_ROWS: SupplyDemandItem[] = []

type AuditFormValues = { status: number; auditRemark?: string }

function renderText(value?: string) {
  return value || '-'
}

function renderLongText(value?: string) {
  return value ? <Typography.Text ellipsis={{ tooltip: value }}>{value}</Typography.Text> : '-'
}

function formatRegion(record: SupplyDemandItem) {
  return [record.province, record.city, record.district].filter(Boolean).join(' / ') || '-'
}

function formatQuantity(record: SupplyDemandItem) {
  if (!record.quantity) return '-'
  return `${record.quantity.toLocaleString('zh-CN')} ${record.unit || ''}`.trim()
}

function formatMoney(value?: number, unit?: string) {
  if (!value) return '-'
  return `¥${value.toLocaleString('zh-CN')}${unit ? ` / ${unit}` : ''}`
}

function toAuditPayload(values: AuditFormValues) {
  return {
    status: values.status,
    auditAction: values.status === 1 ? 'approve' as const : 'reject' as const,
    auditRemark: values.auditRemark,
  }
}

export default function SupplyDemandAuditPage() {
  const [params, setParams] = useState<PageParams & { type?: SupplyDemandType }>({ page: 1, pageSize: 10, status: 4 })
  const [target, setTarget] = useState<SupplyDemandItem | null>(null)
  const [form] = Form.useForm<AuditFormValues>()
  const queryClient = useQueryClient()
  const roleId = useAuthStore((state) => state.adminInfo?.roleId)
  const auditAllowed = canAudit(roleId)

  const { data, isLoading } = useQuery({
    queryKey: ['supply-demands', params],
    queryFn: () => supplyDemandApi.list(params),
  })

  const { mutate: audit, isPending } = useMutation({
    mutationFn: ({ id, values }: { id: number; values: AuditFormValues }) =>
      supplyDemandApi.audit(id, toAuditPayload(values)),
    onSuccess: () => {
      message.success('审核完成')
      setTarget(null)
      form.resetFields()
      void queryClient.invalidateQueries({ queryKey: ['supply-demands'] })
    },
  })

  const rows = data?.data.data.list ?? EMPTY_ROWS
  const total = data?.data.data.total ?? 0

  const columns: ColumnsType<SupplyDemandItem> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '类型', dataIndex: 'type', width: 90, render: (value: SupplyDemandType) => <Tag color={typeMeta[value]?.color}>{typeMeta[value]?.text || value}</Tag> },
    { title: '标题', dataIndex: 'title', width: 240, render: renderLongText },
    { title: '品类/材质/规格', width: 220, render: (_, record) => [record.category, record.material, record.spec].filter(Boolean).join(' / ') || '-' },
    { title: '数量', width: 120, render: (_, record) => formatQuantity(record) },
    { title: '目标价', width: 130, render: (_, record) => formatMoney(record.targetPrice, record.targetPrice ? record.unit : undefined) },
    { title: '预算/报价', width: 130, render: (_, record) => formatMoney(record.budget) },
    { title: '区域', width: 160, render: (_, record) => formatRegion(record) },
    { title: '发布人ID', dataIndex: 'createdBy', width: 100 },
    { title: '企业ID', dataIndex: 'enterpriseId', width: 100, render: (value?: number) => value || '-' },
    { title: '状态', dataIndex: 'status', width: 160, render: (value: number) => <Tag color={statusMeta[value]?.color}>{statusMeta[value]?.text || '未知'}</Tag> },
    { title: '审核备注', dataIndex: 'auditRemark', width: 180, render: renderLongText },
    { title: '提交时间', dataIndex: 'createdAt', width: 180, render: formatDateTime },
    {
      title: '操作',
      width: 120,
      fixed: 'right',
      render: (_, record) => auditAllowed && (record.status === 4 || record.status === 5) ? (
        <Button
          type="link"
          size="small"
          onClick={() => {
            setTarget(record)
            form.setFieldsValue({ status: 1, auditRemark: record.auditRemark })
          }}
        >
          {record.status === 5 ? '重新审核' : '审核'}
        </Button>
      ) : '-',
    },
  ]

  return (
    <>
      <Card title="供求信息审核" extra={<span>共 {total} 条</span>}>
        <Space className="mb-4" wrap>
          <Select
            value={params.status}
            allowClear
            placeholder="审核状态"
            style={{ width: 180 }}
            onChange={(status) => setParams({ ...params, page: 1, status })}
            options={[
              { value: 4, label: '待审核' },
              { value: 1, label: '已通过' },
              { value: 2, label: '已撮合' },
              { value: 3, label: '已关闭' },
              { value: 5, label: '已驳回' },
            ]}
          />
          <Select
            value={params.type}
            allowClear
            placeholder="信息类型"
            style={{ width: 160 }}
            onChange={(type) => setParams({ ...params, page: 1, type })}
            options={[
              { value: 'supply', label: '供应' },
              { value: 'demand', label: '求购' },
              { value: 'warehouse', label: '仓储' },
              { value: 'logistics', label: '物流' },
            ]}
          />
          <Input.Search
            allowClear
            placeholder="搜索标题、品类、联系人"
            style={{ width: 260 }}
            onSearch={(keyword) => setParams({ ...params, page: 1, keyword })}
          />
        </Space>
        <Table
          columns={columns}
          dataSource={rows}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1700 }}
          pagination={{
            current: params.page,
            pageSize: params.pageSize,
            total,
            onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
          }}
        />
      </Card>

      <Modal
        title={`审核供求信息${target ? ` · ${target.title}` : ''}`}
        open={!!target}
        onCancel={() => setTarget(null)}
        onOk={() => form.validateFields().then((values) => target && audit({ id: target.id, values }))}
        confirmLoading={isPending}
        okText="提交审核"
        width={760}
      >
        {target && (
          <Descriptions size="small" column={2} bordered className="mb-4">
            <Descriptions.Item label="信息类型">{typeMeta[target.type]?.text || target.type}</Descriptions.Item>
            <Descriptions.Item label="商城状态">
              <Tag color={statusMeta[target.status]?.color}>{statusMeta[target.status]?.text || '未知'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="品类">{target.category}</Descriptions.Item>
            <Descriptions.Item label="数量">{formatQuantity(target)}</Descriptions.Item>
            <Descriptions.Item label="区域">{formatRegion(target)}</Descriptions.Item>
            <Descriptions.Item label="联系人">{renderText(target.contactName)}</Descriptions.Item>
            <Descriptions.Item label="联系电话">{renderText(target.contactPhone)}</Descriptions.Item>
            <Descriptions.Item label="目标价">{formatMoney(target.targetPrice, target.targetPrice ? target.unit : undefined)}</Descriptions.Item>
            <Descriptions.Item label="描述" span={2}>{renderText(target.description)}</Descriptions.Item>
          </Descriptions>
        )}
        <Form form={form} layout="vertical" initialValues={{ status: 1 }}>
          <Form.Item label="审核结果" name="status" rules={[{ required: true }]}>
            <Segmented block options={[{ label: '通过并展示', value: 1 }, { label: '驳回', value: 5 }]} />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, current) => prev.status !== current.status}>
            {({ getFieldValue }) => (
              <Form.Item
                label="审核备注"
                name="auditRemark"
                rules={getFieldValue('status') === 5 ? [{ required: true, message: '驳回时请输入审核备注' }] : []}
              >
                <Input.TextArea rows={4} placeholder="请输入审核说明" />
              </Form.Item>
            )}
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
