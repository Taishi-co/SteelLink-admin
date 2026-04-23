import { useState } from 'react'
import { Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, message } from 'antd'
import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { adminApi, type AdminItem, type PageParams } from '../../../services/api'
import { formatDateTime } from '../../../utils/format'

interface AdminFormValues {
  username: string
  password?: string
  realName: string
  email?: string
  phone?: string
  status: number
  roleId?: number
}

export default function AdminList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const [editingAdmin, setEditingAdmin] = useState<AdminItem | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm<AdminFormValues>()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['admins', params],
    queryFn: () => adminApi.list(params),
  })

  const { mutate: createAdmin, isPending: creating } = useMutation({
    mutationFn: (values: AdminFormValues) => adminApi.create({ ...values, password: values.password || '' }),
    onSuccess: () => {
      message.success('管理员已创建')
      closeModal()
      void queryClient.invalidateQueries({ queryKey: ['admins'] })
    },
  })

  const { mutate: updateAdmin, isPending: updating } = useMutation({
    mutationFn: ({ id, values }: { id: number; values: AdminFormValues }) => adminApi.update(id, values),
    onSuccess: () => {
      message.success('管理员信息已更新')
      closeModal()
      void queryClient.invalidateQueries({ queryKey: ['admins'] })
    },
  })

  const { mutate: deleteAdmin } = useMutation({
    mutationFn: adminApi.delete,
    onSuccess: () => {
      message.success('删除成功')
      void queryClient.invalidateQueries({ queryKey: ['admins'] })
    },
    onError: (error: unknown) => {
      const messageText =
        typeof error === 'object' &&
        error !== null &&
        'response' in error &&
        (error as { response?: { data?: { message?: string } } }).response?.data?.message
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : '删除失败'
      message.error(messageText)
    },
  })

  const closeModal = () => {
    setModalOpen(false)
    setEditingAdmin(null)
    form.resetFields()
  }

  const openCreateModal = () => {
    setEditingAdmin(null)
    form.setFieldsValue({
      status: 1,
      roleId: 1,
    })
    setModalOpen(true)
  }

  const openEditModal = (record: AdminItem) => {
    setEditingAdmin(record)
    form.setFieldsValue({
      username: record.username,
      realName: record.realName,
      email: record.email,
      phone: record.phone,
      status: record.status,
      roleId: record.roleId,
    })
    setModalOpen(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingAdmin) {
        updateAdmin({ id: editingAdmin.id, values })
      } else {
        createAdmin(values)
      }
    } catch {
      return
    }
  }

  const columns: ColumnsType<AdminItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '账号', dataIndex: 'username', key: 'username' },
    { title: '姓名', dataIndex: 'realName', key: 'realName' },
    { title: '邮箱', dataIndex: 'email', key: 'email' },
    { title: '手机号', dataIndex: 'phone', key: 'phone' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: number) => <Tag color={status === 1 ? 'green' : 'red'}>{status === 1 ? '启用' : '停用'}</Tag>,
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
      width: 180,
      render: (_, record) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} size="small" onClick={() => openEditModal(record)}>
            编辑
          </Button>
          <Popconfirm title="确认删除该管理员吗？" onConfirm={() => deleteAdmin(record.id)}>
            <Button type="link" danger icon={<DeleteOutlined />} size="small">
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card
        title="管理员管理"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
            新增管理员
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

      <Modal
        title={editingAdmin ? `编辑管理员 · ${editingAdmin.username}` : '新增管理员'}
        open={modalOpen}
        onCancel={closeModal}
        onOk={() => void handleSubmit()}
        okText={editingAdmin ? '保存' : '创建'}
        confirmLoading={creating || updating}
        destroyOnHidden
      >
        <Form form={form} layout="vertical" initialValues={{ status: 1, roleId: 1 }}>
          <Form.Item label="账号" name="username" rules={[{ required: true, message: '请输入账号' }]}>
            <Input placeholder="请输入管理员账号" disabled={editingAdmin?.id === 1} />
          </Form.Item>
          {!editingAdmin ? (
            <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
              <Input.Password placeholder="请输入初始密码" />
            </Form.Item>
          ) : null}
          <Form.Item label="姓名" name="realName" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="请输入真实姓名" />
          </Form.Item>
          <Form.Item label="邮箱" name="email">
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          <Form.Item label="手机号" name="phone">
            <Input placeholder="请输入手机号" />
          </Form.Item>
          <Form.Item label="角色 ID" name="roleId">
            <Select
              options={[
                { label: '超级管理员（1）', value: 1 },
                { label: '运营管理员（2）', value: 2 },
                { label: '审核员（3）', value: 3 },
              ]}
            />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true, message: '请选择状态' }]}>
            <Select
              options={[
                { label: '启用', value: 1 },
                { label: '停用', value: 0 },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
