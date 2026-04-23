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
    { title: '菜单名称', dataIndex: 'name', key: 'name' },
    { title: '路径', dataIndex: 'path', key: 'path' },
    { title: '图标', dataIndex: 'icon', key: 'icon' },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: number) => (
        <Tag color={type === 1 ? 'blue' : 'orange'}>
          {type === 1 ? '菜单' : '按钮'}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: number) => (
        <Tag color={status === 1 ? 'green' : 'red'}>
          {status === 1 ? '显示' : '隐藏'}
        </Tag>
      ),
    },
    { title: '排序', dataIndex: 'sort', key: 'sort', width: 80 },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: () => (
        <Space>
          <Button type="link" icon={<EditOutlined />} size="small">
            编辑
          </Button>
          <Button type="link" danger icon={<DeleteOutlined />} size="small">
            删除
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <Card
      title="菜单配置"
      extra={
        <Button type="primary" icon={<PlusOutlined />}>
          新增菜单
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
