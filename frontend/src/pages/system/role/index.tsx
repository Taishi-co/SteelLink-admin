import { useState } from 'react'
import { Card, Table, Button, Space, Tag, Popconfirm, message } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { roleApi, RoleItem, PageParams } from '../../../services/api'
import { formatDateTime } from '../../../utils/format'

export default function RoleList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['roles', params],
    queryFn: () => roleApi.list(params),
  })

  const { mutate: deleteRole } = useMutation({
    mutationFn: roleApi.delete,
    onSuccess: () => {
      message.success('删除成功')
      queryClient.invalidateQueries({ queryKey: ['roles'] })
    },
  })

  const columns: ColumnsType<RoleItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '角色名称', dataIndex: 'name', key: 'name' },
    { title: '角色编码', dataIndex: 'code', key: 'code' },
    { title: '说明', dataIndex: 'description', key: 'description' },
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
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (value: string) => formatDateTime(value),
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} size="small">
            编辑
          </Button>
          <Popconfirm
            title="确认删除该角色吗？"
            onConfirm={() => deleteRole(record.id)}
          >
            <Button type="link" danger icon={<DeleteOutlined />} size="small">
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <Card
      title="角色权限"
      extra={
        <Button type="primary" icon={<PlusOutlined />}>
          新增角色
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
