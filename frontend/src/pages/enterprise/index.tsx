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
  Skeleton,
  Space,
  Table,
  Tag,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { enterpriseApi, EnterpriseItem, PageParams } from '../../services/api'
import { formatDateTime } from '../../utils/format'

const statusMeta = {
  0: { color: 'orange', text: '待审核' },
  1: { color: 'green', text: '已通过' },
  2: { color: 'red', text: '已驳回' },
} as const

export default function EnterpriseList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [auditTarget, setAuditTarget] = useState<EnterpriseItem | null>(null)
  const [auditForm] = Form.useForm<{ status: number; auditRemark?: string }>()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['enterprises', params],
    queryFn: () => enterpriseApi.list(params),
  })

  const { data: detailData, isFetching: detailLoading } = useQuery({
    queryKey: ['enterprise-detail', selectedId],
    queryFn: () => enterpriseApi.get(selectedId!),
    enabled: selectedId !== null,
  })

  const rows = data?.data.data.list ?? []
  const detail = detailData?.data.data
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
    </>
  )
}
