import { useState } from 'react'
import { Card, Table, Tag, Button, Space } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useQuery } from '@tanstack/react-query'
import { userApi, UserItem, PageParams } from '../../services/api'
import { formatDateTime } from '../../utils/format'

export default function UserList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })

  const { data, isLoading } = useQuery({
    queryKey: ['users', params],
    queryFn: () => userApi.list(params),
  })

  const columns: ColumnsType<UserItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '账号', dataIndex: 'username', key: 'username' },
    { title: '姓名', dataIndex: 'realName', key: 'realName' },
    { title: '手机号', dataIndex: 'phone', key: 'phone' },
    { title: '邮箱', dataIndex: 'email', key: 'email' },
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
      width: 120,
      render: () => (
        <Space>
          <Button type="link" size="small">查看</Button>
        </Space>
      ),
    },
  ]

  return (
    <Card title="用户管理">
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
