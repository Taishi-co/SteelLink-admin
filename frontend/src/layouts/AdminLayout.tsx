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
  TeamOutlined,
  SafetyOutlined,
  MenuOutlined,
  AppstoreOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../store/useAuthStore'
import styles from './AdminLayout.module.css'

const { Header, Sider, Content } = Layout

const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '数据概览' },
  {
    key: '/system',
    icon: <SettingOutlined />,
    label: '系统管理',
    children: [
      { key: '/system/admin', icon: <TeamOutlined />, label: '管理员管理' },
      { key: '/system/role', icon: <SafetyOutlined />, label: '角色权限' },
      { key: '/system/menu', icon: <MenuOutlined />, label: '菜单配置' },
    ],
  },
  {
    key: '/business',
    icon: <UserOutlined />,
    label: '业务管理',
    children: [
      { key: '/user', icon: <UserOutlined />, label: '用户管理' },
      { key: '/enterprise', icon: <BankOutlined />, label: '企业审核' },
    ],
  },
  {
    key: '/product-center',
    icon: <ShoppingOutlined />,
    label: '商品中心',
    children: [
      { key: '/product', icon: <AppstoreOutlined />, label: '商品管理' },
      { key: '/product/order', icon: <ShoppingOutlined />, label: '订单管理' },
      { key: '/product/category', icon: <AppstoreOutlined />, label: '类目配置' },
    ],
  },
  {
    key: '/log',
    icon: <FileTextOutlined />,
    label: '操作日志',
  },
]

const pageMeta: Record<string, { title: string; description: string }> = {
  '/dashboard': { title: '数据概览', description: '平台数据总览与待办事项' },
  '/system/admin': { title: '管理员管理', description: '管理后台账号与权限' },
  '/system/role': { title: '角色权限', description: '角色与功能权限配置' },
  '/system/menu': { title: '菜单配置', description: '导航结构与功能入口' },
  '/user': { title: '用户管理', description: '平台注册用户管理' },
  '/enterprise': { title: '企业审核', description: '入驻企业资质审核' },
  '/product': { title: '商品管理', description: '商品信息与审核管理' },
  '/product/order': { title: '订单管理', description: '交易订单与履约管理' },
  '/product/category': { title: '类目配置', description: '商品类目结构管理' },
  '/log': { title: '操作日志', description: '管理员操作审计日志' },
}

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { adminInfo, logout } = useAuthStore()
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken()
  const currentMeta = pageMeta[location.pathname] ?? pageMeta['/dashboard']

  const getOpenKeys = () => {
    const path = location.pathname
    if (path.startsWith('/system')) return ['/system']
    if (path === '/user' || path === '/enterprise') return ['/business']
    if (path.startsWith('/product')) return ['/product-center']
    return []
  }

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
      label: '退出登录',
      onClick: handleLogout,
    },
  ]

  return (
    <Layout className={styles.layout}>
      <Sider trigger={null} collapsible collapsed={collapsed} className={styles.sider}>
        <div className={styles.logo}>
          {collapsed ? 'SL' : (
            <div>
              SteelLink
              <span className={styles.logoSub}>运营管理后台</span>
            </div>
          )}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          defaultOpenKeys={getOpenKeys()}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <Layout>
        <Header className={styles.header} style={{ background: colorBgContainer }}>
          <div className={styles.headerLeft}>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              className={styles.trigger}
            />
            <div className={styles.headerMeta}>
              <div className={styles.pageTitle}>{currentMeta.title}</div>
              <div className={styles.pageDesc}>{currentMeta.description}</div>
            </div>
          </div>
          <div className={styles.headerRight}>
            <div className={styles.statusPill}>
              <span className={styles.statusDot} />
              服务正常
            </div>
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div className={styles.userAvatar}>
                <Avatar size="small" icon={<UserOutlined />} />
                <span className={styles.userName}>
                  {adminInfo?.realName === 'Platform Admin' ? '管理员' : (adminInfo?.realName || '管理员')}
                </span>
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
          <div className={styles.contentShell}>{children || <Outlet />}</div>
        </Content>
      </Layout>
    </Layout>
  )
}
