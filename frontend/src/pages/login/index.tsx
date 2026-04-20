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
      message.success('Login successful')
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
      <Card className={styles.card} title="SteelLink Admin">
        <Form
          name="login"
          initialValues={{ remember: true }}
          onFinish={handleSubmit}
          size="large"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: 'Please input your username' }]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="Username"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: 'Please input your password' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="Password"
            />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              Login
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
