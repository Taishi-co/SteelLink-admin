import { Card, Table, Button, Space, Tag } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useQuery } from '@tanstack/react-query'
import { menuApi, MenuItem } from '../../../services/api'

export default function MenuList() {
  const { data, isLoading } = useQuery({
    queryKey: ['menus'],
    queryFn: () => menuApi.list(),
  })

  const columns: ColumnsType<MenuItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Path', dataIndex: 'path', key: 'path' },
    { title: 'Icon', dataIndex: 'icon', key: 'icon' },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: number) => (
        <Tag color={type === 1 ? 'blue' : 'orange'}>
          {type === 1 ? 'Menu' : 'Button'}
        </Tag>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: number) => (
        <Tag color={status === 1 ? 'green' : 'red'}>
          {status === 1 ? 'Visible' : 'Hidden'}
        </Tag>
      ),
    },
    { title: 'Sort', dataIndex: 'sort', key: 'sort', width: 80 },
    {
      title: 'Actions',
      key: 'actions',
      width: 150,
      render: () => (
        <Space>
          <Button type="link" icon={<EditOutlined />} size="small">
            Edit
          </Button>
          <Button type="link" danger icon={<DeleteOutlined />} size="small">
            Delete
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <Card
      title="Menu Management"
      extra={
        <Button type="primary" icon={<PlusOutlined />}>
          Add Menu
        </Button>
      }
    >
      <Table
        columns={columns}
        dataSource={data?.data.data.list || []}
        rowKey="id"
        loading={isLoading}
        pagination={false}
      />
    </Card>
  )
}
