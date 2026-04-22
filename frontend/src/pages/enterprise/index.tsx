import { useState } from 'react'
import { Card, Table, Tag, Button, Space, Modal, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { enterpriseApi, EnterpriseItem, PageParams } from '../../services/api'

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
      message.success('审核已更新')
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
      render: (status: number) => {
        const colors = { 0: 'orange', 1: 'green', 2: 'red' }
        const texts = { 0: 'Pending', 1: 'Approved', 2: 'Rejected' }
        return <Tag color={colors[status as 0 | 1 | 2]}>{texts[status as 0 | 1 | 2]}</Tag>
      },
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
      width: 120,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => handleAudit(record, 1)}>Approve</Button>
          <Button type="link" danger size="small" onClick={() => handleAudit(record, 2)}>Reject</Button>
        </Space>
      ),
    },
  ]

  return (
    <Card title="Enterprise Management">
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
