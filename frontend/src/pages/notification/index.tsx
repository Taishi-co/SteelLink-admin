import { Button, Card, Form, Input, Select, Space, Switch, Typography, message } from 'antd'
import { SendOutlined } from '@ant-design/icons'
import { useMutation } from '@tanstack/react-query'
import { notificationApi, type SendNotificationPayload } from '../../services/api'

const { TextArea } = Input
const { Paragraph } = Typography

interface NotificationFormValues {
  sendToAll?: boolean
  recipientUserIds?: string
  recipientEnterpriseIds?: string
  recipientRoles?: string[]
  title: string
  content: string
  linkUrl?: string
}

function parseIDs(value?: string): number[] {
  return (value || '')
    .split(/[,\s，]+/)
    .map((item) => Number(item.trim()))
    .filter((item) => Number.isInteger(item) && item > 0)
}

export default function NotificationPage() {
  const [form] = Form.useForm<NotificationFormValues>()
  const sendToAll = Form.useWatch('sendToAll', form)

  const mutation = useMutation({
    mutationFn: (payload: SendNotificationPayload) => notificationApi.send(payload),
    onSuccess: (response) => {
      message.success(`站内信已发送，生成 ${response.data.data.created} 条消息`)
      form.resetFields()
    },
  })

  const handleFinish = (values: NotificationFormValues) => {
    const payload: SendNotificationPayload = {
      send_to_all: Boolean(values.sendToAll),
      recipient_user_ids: parseIDs(values.recipientUserIds),
      recipient_enterprise_ids: parseIDs(values.recipientEnterpriseIds),
      recipient_roles: values.recipientRoles || [],
      type: 'admin',
      title: values.title,
      content: values.content,
      link_url: values.linkUrl,
    }
    mutation.mutate(payload)
  }

  return (
    <Card title="发送站内信">
      <Paragraph type="secondary">
        管理员通知会进入用户买家端或卖家端的站内信收件箱，可按用户 ID、企业 ID、用户角色或全量发送。
      </Paragraph>
      <Form form={form} layout="vertical" initialValues={{ sendToAll: false }} onFinish={handleFinish} style={{ maxWidth: 760 }}>
        <Form.Item name="sendToAll" label="发送范围" valuePropName="checked">
          <Switch checkedChildren="全量发送" unCheckedChildren="指定接收人" />
        </Form.Item>

        {!sendToAll && (
          <>
            <Form.Item name="recipientUserIds" label="用户 ID">
              <Input placeholder="多个 ID 用逗号或空格分隔，例如 1001,1002" />
            </Form.Item>
            <Form.Item name="recipientEnterpriseIds" label="企业 ID">
              <Input placeholder="多个企业 ID 用逗号或空格分隔，例如 2001,2002" />
            </Form.Item>
            <Form.Item name="recipientRoles" label="用户角色">
              <Select
                mode="multiple"
                allowClear
                options={[
                  { label: '个人用户', value: 'personal' },
                  { label: '企业用户', value: 'enterprise' },
                  { label: '运营人员', value: 'operator' },
                ]}
              />
            </Form.Item>
          </>
        )}

        <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }, { max: 120, message: '标题最多 120 字' }]}>
          <Input placeholder="例如：平台维护通知" />
        </Form.Item>
        <Form.Item name="content" label="内容" rules={[{ required: true, message: '请输入内容' }, { max: 5000, message: '内容最多 5000 字' }]}>
          <TextArea rows={6} placeholder="填写通知正文" />
        </Form.Item>
        <Form.Item name="linkUrl" label="站内链接">
          <Input placeholder="可选，例如 /dashboard/orders 或 /seller/products" />
        </Form.Item>

        <Space>
          <Button type="primary" htmlType="submit" icon={<SendOutlined />} loading={mutation.isPending}>
            发送
          </Button>
          <Button onClick={() => form.resetFields()}>重置</Button>
        </Space>
      </Form>
    </Card>
  )
}
