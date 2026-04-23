import { useState } from 'react'
import { Layout, Menu, Avatar, Dropdown, Button, theme } from 'antd'
import {
  DashboardOutlined,
  SettingOutlined,
  UserOutlined,
  BankOutlined,
  ShoppingOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../store/useAuthStore'
import styles from './AdminLayout.module.css'

const { Header, Sider, Content } = Layout

const menuItems = [
  {
    key: '/dashboard',
    icon: <DashboardOutlined />,
    label: 'Dashboard',
  },
  {
    key: '/system',
    icon: <SettingOutlined />,
    label: 'System',
    children: [
      { key: '/system/admin', label: 'Admin Management' },
      { key: '/system/role', label: 'Role Management' },
      { key: '/system/menu', label: 'Menu Management' },
    ],
  },
  {
    key: '/user',
    icon: <UserOutlined />,
    label: 'User Management',
  },
  {
    key: '/enterprise',
    icon: <BankOutlined />,
    label: 'Enterprise Management',
  },
  {
    key: '/product',
    icon: <ShoppingOutlined />,
    label: 'Product Management',
    children: [
      { key: '/product', label: 'Product List' },
      { key: '/product/order', label: 'Order List' },
      { key: '/product/category', label: 'Category Management' },
    ],
  },
]

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { adminInfo, logout } = useAuthStore()
  const { token: { colorBgContainer, borderRadiusLG } } = theme.useToken()

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
  }

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const userMenuItems = [
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Logout',
      onClick: handleLogout,
    },
  ]

  const getSelectedKeys = () => {
    const path = location.pathname
    return [path]
  }

  const getOpenKeys = () => {
    const path = location.pathname
    const parts = path.split('/').filter(Boolean)
    if (parts.length > 1) {
      return [`/${parts[0]}`]
    }
    return []
  }

  return (
    <Layout className={styles.layout}>
      <Sider trigger={null} collapsible collapsed={collapsed} className={styles.sider}>
        <div className={styles.logo}>
          {collapsed ? 'SL' : 'SteelLink Admin'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <Layout>
        <Header className={styles.header} style={{ background: colorBgContainer }}>
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            className={styles.trigger}
          />
          <div className={styles.userInfo}>
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div className={styles.userAvatar}>
                <Avatar icon={<UserOutlined />} />
                <span className={styles.userName}>{adminInfo?.realName || 'Admin'}</span>
              </div>
            </Dropdown>
          </div>
        </Header>
        <Content
          className={styles.content}
          style={{
            background: colorBgContainer,
            borderRadius: borderRadiusLG,
          }}
        >
          {children || <Outlet />}
        </Content>
      </Layout>
    </Layout>
  )
}
