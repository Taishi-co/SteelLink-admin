import { useState } from 'react'
import { Card, Table, Button, Space, Tag, Popconfirm, message } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { adminApi, AdminItem, PageParams } from '../../../services/api'

export default function AdminList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['admins', params],
    queryFn: () => adminApi.list(params),
  })

  const { mutate: deleteAdmin } = useMutation({
    mutationFn: adminApi.delete,
    onSuccess: () => {
      message.success('Deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['admins'] })
    },
  })

  const columns: ColumnsType<AdminItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: 'Username', dataIndex: 'username', key: 'username' },
    { title: 'Real Name', dataIndex: 'realName', key: 'realName' },
    { title: 'Email', dataIndex: 'email', key: 'email' },
    { title: 'Phone', dataIndex: 'phone', key: 'phone' },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: number) => (
        <Tag color={status === 1 ? 'green' : 'red'}>
          {status === 1 ? 'Enabled' : 'Disabled'}
        </Tag>
      ),
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
      width: 150,
      render: (_, record) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} size="small">
            Edit
          </Button>
          <Popconfirm
            title="Are you sure to delete this admin?"
            onConfirm={() => deleteAdmin(record.id)}
          >
            <Button type="link" danger icon={<DeleteOutlined />} size="small">
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <Card
      title="Admin Management"
      extra={
        <Button type="primary" icon={<PlusOutlined />}>
          Add Admin
        </Button>
      }
    >
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
