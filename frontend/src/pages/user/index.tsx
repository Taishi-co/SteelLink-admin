import { useState } from 'react'
import { Card, Table, Tag, Button, Space, Input, Drawer, Descriptions, Popconfirm, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi, UserItem, PageParams } from '../../services/api'
import { formatDateTime } from '../../utils/format'
import { SearchOutlined, EyeOutlined } from '@ant-design/icons'

export default function UserList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const [keyword, setKeyword] = useState('')
  const [selectedUser, setSelectedUser] = useState<UserItem | null>(null)
  const queryClient = useQueryClient()

  const queryParams = { ...params, keyword: keyword || undefined }

  const { data, isLoading } = useQuery({
    queryKey: ['users', queryParams],
    queryFn: () => userApi.list(queryParams),
  })

  const { mutate: toggleStatus } = useMutation({
    mutationFn: ({ id, status }: { id: number; status: number }) => userApi.updateStatus(id, status),
    onSuccess: () => {
      message.success('状态已更新')
      void queryClient.invalidateQueries({ queryKey: ['users'] })
    },
    onError: () => message.error('操作失败'),
  })

  const columns: ColumnsType<UserItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: '账号', dataIndex: 'username', key: 'username' },
    { title: '姓名', dataIndex: 'realName', key: 'realName' },
    { title: '手机号', dataIndex: 'phone', key: 'phone', width: 130 },
    { title: '邮箱', dataIndex: 'email', key: 'email' },
    {
      title: '企业ID', dataIndex: 'enterpriseId', key: 'enterpriseId', width: 80,
      render: (v?: number) => v || '-',
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80,
      render: (status: number) => <Tag color={status === 1 ? 'green' : 'red'}>{status === 1 ? '启用' : '停用'}</Tag>,
    },
    {
      title: '注册时间', dataIndex: 'createdAt', key: 'createdAt', width: 170,
      render: (v: string) => formatDateTime(v),
    },
    {
      title: '操作', key: 'actions', width: 140,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setSelectedUser(record)}>详情</Button>
          <Popconfirm
            title={`确认${record.status === 1 ? '停用' : '启用'}该用户？`}
            onConfirm={() => toggleStatus({ id: record.id, status: record.status === 1 ? 0 : 1 })}
          >
            <Button type="link" size="small" danger={record.status === 1}>
              {record.status === 1 ? '停用' : '启用'}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card title="用户管理">
        <Space className="mb-4" wrap>
          <Input
            placeholder="搜索账号 / 手机号 / 姓名"
            prefix={<SearchOutlined />}
            allowClear
            style={{ width: 260 }}
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onPressEnter={() => setParams({ ...params, page: 1 })}
          />
          <Button type="primary" onClick={() => setParams({ ...params, page: 1 })}>搜索</Button>
        </Space>

        <Table
          columns={columns}
          dataSource={data?.data.data.list || []}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1000 }}
          pagination={{
            current: params.page,
            pageSize: params.pageSize,
            total: data?.data.data.total || 0,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
          }}
        />
      </Card>

      <Drawer
        title={`用户详情 · ${selectedUser?.username || ''}`}
        width={480}
        open={selectedUser !== null}
        onClose={() => setSelectedUser(null)}
        destroyOnClose
      >
        {selectedUser && (
          <Descriptions bordered column={1} size="small">
            <Descriptions.Item label="ID">{selectedUser.id}</Descriptions.Item>
            <Descriptions.Item label="账号">{selectedUser.username}</Descriptions.Item>
            <Descriptions.Item label="姓名">{selectedUser.realName || '-'}</Descriptions.Item>
            <Descriptions.Item label="手机号">{selectedUser.phone || '-'}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{selectedUser.email || '-'}</Descriptions.Item>
            <Descriptions.Item label="企业 ID">{selectedUser.enterpriseId || '未绑定'}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={selectedUser.status === 1 ? 'green' : 'red'}>{selectedUser.status === 1 ? '启用' : '停用'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="注册时间">{formatDateTime(selectedUser.createdAt)}</Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>
    </>
  )
}
