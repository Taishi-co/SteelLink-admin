import { useState, type Key } from 'react'
import { Card, Table, Button, Space, Tag, Popconfirm, message, Modal, Form, Input, Select, InputNumber, Tree } from 'antd'
import type { TreeDataNode } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, SafetyOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { roleApi, menuApi, RoleItem, MenuItem, PageParams } from '../../../services/api'
import { formatDateTime } from '../../../utils/format'

interface RoleFormValues {
  name: string
  code: string
  description?: string
  status: number
  sort: number
}

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

function menuToTreeData(menus: MenuItem[]): TreeDataNode[] {
  return menus.map(m => ({
    title: m.name,
    key: m.id,
    children: m.children ? menuToTreeData(m.children) : [],
  }))
}

export default function RoleList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const [modalOpen, setModalOpen] = useState(false)
  const [permModalOpen, setPermModalOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<RoleItem | null>(null)
  const [permRoleId, setPermRoleId] = useState<number | null>(null)
  const [checkedMenuIds, setCheckedMenuIds] = useState<Key[]>([])
  const [form] = Form.useForm<RoleFormValues>()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['roles', params],
    queryFn: () => roleApi.list(params),
  })

  const { data: menuData } = useQuery({
    queryKey: ['menus'],
    queryFn: () => menuApi.list(),
  })

  const menuTree = menuData?.data.data.list ? buildMenuTree(menuData.data.data.list) : []
  const treeData = menuToTreeData(menuTree)

  const { mutate: createRole, isPending: creating } = useMutation({
    mutationFn: (values: RoleFormValues) => roleApi.create(values),
    onSuccess: () => {
      message.success('角色已创建')
      closeModal()
      void queryClient.invalidateQueries({ queryKey: ['roles'] })
    },
  })

  const { mutate: updateRole, isPending: updating } = useMutation({
    mutationFn: ({ id, values }: { id: number; values: RoleFormValues }) => roleApi.update(id, values),
    onSuccess: () => {
      message.success('角色已更新')
      closeModal()
      void queryClient.invalidateQueries({ queryKey: ['roles'] })
    },
  })

  const { mutate: deleteRole } = useMutation({
    mutationFn: roleApi.delete,
    onSuccess: () => {
      message.success('删除成功')
      void queryClient.invalidateQueries({ queryKey: ['roles'] })
    },
  })

  const { mutate: saveMenus, isPending: savingMenus } = useMutation({
    mutationFn: ({ id, menuIds }: { id: number; menuIds: number[] }) => roleApi.updateMenus(id, menuIds),
    onSuccess: () => {
      message.success('权限已保存')
      setPermModalOpen(false)
    },
  })

  const closeModal = () => {
    setModalOpen(false)
    setEditingRole(null)
    form.resetFields()
  }

  const openCreate = () => {
    setEditingRole(null)
    form.setFieldsValue({ status: 1, sort: 0 })
    setModalOpen(true)
  }

  const openEdit = (record: RoleItem) => {
    setEditingRole(record)
    form.setFieldsValue({
      name: record.name,
      code: record.code,
      description: record.description,
      status: record.status,
      sort: record.sort,
    })
    setModalOpen(true)
  }

  const openPerm = async (record: RoleItem) => {
    setPermRoleId(record.id)
    try {
      const res = await roleApi.getMenus(record.id)
      setCheckedMenuIds(res.data.data || [])
    } catch {
      setCheckedMenuIds([])
    }
    setPermModalOpen(true)
  }

  const handleSubmit = async () => {
    const values = await form.validateFields()
    if (editingRole) {
      updateRole({ id: editingRole.id, values })
    } else {
      createRole(values)
    }
  }

  const columns: ColumnsType<RoleItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '角色名称', dataIndex: 'name', key: 'name' },
    { title: '角色编码', dataIndex: 'code', key: 'code' },
    { title: '说明', dataIndex: 'description', key: 'description' },
    {
      title: '状态', dataIndex: 'status', key: 'status',
      render: (status: number) => <Tag color={status === 1 ? 'green' : 'red'}>{status === 1 ? '启用' : '停用'}</Tag>,
    },
    { title: '排序', dataIndex: 'sort', key: 'sort', width: 80 },
    {
      title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 180,
      render: (value: string) => formatDateTime(value),
    },
    {
      title: '操作', key: 'actions', width: 220,
      render: (_, record) => (
        <Space>
          <Button type="link" icon={<SafetyOutlined />} size="small" onClick={() => openPerm(record)}>权限</Button>
          <Button type="link" icon={<EditOutlined />} size="small" onClick={() => openEdit(record)}>编辑</Button>
          <Popconfirm title="确认删除该角色吗？" onConfirm={() => deleteRole(record.id)}>
            <Button type="link" danger icon={<DeleteOutlined />} size="small">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card
        title="角色列表"
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增角色</Button>}
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
        title={editingRole ? `编辑角色 · ${editingRole.name}` : '新增角色'}
        open={modalOpen}
        onCancel={closeModal}
        onOk={() => void handleSubmit()}
        okText={editingRole ? '保存' : '创建'}
        confirmLoading={creating || updating}
        destroyOnHidden
      >
        <Form form={form} layout="vertical">
          <Form.Item label="角色名称" name="name" rules={[{ required: true, message: '请输入角色名称' }]}>
            <Input placeholder="如：运营管理员" />
          </Form.Item>
          <Form.Item label="角色编码" name="code" rules={[{ required: true, message: '请输入角色编码' }]}>
            <Input placeholder="如：operator" />
          </Form.Item>
          <Form.Item label="说明" name="description">
            <Input.TextArea rows={2} placeholder="角色职责描述" />
          </Form.Item>
          <Form.Item label="排序" name="sort">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true }]}>
            <Select options={[{ label: '启用', value: 1 }, { label: '停用', value: 0 }]} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="分配菜单权限"
        open={permModalOpen}
        onCancel={() => setPermModalOpen(false)}
        onOk={() => permRoleId && saveMenus({ id: permRoleId, menuIds: checkedMenuIds.map((key) => Number(key)) })}
        okText="保存"
        confirmLoading={savingMenus}
        destroyOnHidden
      >
        <Tree
          checkable
          defaultExpandAll
          treeData={treeData}
          checkedKeys={checkedMenuIds}
          onCheck={(checked) => setCheckedMenuIds(Array.isArray(checked) ? checked : checked.checked)}
        />
      </Modal>
    </>
  )
}
