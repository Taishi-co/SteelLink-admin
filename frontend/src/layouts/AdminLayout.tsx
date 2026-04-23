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
  { key: '/dashboard', icon: <DashboardOutlined />, label: '数据概览' },
  { key: '/system/admin', icon: <SettingOutlined />, label: '管理员管理' },
  { key: '/system/role', icon: <SettingOutlined />, label: '角色权限' },
  { key: '/system/menu', icon: <SettingOutlined />, label: '菜单配置' },
  { key: '/user', icon: <UserOutlined />, label: '用户管理' },
  { key: '/enterprise', icon: <BankOutlined />, label: '企业审核' },
  { key: '/product', icon: <ShoppingOutlined />, label: '商品管理' },
  { key: '/product/category', icon: <ShoppingOutlined />, label: '类目配置' },
]

const pageMeta: Record<string, { title: string; description: string }> = {
  '/dashboard': { title: '数据概览', description: '查看平台用户、企业、商品与待处理事项总览。' },
  '/system/admin': { title: '管理员管理', description: '维护管理账号、身份信息与后台访问权限。' },
  '/system/role': { title: '角色权限', description: '配置角色分工与平台功能权限边界。' },
  '/system/menu': { title: '菜单配置', description: '梳理运营后台导航结构与功能入口。' },
  '/user': { title: '用户管理', description: '统一查看平台注册用户、联系方式与状态。' },
  '/enterprise': { title: '企业审核', description: '跟踪入驻企业资料、认证状态与联系人信息。' },
  '/product': { title: '商品管理', description: '审核平台商品、库存与销售基础信息。' },
  '/product/category': { title: '类目配置', description: '管理商品类目结构，为卖家发布提供标准。' },
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
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <Layout>
        <Header className={styles.header} style={{ background: colorBgContainer }}>
          <div className={styles.userInfo}>
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
          <div className={styles.userInfo}>
            <div className={styles.statusPill}>
              <span className={styles.statusDot} />
              平台服务正常
            </div>
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div className={styles.userAvatar}>
                <Avatar icon={<UserOutlined />} />
                <span className={styles.userName}>
                  {adminInfo?.realName === 'Platform Admin' ? '平台管理员' : (adminInfo?.realName || '管理员')}
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
