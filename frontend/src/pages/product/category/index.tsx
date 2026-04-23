import { Card, Table, Tag, Button, Space } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useQuery } from '@tanstack/react-query'
import { categoryApi, CategoryItem } from '../../../services/api'

export default function CategoryList() {
  const { data, isLoading } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list(),
  })

  const columns: ColumnsType<CategoryItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '类目名称', dataIndex: 'name', key: 'name' },
    { title: '编码', dataIndex: 'code', key: 'code' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: number) => (
        <Tag color={status === 1 ? 'green' : 'red'}>
          {status === 1 ? '启用' : '停用'}
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
      title="类目配置"
      extra={
        <Button type="primary" icon={<PlusOutlined />}>
          新增类目
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
