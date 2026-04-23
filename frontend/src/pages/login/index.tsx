import { useState } from 'react'
import { Form, Input, Button, Card, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { authApi } from '../../services/api'
import { useAuthStore } from '../../store/useAuthStore'
import styles from './index.module.css'

interface LoginForm {
  username: string
  password: string
}

export default function Login() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const { setToken, setAdminInfo } = useAuthStore()

  const { mutate: login } = useMutation({
    mutationFn: authApi.login,
    onSuccess: (response) => {
      const { token, adminInfo } = response.data.data
      setToken(token)
      setAdminInfo(adminInfo)
      message.success('登录成功')
      navigate('/dashboard')
    },
    onError: () => {
      setLoading(false)
    },
  })

  const handleSubmit = (values: LoginForm) => {
    setLoading(true)
    login(values)
  }

  return (
    <div className={styles.container}>
      <section className={styles.hero}>
        <div className={styles.eyebrow}>SteelLink 运营平台</div>
        <h1 className={styles.title}>钢材交易的管理中枢，统一看板一屏掌控。</h1>
        <p className={styles.subtitle}>
          面向平台运营、商品审核、企业管理与系统配置的后台工作台。
          风格延续买卖家端的工业蓝体系，重点突出状态、效率与决策感。
        </p>
        <div className={styles.metrics}>
          <div className={styles.metricCard}>
            <span className={styles.metricValue}>用户</span>
            <span className={styles.metricLabel}>统一管理买家、卖家与管理员</span>
          </div>
          <div className={styles.metricCard}>
            <span className={styles.metricValue}>商品</span>
            <span className={styles.metricLabel}>商品审核、类目与库存概览</span>
          </div>
          <div className={styles.metricCard}>
            <span className={styles.metricValue}>风控</span>
            <span className={styles.metricLabel}>待审事项、权限与运营追踪</span>
          </div>
        </div>
      </section>
      <Card className={styles.card} title="管理端登录">
        <p className={styles.helper}>请输入管理员账号与密码，进入 SteelLink 平台运营后台。</p>
        <Form
          name="login"
          initialValues={{ remember: true }}
          onFinish={handleSubmit}
          size="large"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入账号' }]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="请输入账号"
              autoComplete="username"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请输入密码"
              autoComplete="current-password"
            />
          </Form.Item>

          <Form.Item>
            <Button className={styles.submit} type="primary" htmlType="submit" loading={loading} block>
              进入管理后台
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
