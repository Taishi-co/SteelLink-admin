import { useState, useEffect, useMemo } from 'react'
import { Layout, Menu, Avatar, Dropdown, Button, theme, Spin, Result } from 'antd'
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
  ThunderboltOutlined,
  ShopOutlined,
  SolutionOutlined,
  HomeOutlined,
  BarChartOutlined,
  ProfileOutlined,
  PartitionOutlined,
  ControlOutlined,
  ReconciliationOutlined,
  ContainerOutlined,
  AuditOutlined,
  ToolOutlined,
  ApiOutlined,
  CloudServerOutlined,
  RocketOutlined,
  ExperimentOutlined,
  TruckOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../store/useAuthStore'
import { authApi } from '../services/api'
import type { MenuItem as ApiMenuItem } from '../services/api'
import styles from './AdminLayout.module.css'

type AntdMenuItem = NonNullable<MenuProps['items']>[number] & {
  children?: MenuProps['items']
}

const { Header, Sider, Content } = Layout

const iconMap: Record<string, React.ReactNode> = {
  DashboardOutlined: <DashboardOutlined />,
  SettingOutlined: <SettingOutlined />,
  UserOutlined: <UserOutlined />,
  BankOutlined: <BankOutlined />,
  ShoppingOutlined: <ShoppingOutlined />,
  TeamOutlined: <TeamOutlined />,
  SafetyOutlined: <SafetyOutlined />,
  MenuOutlined: <MenuOutlined />,
  AppstoreOutlined: <AppstoreOutlined />,
  FileTextOutlined: <FileTextOutlined />,
  ThunderboltOutlined: <ThunderboltOutlined />,
  ShopOutlined: <ShopOutlined />,
  SolutionOutlined: <SolutionOutlined />,
  HomeOutlined: <HomeOutlined />,
  BarChartOutlined: <BarChartOutlined />,
  ProfileOutlined: <ProfileOutlined />,
  PartitionOutlined: <PartitionOutlined />,
  ControlOutlined: <ControlOutlined />,
  ReconciliationOutlined: <ReconciliationOutlined />,
  ContainerOutlined: <ContainerOutlined />,
  AuditOutlined: <AuditOutlined />,
  ToolOutlined: <ToolOutlined />,
  ApiOutlined: <ApiOutlined />,
  CloudServerOutlined: <CloudServerOutlined />,
  RocketOutlined: <RocketOutlined />,
  ExperimentOutlined: <ExperimentOutlined />,
  TruckOutlined: <TruckOutlined />,
}

const fallbackMenuItems: MenuProps['items'] = [
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
      { key: '/enterprise/services', icon: <AuditOutlined />, label: '服务权限审核' },
      { key: '/service-offers/warehouse', icon: <ContainerOutlined />, label: '仓储资源审核' },
      { key: '/service-offers/logistics', icon: <TruckOutlined />, label: '物流资源审核' },
      { key: '/service-offers/finance', icon: <BankOutlined />, label: '金融服务审核' },
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
    key: '/ai-procurement',
    icon: <ThunderboltOutlined />,
    label: 'AI采购管理',
    children: [
      { key: '/ai-procurement/config', icon: <SettingOutlined />, label: '采购配置' },
    ],
  },
  {
    key: '/log',
    icon: <FileTextOutlined />,
    label: '操作日志',
  },
]

function apiMenuToAntd(items: ApiMenuItem[]): MenuProps['items'] {
  if (!items || items.length === 0) return fallbackMenuItems

  return items.map((item) => {
    const icon = item.icon ? (iconMap[item.icon] || null) : null
    const result: AntdMenuItem = {
      key: item.path || String(item.id),
      icon,
      label: item.name,
    }
    if (item.children && item.children.length > 0) {
      result.children = apiMenuToAntd(item.children)
    }
    return result
  })
}

function collectMenuPaths(items: ApiMenuItem[]): string[] {
  const paths: string[] = []
  items.forEach((item) => {
    if (item.path && item.type === 1) {
      paths.push(item.path)
    }
    if (item.children?.length) {
      paths.push(...collectMenuPaths(item.children))
    }
  })
  return paths
}

const pageMeta: Record<string, { title: string; description: string }> = {
  '/dashboard': { title: '数据概览', description: '平台数据总览与待办事项' },
  '/system/admin': { title: '管理员管理', description: '管理后台账号与权限' },
  '/system/role': { title: '角色权限', description: '角色与功能权限配置' },
  '/system/menu': { title: '菜单配置', description: '导航结构与功能入口' },
  '/user': { title: '用户管理', description: '平台注册用户管理' },
  '/enterprise': { title: '企业审核', description: '入驻企业资质审核' },
  '/enterprise/services': { title: '服务权限审核', description: '审核企业销售、采购、物流、仓储和金融服务权限' },
  '/service-offers/warehouse': { title: '仓储资源审核', description: '审核仓储方上架的仓储服务资源' },
  '/service-offers/logistics': { title: '物流资源审核', description: '审核物流方上架的物流服务资源' },
  '/service-offers/finance': { title: '金融服务审核', description: '审核资金方上架的金融服务产品' },
  '/product': { title: '商品管理', description: '商品信息与审核管理' },
  '/product/order': { title: '订单管理', description: '交易订单与履约管理' },
  '/product/category': { title: '类目配置', description: '商品类目结构管理' },
  '/ai-procurement/config': { title: 'AI采购配置', description: '配置AI采购流程、模型、商品匹配、融资与审核策略' },
  '/log': { title: '操作日志', description: '管理员操作审计日志' },
}

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  const [menuLoading, setMenuLoading] = useState(true)
  const [apiMenus, setApiMenus] = useState<ApiMenuItem[]>([])
  const navigate = useNavigate()
  const location = useLocation()
  const { adminInfo, logout } = useAuthStore()
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken()

  useEffect(() => {
    authApi.getMyMenus()
      .then((data) => {
        if (Array.isArray(data) && data.length > 0) {
          setApiMenus(data)
        }
      })
      .catch(() => {})
      .finally(() => setMenuLoading(false))
  }, [])

  const menuItems = useMemo(() => {
    if (apiMenus.length > 0) {
      return apiMenuToAntd(apiMenus)
    }
    return fallbackMenuItems
  }, [apiMenus])
  const allowedPaths = useMemo(() => collectMenuPaths(apiMenus), [apiMenus])
  const firstAllowedPath = allowedPaths[0] || '/dashboard'
  const routeAllowed = menuLoading || allowedPaths.length === 0 || allowedPaths.includes(location.pathname)

  useEffect(() => {
    if (!menuLoading && allowedPaths.length > 0 && !allowedPaths.includes(location.pathname)) {
      navigate(firstAllowedPath, { replace: true })
    }
  }, [allowedPaths, firstAllowedPath, location.pathname, menuLoading, navigate])

  const currentMeta = pageMeta[location.pathname] ?? pageMeta['/dashboard']

  const getOpenKeys = () => {
    const path = location.pathname
    if (path.startsWith('/system')) return ['/system']
    if (path === '/user' || path.startsWith('/enterprise') || path.startsWith('/service-offers')) return ['/business']
    if (path.startsWith('/product')) return ['/product-center']
    if (path.startsWith('/ai-procurement')) return ['/ai-procurement']
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
        {menuLoading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: '24px 0' }}>
            <Spin />
          </div>
        ) : (
          <Menu
            theme="dark"
            mode="inline"
            selectedKeys={[location.pathname]}
            defaultOpenKeys={getOpenKeys()}
            items={menuItems}
            onClick={handleMenuClick}
          />
        )}
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
          <div className={styles.contentShell}>
            {routeAllowed ? (
              children || <Outlet />
            ) : (
              <Result status="403" title="无权访问" subTitle="当前角色没有该页面权限。" />
            )}
          </div>
        </Content>
      </Layout>
    </Layout>
  )
}
