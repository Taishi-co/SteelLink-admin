import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom'
import { lazy, Suspense } from 'react'
import { Spin } from 'antd'
import AdminLayout from '../layouts/AdminLayout'
import { useAuthStore } from '../store/useAuthStore'

// Lazy load pages
const Login = lazy(() => import('../pages/login'))
const Dashboard = lazy(() => import('../pages/dashboard'))
const AdminList = lazy(() => import('../pages/system/admin'))
const RoleList = lazy(() => import('../pages/system/role'))
const MenuList = lazy(() => import('../pages/system/menu'))
const UserList = lazy(() => import('../pages/user'))
const EnterpriseList = lazy(() => import('../pages/enterprise'))
const ProductList = lazy(() => import('../pages/product'))
const OrderList = lazy(() => import('../pages/order'))
const CategoryList = lazy(() => import('../pages/product/category'))
const OperationLogList = lazy(() => import('../pages/log'))

// Loading component
const Loading = () => (
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
    <Spin size="large" />
  </div>
)

// Auth guard component
function AuthGuard() {
  const { token } = useAuthStore()

  if (!token) {
    return <Navigate to="/login" replace />
  }

  return (
    <AdminLayout>
      <Suspense fallback={<Loading />}>
        <Outlet />
      </Suspense>
    </AdminLayout>
  )
}

// Router configuration
const router = createBrowserRouter([
  {
    path: '/login',
    element: (
      <Suspense fallback={<Loading />}>
        <Login />
      </Suspense>
    ),
  },
  {
    path: '/',
    element: <AuthGuard />,
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: 'dashboard',
        element: <Dashboard />,
      },
      {
        path: 'system',
        children: [
          {
            path: 'admin',
            element: <AdminList />,
          },
          {
            path: 'role',
            element: <RoleList />,
          },
          {
            path: 'menu',
            element: <MenuList />,
          },
        ],
      },
      {
        path: 'user',
        element: <UserList />,
      },
      {
        path: 'enterprise',
        element: <EnterpriseList />,
      },
      {
        path: 'product',
        children: [
          {
            index: true,
            element: <ProductList />,
          },
          {
            path: 'order',
            element: <OrderList />,
          },
          {
            path: 'category',
            element: <CategoryList />,
          },
        ],
      },
      {
        path: 'log',
        element: <OperationLogList />,
      },
    ],
  },
  {
    path: '*',
    element: <Navigate to="/dashboard" replace />,
  },
])

export default router
