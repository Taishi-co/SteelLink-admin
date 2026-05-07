import { useState } from 'react'
import { Button, Card, Descriptions, Drawer, Form, Input, Modal, Segmented, Select, Space, Table, Tag, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { enterpriseServiceApi, PageParams, ServiceApplicationItem } from '../../services/api'
import { formatDateTime } from '../../utils/format'

const serviceMeta: Record<string, { text: string; color: string }> = {
  sales: { text: '销售服务', color: 'blue' },
  procurement: { text: '采购服务', color: 'green' },
  logistics: { text: '物流服务', color: 'cyan' },
  warehouse: { text: '仓储服务', color: 'purple' },
  finance: { text: '金融服务', color: 'gold' },
}

const enterpriseRoleMeta: Record<string, string> = {
  seller: '供应商',
  buyer: '采购商',
  logistics: '物流方',
  warehouse: '仓储方',
  fund: '资金方',
}

const statusMeta: Record<number, { text: string; color: string }> = {
  1: { text: '审核中', color: 'processing' },
  2: { text: '已通过', color: 'success' },
  3: { text: '已驳回', color: 'error' },
}

const EMPTY_ROWS: ServiceApplicationItem[] = []

export default function EnterpriseServiceApplications() {
  const [params, setParams] = useState<PageParams & { service?: string }>({ page: 1, pageSize: 10 })
  const [selected, setSelected] = useState<ServiceApplicationItem | null>(null)
  const [auditTarget, setAuditTarget] = useState<ServiceApplicationItem | null>(null)
  const [auditForm] = Form.useForm<{ status: number; reason?: string }>()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['enterprise-service-applications', params],
    queryFn: () => enterpriseServiceApi.listApplications(params),
  })

  const rows = data?.data.data.list ?? EMPTY_ROWS
  const total = data?.data.data.total ?? 0

  const { mutate: audit, isPending } = useMutation({
    mutationFn: ({ id, values }: { id: number; values: { status: number; reason?: string } }) =>
      enterpriseServiceApi.auditApplication(id, values),
    onSuccess: () => {
      message.success('服务申请审核完成')
      setAuditTarget(null)
      auditForm.resetFields()
      void queryClient.invalidateQueries({ queryKey: ['enterprise-service-applications'] })
    },
  })

  const columns: ColumnsType<ServiceApplicationItem> = [
    { title: '申请ID', dataIndex: 'id', width: 90 },
    { title: '企业', dataIndex: 'enterpriseName', width: 220, render: (v, r) => v || r.enterpriseCode },
    { title: '企业代码', dataIndex: 'enterpriseCode', width: 150 },
    {
      title: '平台服务',
      dataIndex: 'service',
      width: 120,
      render: (value: string) => <Tag color={serviceMeta[value]?.color}>{serviceMeta[value]?.text || value}</Tag>,
    },
    {
      title: '供应链角色',
      dataIndex: 'enterpriseRole',
      width: 120,
      render: (value: string) => enterpriseRoleMeta[value] || value,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 110,
      render: (value: number) => <Tag color={statusMeta[value]?.color}>{statusMeta[value]?.text || '未知'}</Tag>,
    },
    { title: '提交时间', dataIndex: 'createdAt', width: 180, render: formatDateTime },
    {
      title: '操作',
      width: 160,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => setSelected(record)}>查看</Button>
          {record.status === 1 && (
            <Button
              type="link"
              size="small"
              onClick={() => {
                setAuditTarget(record)
                auditForm.setFieldsValue({ status: 2, reason: '' })
              }}
            >
              审核
            </Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card title="服务权限审核" extra={<span>共 {total} 条</span>}>
        <Space className="mb-4" wrap>
          <Select
            allowClear
            placeholder="筛选平台服务"
            style={{ width: 180 }}
            value={params.service}
            onChange={(service) => setParams({ ...params, page: 1, service })}
            options={Object.entries(serviceMeta).map(([value, meta]) => ({ value, label: meta.text }))}
          />
          <Select
            allowClear
            placeholder="筛选审核状态"
            style={{ width: 180 }}
            value={params.status}
            onChange={(status) => setParams({ ...params, page: 1, status })}
            options={[
              { value: 1, label: '审核中' },
              { value: 2, label: '已通过' },
              { value: 3, label: '已驳回' },
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

      <Drawer title="服务申请详情" width={620} open={!!selected} onClose={() => setSelected(null)} destroyOnHidden>
        {selected && (
          <Descriptions bordered column={1} size="small">
            <Descriptions.Item label="企业">{selected.enterpriseName || selected.enterpriseCode}</Descriptions.Item>
            <Descriptions.Item label="平台服务">{serviceMeta[selected.service]?.text || selected.service}</Descriptions.Item>
            <Descriptions.Item label="供应链角色">{enterpriseRoleMeta[selected.enterpriseRole] || selected.enterpriseRole}</Descriptions.Item>
            <Descriptions.Item label="审核状态">
              <Tag color={statusMeta[selected.status]?.color}>{statusMeta[selected.status]?.text}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="审核意见">{selected.auditReason || '-'}</Descriptions.Item>
            <Descriptions.Item label="提交材料">
              <Space direction="vertical">
                {selected.documents?.length ? selected.documents.map((doc) => (
                  <a key={doc.id} href={doc.fileUrl} target="_blank" rel="noreferrer">
                    {doc.documentType}
                  </a>
                )) : '-'}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="提交时间">{formatDateTime(selected.createdAt)}</Descriptions.Item>
            <Descriptions.Item label="审核时间">{formatDateTime(selected.auditedAt)}</Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>

      <Modal
        title={`审核服务申请${auditTarget ? ` · ${serviceMeta[auditTarget.service]?.text}` : ''}`}
        open={!!auditTarget}
        onCancel={() => setAuditTarget(null)}
        onOk={() => auditForm.validateFields().then((values) => auditTarget && audit({ id: auditTarget.id, values }))}
        confirmLoading={isPending}
        okText="提交审核"
        destroyOnHidden
      >
        <Form form={auditForm} layout="vertical" initialValues={{ status: 2 }}>
          <Form.Item label="审核结果" name="status" rules={[{ required: true }]}>
            <Segmented block options={[{ label: '通过', value: 2 }, { label: '驳回', value: 3 }]} />
          </Form.Item>
          <Form.Item label="审核意见" name="reason">
            <Input.TextArea rows={4} placeholder="请输入审核意见" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
