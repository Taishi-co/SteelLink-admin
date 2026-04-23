import { useState } from 'react'
import { Card, Table, Button, Space, Modal, Select, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { enterpriseApi, EnterpriseItem, PageParams } from '../../services/api'
import { enterpriseStatusMeta, renderStatusTag } from '../../utils/status'

export default function EnterpriseList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['enterprises', params],
    queryFn: () => enterpriseApi.list(params),
  })

  const auditMutation = useMutation({
    mutationFn: ({ id, status, auditRemark }: { id: number; status: number; auditRemark?: string }) =>
      enterpriseApi.audit(id, { status, auditRemark }),
    onSuccess: () => {
      message.success('企业审核状态已更新')
      queryClient.invalidateQueries({ queryKey: ['enterprises'] })
    },
  })

  const handleAudit = (record: EnterpriseItem, status: number) => {
    Modal.confirm({
      title: status === 1 ? '确认通过企业审核？' : '确认驳回企业审核？',
      content: `${record.name}（${record.creditCode || '无信用代码'}）`,
      onOk: () => auditMutation.mutate({ id: record.id, status }),
    })
  }

  const columns: ColumnsType<EnterpriseItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Enterprise Code', dataIndex: 'shortName', key: 'shortName' },
    { title: 'Credit Code', dataIndex: 'creditCode', key: 'creditCode' },
    { title: 'Contact', dataIndex: 'contactPerson', key: 'contactPerson' },
    { title: 'Phone', dataIndex: 'contactPhone', key: 'contactPhone' },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: number) => renderStatusTag(status, enterpriseStatusMeta),
    },
    {
      title: '审核备注',
      dataIndex: 'auditRemark',
      key: 'auditRemark',
      width: 220,
      render: (value?: string) => value || '-',
    },
    {
      title: 'Created At',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 180,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" disabled={record.status === 1} onClick={() => handleAudit(record, 1)}>
            通过
          </Button>
          <Button type="link" danger size="small" disabled={record.status === 2} onClick={() => handleAudit(record, 2)}>
            驳回
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <Card title="企业审核">
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
        dataSource={data?.data.data.list || []}
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: params.page,
          pageSize: params.pageSize,
          total: data?.data.data.total || 0,
          onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
        }}
      />
    </Card>
  )
}
