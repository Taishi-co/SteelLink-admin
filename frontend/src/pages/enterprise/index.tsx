import { useState } from 'react'
import { Card, Table, Tag, Button, Space } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useQuery } from '@tanstack/react-query'
import { enterpriseApi, EnterpriseItem, PageParams } from '../../services/api'

export default function EnterpriseList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })

  const { data, isLoading } = useQuery({
    queryKey: ['enterprises', params],
    queryFn: () => enterpriseApi.list(params),
  })

  const columns: ColumnsType<EnterpriseItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Short Name', dataIndex: 'shortName', key: 'shortName' },
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
      render: () => (
        <Space>
          <Button type="link" size="small">View</Button>
          <Button type="link" size="small">Audit</Button>
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
