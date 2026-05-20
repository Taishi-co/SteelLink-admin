import { useMemo, useState } from 'react'
import {
  Button,
  Card,
  Descriptions,
  Drawer,
  Form,
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
import { enterpriseApi, EnterpriseItem, PageParams, ServicePermissionItem } from '../../services/api'
import { formatDateTime } from '../../utils/format'
import { canAudit, canOperate } from '../../utils/permissions'
import { useAuthStore } from '../../store/useAuthStore'

const statusMeta = {
  0: { color: 'orange', text: '待审核' },
  1: { color: 'green', text: '已通过' },
  2: { color: 'red', text: '已驳回' },
} as const

const EMPTY_ENTERPRISE_ROWS: EnterpriseItem[] = []
const EMPTY_SERVICE_ROWS: ServicePermissionItem[] = []

const serviceMeta: Record<string, { text: string; color: string }> = {
  sales: { text: '销售交易能力', color: 'blue' },
  procurement: { text: '采购交易能力', color: 'green' },
  logistics: { text: '物流服务商能力', color: 'cyan' },
  warehouse: { text: '仓储服务商能力', color: 'purple' },
  finance: { text: '金融服务商能力', color: 'gold' },
}

const serviceStatusMeta: Record<number, { text: string; color: string }> = {
  1: { text: '审核中', color: 'processing' },
  2: { text: '已开通', color: 'success' },
  3: { text: '已驳回', color: 'error' },
  4: { text: '已暂停', color: 'warning' },
}

export default function EnterpriseList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [auditTarget, setAuditTarget] = useState<EnterpriseItem | null>(null)
  const [serviceTarget, setServiceTarget] = useState<{ enterpriseId: number; service: string } | null>(null)
  const [auditForm] = Form.useForm<{ status: number; auditRemark?: string }>()
  const [serviceForm] = Form.useForm<{ status: number; reason?: string }>()
  const queryClient = useQueryClient()
  const roleId = useAuthStore((state) => state.adminInfo?.roleId)
  const auditAllowed = canAudit(roleId)
  const operateAllowed = canOperate(roleId)

  const { data, isLoading } = useQuery({
    queryKey: ['enterprises', params],
    queryFn: () => enterpriseApi.list(params),
  })

  const { data: detailData, isFetching: detailLoading } = useQuery({
    queryKey: ['enterprise-detail', selectedId],
    queryFn: () => enterpriseApi.get(selectedId!),
    enabled: selectedId !== null,
  })

  const { data: serviceData, isFetching: serviceLoading } = useQuery({
    queryKey: ['enterprise-services', selectedId],
    queryFn: () => enterpriseApi.services(selectedId!),
    enabled: selectedId !== null,
  })

  const rows = data?.data.data.list ?? EMPTY_ENTERPRISE_ROWS
  const detail = detailData?.data.data
  const serviceRows = serviceData?.data.data ?? EMPTY_SERVICE_ROWS
  const total = data?.data.data.total ?? 0

  const summary = useMemo(
    () =>
      rows.reduce(
        (acc, item) => {
          acc.total += 1
          acc[item.status as 0 | 1 | 2] += 1
          return acc
        },
        { total: 0, 0: 0, 1: 0, 2: 0 },
      ),
    [rows],
  )

  const { mutate: auditEnterprise, isPending: auditSubmitting } = useMutation({
    mutationFn: ({ id, values }: { id: number; values: { status: number; auditRemark?: string } }) =>
      enterpriseApi.audit(id, values),
    onSuccess: () => {
      message.success('企业审核结果已更新')
      setAuditTarget(null)
      auditForm.resetFields()
      void queryClient.invalidateQueries({ queryKey: ['enterprises'] })
      if (selectedId !== null) {
        void queryClient.invalidateQueries({ queryKey: ['enterprise-detail', selectedId] })
      }
    },
  })

  const { mutate: updateServiceStatus, isPending: serviceSubmitting } = useMutation({
    mutationFn: ({ enterpriseId, service, values }: { enterpriseId: number; service: string; values: { status: number; reason?: string } }) =>
      enterpriseApi.updateServiceStatus(enterpriseId, service, values),
    onSuccess: () => {
      message.success('企业能力已更新')
      setServiceTarget(null)
      serviceForm.resetFields()
      void queryClient.invalidateQueries({ queryKey: ['enterprise-services', selectedId] })
    },
  })

  const columns: ColumnsType<EnterpriseItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '企业名称', dataIndex: 'name', key: 'name', width: 220 },
    { title: '统一社会信用代码', dataIndex: 'creditCode', key: 'creditCode', width: 220 },
    { title: '联系人', dataIndex: 'contactPerson', key: 'contactPerson', width: 120 },
    { title: '联系电话', dataIndex: 'contactPhone', key: 'contactPhone', width: 140 },
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
          {auditAllowed && (
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
          )}
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card
        title="企业审核"
        extra={
          <Space size={16}>
            <span>总数 {total}</span>
            <span>当前页待审核 {summary[0]}</span>
            <span>当前页已通过 {summary[1]}</span>
            <span>当前页已驳回 {summary[2]}</span>
          </Space>
        }
      >
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
          scroll={{ x: 1200 }}
          pagination={{
            current: params.page,
            pageSize: params.pageSize,
            total,
            onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
          }}
        />
      </Card>

      <Drawer
        title="企业详情"
        width={560}
        open={selectedId !== null}
        onClose={() => setSelectedId(null)}
        destroyOnHidden
      >
        {detailLoading ? (
          <Skeleton active paragraph={{ rows: 8 }} />
        ) : (
          <Descriptions bordered column={1} size="small">
            <Descriptions.Item label="企业名称">{detail?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="统一社会信用代码">{detail?.creditCode || '-'}</Descriptions.Item>
            <Descriptions.Item label="法定代表人">{detail?.legalPerson || '-'}</Descriptions.Item>
            <Descriptions.Item label="联系人">{detail?.contactPerson || '-'}</Descriptions.Item>
            <Descriptions.Item label="联系电话">{detail?.contactPhone || '-'}</Descriptions.Item>
            <Descriptions.Item label="地址">{detail?.address || '-'}</Descriptions.Item>
            <Descriptions.Item label="营业执照">
              {detail?.businessLicense ? (
                <a href={detail.businessLicense} target="_blank" rel="noreferrer">
                  查看附件
                </a>
              ) : (
                '-'
              )}
            </Descriptions.Item>
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
            <Descriptions.Item label="审核时间">{formatDateTime(detail?.auditedAt)}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{formatDateTime(detail?.createdAt)}</Descriptions.Item>
            <Descriptions.Item label="更新时间">{formatDateTime(detail?.updatedAt)}</Descriptions.Item>
          </Descriptions>
        )}

        <div className="mt-4">
          <Card size="small" title="企业能力" loading={serviceLoading}>
            <Table
              size="small"
              pagination={false}
              rowKey="service"
              dataSource={Object.keys(serviceMeta).map((service) => serviceRows.find((row) => row.service === service) || {
                id: 0,
                enterpriseId: selectedId || 0,
	                enterpriseCode: detail?.name || '',
	                service,
	                enterpriseRole: '',
	                status: 0,
                updatedAt: '',
              })}
              columns={[
                {
                  title: '服务',
                  dataIndex: 'service',
                  render: (service: string) => <Tag color={serviceMeta[service]?.color}>{serviceMeta[service]?.text}</Tag>,
                },
                {
                  title: '状态',
                  dataIndex: 'status',
                  render: (status: number) => status ? <Tag color={serviceStatusMeta[status]?.color}>{serviceStatusMeta[status]?.text}</Tag> : <Tag>未开通</Tag>,
                },
                {
                  title: '操作',
                  render: (_, record) => (
                    operateAllowed ? (
                      <Button
                        type="link"
                        size="small"
                        onClick={() => {
                          setServiceTarget({ enterpriseId: selectedId!, service: record.service })
                          serviceForm.setFieldsValue({ status: record.status || 2, reason: record.auditReason || '' })
                        }}
                      >
                        控制
                      </Button>
                    ) : '-'
                  ),
                },
              ]}
            />
          </Card>
        </div>
      </Drawer>

      <Modal
        title={`企业审核${auditTarget ? ` · ${auditTarget.name}` : ''}`}
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
              auditEnterprise({ id: auditTarget.id, values })
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

      <Modal
        title={`企业能力控制${serviceTarget ? ` · ${serviceMeta[serviceTarget.service]?.text}` : ''}`}
        open={serviceTarget !== null}
        onCancel={() => {
          setServiceTarget(null)
          serviceForm.resetFields()
        }}
        onOk={() => {
          serviceForm.validateFields().then((values) => {
            if (!serviceTarget) return
            updateServiceStatus({ enterpriseId: serviceTarget.enterpriseId, service: serviceTarget.service, values })
          })
        }}
        confirmLoading={serviceSubmitting}
        okText="保存"
        destroyOnHidden
      >
        <Form form={serviceForm} layout="vertical" initialValues={{ status: 2 }}>
          <Form.Item label="权限状态" name="status" rules={[{ required: true, message: '请选择权限状态' }]}>
            <Segmented
              block
              options={[
                { label: '开通', value: 2 },
                { label: '驳回', value: 3 },
                { label: '暂停', value: 4 },
              ]}
            />
          </Form.Item>
          <Form.Item label="原因说明" name="reason">
            <Input.TextArea rows={4} placeholder="请输入开通、驳回或暂停原因" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
