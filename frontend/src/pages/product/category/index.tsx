import { useState } from 'react'
import { Card, Table, Tag, Button, Space, Popconfirm, message, Modal, Form, Input, InputNumber, Select } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { categoryApi, CategoryItem } from '../../../services/api'

interface CategoryFormValues {
  name: string
  displayName: string
  code: string
  icon: string
  imageUrl: string
  sort: number
  status: number
}

export default function CategoryList() {
  const [modalOpen, setModalOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<CategoryItem | null>(null)
  const [form] = Form.useForm<CategoryFormValues>()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list(),
  })

  const { mutate: createCategory, isPending: creating } = useMutation({
    mutationFn: (values: CategoryFormValues) => categoryApi.create(values),
    onSuccess: () => { message.success('类目已创建'); closeModal(); void queryClient.invalidateQueries({ queryKey: ['categories'] }) },
  })

  const { mutate: updateCategory, isPending: updating } = useMutation({
    mutationFn: ({ id, values }: { id: number; values: CategoryFormValues }) => categoryApi.update(id, values),
    onSuccess: () => { message.success('类目已更新'); closeModal(); void queryClient.invalidateQueries({ queryKey: ['categories'] }) },
  })

  const { mutate: deleteCategory } = useMutation({
    mutationFn: categoryApi.delete,
    onSuccess: () => { message.success('删除成功'); void queryClient.invalidateQueries({ queryKey: ['categories'] }) },
  })

  const closeModal = () => { setModalOpen(false); setEditingItem(null); form.resetFields() }

  const openCreate = () => { setEditingItem(null); form.setFieldsValue({ displayName: '', code: '', icon: '', imageUrl: '', status: 1, sort: 0 }); setModalOpen(true) }

  const openEdit = (record: CategoryItem) => {
    setEditingItem(record)
    form.setFieldsValue({
      name: record.name,
      displayName: record.displayName || record.name,
      code: record.code || '',
      icon: record.icon || '',
      imageUrl: record.imageUrl || '',
      sort: record.sort,
      status: record.status,
    })
    setModalOpen(true)
  }

  const handleSubmit = async () => {
    const values = await form.validateFields()
    if (editingItem) { updateCategory({ id: editingItem.id, values }) } else { createCategory(values) }
  }

  const columns: ColumnsType<CategoryItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: '业务品类', dataIndex: 'name', key: 'name' },
    { title: '展示名', dataIndex: 'displayName', key: 'displayName', render: (_, record) => record.displayName || record.name },
    { title: '编码', dataIndex: 'code', key: 'code' },
    {
      title: '导航图片', dataIndex: 'imageUrl', key: 'imageUrl', width: 120,
      render: (imageUrl: string | undefined) => imageUrl ? <img src={imageUrl} alt="品类导航" style={{ width: 56, height: 36, objectFit: 'cover', borderRadius: 4 }} /> : '-',
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80,
      render: (status: number) => <Tag color={status === 1 ? 'green' : 'red'}>{status === 1 ? '启用' : '停用'}</Tag>,
    },
    { title: '排序', dataIndex: 'sort', key: 'sort', width: 60 },
    {
      title: '操作', key: 'actions', width: 150,
      render: (_, record) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} size="small" onClick={() => openEdit(record)}>编辑</Button>
          <Popconfirm title="确认删除该类目？" onConfirm={() => deleteCategory(record.id)}>
            <Button type="link" danger icon={<DeleteOutlined />} size="small">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card title="类目配置" extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增类目</Button>}>
        <Table
          columns={columns}
          dataSource={data?.data.data.list || []}
          rowKey="id"
          loading={isLoading}
          pagination={false}
        />
      </Card>

      <Modal
        title={editingItem ? `编辑类目 · ${editingItem.name}` : '新增类目'}
        open={modalOpen}
        onCancel={closeModal}
        onOk={() => void handleSubmit()}
        okText={editingItem ? '保存' : '创建'}
        confirmLoading={creating || updating}
        destroyOnHidden
      >
        <Form form={form} layout="vertical">
          <Form.Item label="业务品类" name="name" rules={[{ required: true, message: '请输入业务品类' }]}>
            <Input placeholder="如：热轧卷板" />
          </Form.Item>
          <Form.Item label="展示名" name="displayName">
            <Input placeholder="如：热轧板卷专区；为空时使用业务品类" />
          </Form.Item>
          <Form.Item label="编码" name="code">
            <Input placeholder="如：hot_rolled_coil" />
          </Form.Item>
          <Form.Item label="导航图标" name="icon">
            <Input placeholder="如：钢；无图片时作为文本图标展示" />
          </Form.Item>
          <Form.Item label="导航图片 URL" name="imageUrl">
            <Input placeholder="https://..." />
          </Form.Item>
          <Form.Item label="排序" name="sort">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="状态" name="status">
            <Select options={[{ label: '启用', value: 1 }, { label: '停用', value: 0 }]} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
