import { useState } from 'react'
import { Card, Table, Button, Space, Tag, Popconfirm, message, Modal, Form, Input, Select, InputNumber, TreeSelect } from 'antd'
import type { TreeSelectProps } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { menuApi, MenuItem } from '../../../services/api'

interface MenuFormValues {
  parentId: number
  name: string
  path: string
  component: string
  icon: string
  type: number
  permission: string
  status: number
  sort: number
}

type MenuTreeSelectNode = NonNullable<TreeSelectProps['treeData']>[number]

function buildMenuTree(menus: MenuItem[]): MenuItem[] {
  const map = new Map<number, MenuItem & { children: MenuItem[] }>()
  const roots: MenuItem[] = []
  menus.forEach(m => map.set(m.id, { ...m, children: [] }))
  menus.forEach(m => {
    if (m.parentId === 0) {
      roots.push(map.get(m.id)!)
    } else {
      map.get(m.parentId)?.children.push(map.get(m.id)!)
    }
  })
  return roots
}

function menuToTreeSelectData(menus: MenuItem[]): MenuTreeSelectNode[] {
  return menus.map(m => ({
    key: m.id,
    title: m.name,
    value: m.id,
    children: m.children ? menuToTreeSelectData(m.children) : [],
  }))
}

const iconOptions = [
  'DashboardOutlined', 'SettingOutlined', 'UserOutlined', 'TeamOutlined',
  'SafetyOutlined', 'MenuOutlined', 'BankOutlined', 'ShoppingOutlined',
  'AppstoreOutlined', 'ShopOutlined', 'SolutionOutlined',
].map(v => ({ label: v, value: v }))

export default function MenuList() {
  const [modalOpen, setModalOpen] = useState(false)
  const [editingMenu, setEditingMenu] = useState<MenuItem | null>(null)
  const [form] = Form.useForm<MenuFormValues>()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['menus'],
    queryFn: () => menuApi.list(),
  })

  const flatMenus = data?.data.data.list || []
  const menuTree = buildMenuTree(flatMenus)
  const treeSelectData: MenuTreeSelectNode[] = [{ key: 0, title: '顶级菜单', value: 0, children: menuToTreeSelectData(menuTree) }]

  const { mutate: createMenu, isPending: creating } = useMutation({
    mutationFn: (values: MenuFormValues) => menuApi.create(values),
    onSuccess: () => {
      message.success('菜单已创建')
      closeModal()
      void queryClient.invalidateQueries({ queryKey: ['menus'] })
    },
  })

  const { mutate: updateMenu, isPending: updating } = useMutation({
    mutationFn: ({ id, values }: { id: number; values: MenuFormValues }) => menuApi.update(id, values),
    onSuccess: () => {
      message.success('菜单已更新')
      closeModal()
      void queryClient.invalidateQueries({ queryKey: ['menus'] })
    },
  })

  const { mutate: deleteMenu } = useMutation({
    mutationFn: menuApi.delete,
    onSuccess: () => {
      message.success('删除成功')
      void queryClient.invalidateQueries({ queryKey: ['menus'] })
    },
  })

  const closeModal = () => {
    setModalOpen(false)
    setEditingMenu(null)
    form.resetFields()
  }

  const openCreate = () => {
    setEditingMenu(null)
    form.setFieldsValue({ parentId: 0, type: 1, status: 1, sort: 0 })
    setModalOpen(true)
  }

  const openEdit = (record: MenuItem) => {
    setEditingMenu(record)
    form.setFieldsValue({
      parentId: record.parentId,
      name: record.name,
      path: record.path,
      component: record.component || '',
      icon: record.icon || '',
      type: record.type,
      permission: record.permission || '',
      status: record.status,
      sort: record.sort,
    })
    setModalOpen(true)
  }

  const handleSubmit = async () => {
    const values = await form.validateFields()
    if (editingMenu) {
      updateMenu({ id: editingMenu.id, values })
    } else {
      createMenu(values)
    }
  }

  const columns: ColumnsType<MenuItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: '菜单名称', dataIndex: 'name', key: 'name' },
    { title: '路径', dataIndex: 'path', key: 'path' },
    { title: '图标', dataIndex: 'icon', key: 'icon', width: 160 },
    { title: '权限标识', dataIndex: 'permission', key: 'permission' },
    {
      title: '类型', dataIndex: 'type', key: 'type', width: 80,
      render: (type: number) => <Tag color={type === 1 ? 'blue' : 'orange'}>{type === 1 ? '菜单' : '按钮'}</Tag>,
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80,
      render: (status: number) => <Tag color={status === 1 ? 'green' : 'red'}>{status === 1 ? '显示' : '隐藏'}</Tag>,
    },
    { title: '排序', dataIndex: 'sort', key: 'sort', width: 60 },
    {
      title: '操作', key: 'actions', width: 150,
      render: (_, record) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} size="small" onClick={() => openEdit(record)}>编辑</Button>
          <Popconfirm title="确认删除该菜单吗？" onConfirm={() => deleteMenu(record.id)}>
            <Button type="link" danger icon={<DeleteOutlined />} size="small">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card
        title="菜单列表"
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增菜单</Button>}
      >
        <Table
          columns={columns}
          dataSource={menuTree}
          rowKey="id"
          loading={isLoading}
          pagination={false}
          expandable={{ childrenColumnName: 'children' }}
        />
      </Card>

      <Modal
        title={editingMenu ? `编辑菜单 · ${editingMenu.name}` : '新增菜单'}
        open={modalOpen}
        onCancel={closeModal}
        onOk={() => void handleSubmit()}
        okText={editingMenu ? '保存' : '创建'}
        confirmLoading={creating || updating}
        destroyOnHidden
        width={560}
      >
        <Form form={form} layout="vertical">
          <Form.Item label="上级菜单" name="parentId">
            <TreeSelect treeData={treeSelectData} treeDefaultExpandAll placeholder="选择上级菜单" />
          </Form.Item>
          <Form.Item label="菜单名称" name="name" rules={[{ required: true, message: '请输入菜单名称' }]}>
            <Input placeholder="如：用户管理" />
          </Form.Item>
          <Form.Item label="路由路径" name="path">
            <Input placeholder="如：/user" />
          </Form.Item>
          <Form.Item label="组件路径" name="component">
            <Input placeholder="如：user/index" />
          </Form.Item>
          <Form.Item label="图标" name="icon">
            <Select options={iconOptions} allowClear placeholder="选择图标" />
          </Form.Item>
          <Form.Item label="类型" name="type">
            <Select options={[{ label: '菜单', value: 1 }, { label: '按钮', value: 2 }]} />
          </Form.Item>
          <Form.Item label="权限标识" name="permission">
            <Input placeholder="如：system:admin:list" />
          </Form.Item>
          <Form.Item label="排序" name="sort">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="状态" name="status">
            <Select options={[{ label: '显示', value: 1 }, { label: '隐藏', value: 0 }]} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
