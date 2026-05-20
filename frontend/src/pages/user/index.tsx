import { useState } from 'react'
import { Card, Table, Tag, Button, Space, Input, Drawer, Descriptions, Popconfirm, message, Modal, Form, Image, Divider, Alert } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fileApi, userApi, UserItem, PageParams } from '../../services/api'
import { formatDateTime } from '../../utils/format'
import { canAudit, canOperate } from '../../utils/permissions'
import { useAuthStore } from '../../store/useAuthStore'
import { SearchOutlined, EyeOutlined, CheckCircleOutlined, CloseCircleOutlined, IdcardOutlined } from '@ant-design/icons'

const identityStatusMap: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '未提交' },
  1: { color: 'processing', text: '审核中' },
  2: { color: 'success', text: '已通过' },
  3: { color: 'error', text: '已驳回' },
}

export default function UserList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const [keyword, setKeyword] = useState('')
  const [selectedUser, setSelectedUser] = useState<UserItem | null>(null)
  const [auditModal, setAuditModal] = useState<{ open: boolean; userId: number; action: 'approve' | 'reject' }>({ open: false, userId: 0, action: 'approve' })
  const [auditForm] = Form.useForm()
  const queryClient = useQueryClient()
  const roleId = useAuthStore((state) => state.adminInfo?.roleId)
  const auditAllowed = canAudit(roleId)
  const operateAllowed = canOperate(roleId)

  const queryParams = { ...params, keyword: keyword || undefined }

  const { data, isLoading } = useQuery({
    queryKey: ['users', queryParams],
    queryFn: () => userApi.list(queryParams),
  })

  const { data: signedIdentityImages } = useQuery({
    queryKey: ['identity-images', selectedUser?.id, selectedUser?.idCardFrontUrl, selectedUser?.idCardBackUrl],
    enabled: Boolean(selectedUser?.id && (selectedUser.idCardFrontUrl || selectedUser.idCardBackUrl)),
    staleTime: 5 * 60 * 1000,
    queryFn: async () => {
      const entries = await Promise.all(
        [
          ['front', selectedUser?.idCardFrontUrl],
          ['back', selectedUser?.idCardBackUrl],
        ].map(async ([side, url]) => {
          if (!url) return [side, ''] as const
          try {
            const response = await fileApi.getPrivateUrl(url)
            return [side, response.data.data.url] as const
          } catch {
            return [side, url] as const
          }
        })
      )
      return Object.fromEntries(entries) as { front?: string; back?: string }
    },
  })

  const { mutate: toggleStatus } = useMutation({
    mutationFn: ({ id, status }: { id: number; status: number }) => userApi.updateStatus(id, status),
    onSuccess: () => {
      message.success('状态已更新')
      void queryClient.invalidateQueries({ queryKey: ['users'] })
    },
    onError: () => message.error('操作失败'),
  })

  const { mutate: auditIdentity } = useMutation({
    mutationFn: ({ id, data }: { id: number; data: { status: number; reject?: string } }) => userApi.auditIdentity(id, data),
    onSuccess: (_, variables) => {
      message.success('审核完成')
      void queryClient.invalidateQueries({ queryKey: ['users'] })
      setAuditModal({ open: false, userId: 0, action: 'approve' })
      auditForm.resetFields()
      if (selectedUser) {
        setSelectedUser({
          ...selectedUser,
          identityStatus: variables.data.status,
          identityReject: variables.data.status === 3 ? variables.data.reject || '' : '',
        })
      }
    },
    onError: () => message.error('审核失败'),
  })

  const handleAudit = () => {
    auditForm.validateFields().then((values) => {
      const status = auditModal.action === 'approve' ? 2 : 3
      auditIdentity({
        id: auditModal.userId,
        data: { status, reject: auditModal.action === 'reject' ? values.reject : undefined },
      })
    })
  }

  const hasIdentityInfo = (user: UserItem) => {
    return user.identityStatus !== undefined && user.identityStatus > 0
  }

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
      title: '实名状态', dataIndex: 'identityStatus', key: 'identityStatus', width: 90,
      render: (status?: number) => {
        const meta = identityStatusMap[status ?? 0]
        return <Tag color={meta?.color || 'default'}>{meta?.text || '未知'}</Tag>
      },
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
      title: '操作', key: 'actions', width: 200,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setSelectedUser(record)}>详情</Button>
          {auditAllowed && record.identityStatus === 1 && (
            <Button type="link" size="small" icon={<IdcardOutlined />} style={{ color: '#f59e0b' }} onClick={() => setSelectedUser(record)}>
              审核
            </Button>
          )}
          {operateAllowed && (
            <Popconfirm
              title={`确认${record.status === 1 ? '停用' : '启用'}该用户？`}
              onConfirm={() => toggleStatus({ id: record.id, status: record.status === 1 ? 0 : 1 })}
            >
              <Button type="link" size="small" danger={record.status === 1}>
                {record.status === 1 ? '停用' : '启用'}
              </Button>
            </Popconfirm>
          )}
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
          scroll={{ x: 1100 }}
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
        width={520}
        open={selectedUser !== null}
        onClose={() => setSelectedUser(null)}
        destroyOnClose
      >
        {selectedUser && (
          <>
            <Descriptions bordered column={1} size="small">
              <Descriptions.Item label="ID">{selectedUser.id}</Descriptions.Item>
              <Descriptions.Item label="账号">{selectedUser.username}</Descriptions.Item>
              <Descriptions.Item label="姓名">{selectedUser.realName || '-'}</Descriptions.Item>
              <Descriptions.Item label="手机号">{selectedUser.phone || '-'}</Descriptions.Item>
              <Descriptions.Item label="邮箱">{selectedUser.email || '-'}</Descriptions.Item>
              <Descriptions.Item label="企业 ID">{selectedUser.enterpriseId || '未绑定'}</Descriptions.Item>
              <Descriptions.Item label="账号状态">
                <Tag color={selectedUser.status === 1 ? 'green' : 'red'}>{selectedUser.status === 1 ? '启用' : '停用'}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="注册时间">{formatDateTime(selectedUser.createdAt)}</Descriptions.Item>
            </Descriptions>

            <Divider orientation="left" orientationMargin={0}>
              <IdcardOutlined className="mr-1" />
              实名认证信息
            </Divider>

            {hasIdentityInfo(selectedUser) ? (
              <>
                {auditAllowed && selectedUser.identityStatus === 1 && (
                  <Alert
                    type="warning"
                    message="该用户已提交实名认证，等待审核"
                    showIcon
                    className="mb-4"
                  />
                )}
                {selectedUser.identityStatus === 2 && (
                  <Alert
                    type="success"
                    message="该用户实名认证已通过"
                    showIcon
                    className="mb-4"
                  />
                )}
                {selectedUser.identityStatus === 3 && (
                  <Alert
                    type="error"
                    message={`实名认证已驳回：${selectedUser.identityReject || '无驳回原因'}`}
                    showIcon
                    className="mb-4"
                  />
                )}

                <Descriptions bordered column={1} size="small">
                  <Descriptions.Item label="认证状态">
                    <Tag color={identityStatusMap[selectedUser.identityStatus ?? 0]?.color}>
                      {identityStatusMap[selectedUser.identityStatus ?? 0]?.text || '未知'}
                    </Tag>
                  </Descriptions.Item>
                  {selectedUser.realName && selectedUser.realName !== selectedUser.username && (
                    <Descriptions.Item label="真实姓名">{selectedUser.realName}</Descriptions.Item>
                  )}
                  {selectedUser.idCard && (
                    <Descriptions.Item label="身份证号">{selectedUser.idCard}</Descriptions.Item>
                  )}
                  {selectedUser.identityReject && selectedUser.identityStatus === 3 && (
                    <Descriptions.Item label="驳回原因">
                      <span className="text-red-500">{selectedUser.identityReject}</span>
                    </Descriptions.Item>
                  )}
                </Descriptions>

                {(selectedUser.idCardFrontUrl || selectedUser.idCardBackUrl) && (
                  <div className="mt-4">
                    <div className="text-gray-500 text-sm mb-2">证件照片</div>
                    <Space size="middle">
                      {selectedUser.idCardFrontUrl && (
                        <div className="text-center">
                          <Image
                            src={signedIdentityImages?.front || selectedUser.idCardFrontUrl}
                            alt="身份证正面"
                            width={200}
                            className="rounded border object-cover"
                            fallback="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjEyNSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMjAwIiBoZWlnaHQ9IjEyNSIgZmlsbD0iI2YwZjBmMCIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBkb21pbmFudC1iYXNlbGluZT0ibWlkZGxlIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBmaWxsPSIjY2NjIiBmb250LXNpemU9IjE0Ij7lm77niYc0MDQ8L3RleHQ+PC9zdmc+"
                          />
                          <div className="text-gray-400 text-xs mt-1">人像面</div>
                        </div>
                      )}
                      {selectedUser.idCardBackUrl && (
                        <div className="text-center">
                          <Image
                            src={signedIdentityImages?.back || selectedUser.idCardBackUrl}
                            alt="身份证背面"
                            width={200}
                            className="rounded border object-cover"
                            fallback="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjEyNSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMjAwIiBoZWlnaHQ9IjEyNSIgZmlsbD0iI2YwZjBmMCIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBkb21pbmFudC1iYXNlbGluZT0ibWlkZGxlIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBmaWxsPSIjY2NjIiBmb250LXNpemU9IjE0Ij7lm77niYc0MDQ8L3RleHQ+PC9zdmc+"
                          />
                          <div className="text-gray-400 text-xs mt-1">国徽面</div>
                        </div>
                      )}
                    </Space>
                  </div>
                )}

                {selectedUser.identityStatus === 1 && (
                  <div className="mt-4 flex gap-2">
                    <Button
                      type="primary"
                      icon={<CheckCircleOutlined />}
                      onClick={() => auditIdentity({ id: selectedUser.id, data: { status: 2 } })}
                    >
                      审核通过
                    </Button>
                    <Button
                      danger
                      icon={<CloseCircleOutlined />}
                      onClick={() => setAuditModal({ open: true, userId: selectedUser.id, action: 'reject' })}
                    >
                      驳回
                    </Button>
                  </div>
                )}
              </>
            ) : (
              <div className="text-center py-6 text-gray-400">
                <IdcardOutlined className="text-3xl mb-2 block" />
                该用户尚未提交实名认证
              </div>
            )}
          </>
        )}
      </Drawer>

      <Modal
        title="驳回实名认证"
        open={auditModal.open}
        onOk={handleAudit}
        onCancel={() => {
          setAuditModal({ open: false, userId: 0, action: 'approve' })
          auditForm.resetFields()
        }}
        okText="确认驳回"
        okButtonProps={{ danger: true }}
      >
        <Form form={auditForm} layout="vertical">
          <Form.Item
            name="reject"
            label="驳回原因"
            rules={[{ required: true, message: '请输入驳回原因' }]}
          >
            <Input.TextArea rows={3} placeholder="请输入驳回原因，用户将看到此信息" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
