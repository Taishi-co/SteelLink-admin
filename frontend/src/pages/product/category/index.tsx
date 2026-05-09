import { useState } from 'react'
import { Card, Table, Tag, Button, Space, Popconfirm, message, Modal, Form, Input, InputNumber, Select, Upload, Tooltip } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, UploadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { UploadProps } from 'antd'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { categoryApi, CategoryItem, fileApi } from '../../../services/api'

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
    onError: () => message.error('删除失败，请确认该类目下没有关联商品'),
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

  const handleImageUpload: UploadProps['customRequest'] = async (options) => {
    const file = options.file as File
    try {
      const response = await fileApi.upload(file, 'product_category_nav')
      const imageUrl = response.data.data.url
      form.setFieldValue('imageUrl', imageUrl)
      options.onSuccess?.(response.data, file)
      message.success('导航图片已上传')
    } catch (error) {
      options.onError?.(error as Error)
    }
  }

  const beforeImageUpload: UploadProps['beforeUpload'] = (file) => {
    if (!file.type.startsWith('image/')) {
      message.error('只能上传图片文件')
      return Upload.LIST_IGNORE
    }
    if (file.size > 5 * 1024 * 1024) {
      message.error('图片不能超过 5MB')
      return Upload.LIST_IGNORE
    }
    return true
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
    {
      title: '商品数量',
      dataIndex: 'productCount',
      key: 'productCount',
      width: 100,
      render: (count: number | undefined) => {
        const value = count || 0
        return <Tag color={value > 0 ? 'blue' : 'default'}>{value} 个</Tag>
      },
    },
    { title: '排序', dataIndex: 'sort', key: 'sort', width: 60 },
    {
      title: '操作', key: 'actions', width: 150,
      render: (_, record) => {
        const productCount = record.productCount || 0
        const deleteButton = (
          <Button type="link" danger icon={<DeleteOutlined />} size="small" disabled={productCount > 0}>删除</Button>
        )
        return (
          <Space>
            <Button type="link" icon={<EditOutlined />} size="small" onClick={() => openEdit(record)}>编辑</Button>
            {productCount > 0 ? (
              <Tooltip title={`该类目下还有 ${productCount} 个商品，不能删除`}>
                <span>{deleteButton}</span>
              </Tooltip>
            ) : (
              <Popconfirm
                title="确认删除该类目？"
                description="该类目当前没有关联商品，删除后不可恢复。"
                onConfirm={() => deleteCategory(record.id)}
              >
                {deleteButton}
              </Popconfirm>
            )}
          </Space>
        )
      },
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
          <Form.Item label="导航图片" extra="可直接上传图片，也可粘贴已有图片 URL。">
            <Space direction="vertical" style={{ width: '100%' }} size={8}>
              <Space.Compact style={{ width: '100%' }}>
                <Upload
                  accept="image/*"
                  maxCount={1}
                  showUploadList={false}
                  beforeUpload={beforeImageUpload}
                  customRequest={handleImageUpload}
                >
                  <Button icon={<UploadOutlined />}>上传</Button>
                </Upload>
                <Form.Item name="imageUrl" noStyle>
                  <Input placeholder="https://..." />
                </Form.Item>
              </Space.Compact>
              <Form.Item shouldUpdate noStyle>
                {({ getFieldValue }) => {
                  const imageUrl = getFieldValue('imageUrl') as string | undefined
                  return imageUrl ? (
                    <img
                      src={imageUrl}
                      alt="导航图片预览"
                      style={{ width: 112, height: 72, objectFit: 'cover', borderRadius: 4, border: '1px solid #f0f0f0' }}
                    />
                  ) : null
                }}
              </Form.Item>
            </Space>
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
