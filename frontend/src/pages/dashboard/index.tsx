import { Card, Row, Col, Statistic, Table, Tag, Typography } from 'antd'
import { UserOutlined, BankOutlined, ShoppingOutlined, TeamOutlined } from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import { dashboardApi } from '../../services/api'

export default function Dashboard() {
  const { data } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: () => dashboardApi.stats(),
  })

  const stats = data?.data.data
  const { Paragraph } = Typography

  const quickRows = [
    { key: '1', label: '平台用户', value: `${stats?.totalUsers ?? 0} 人`, status: '稳定' },
    { key: '2', label: '入驻企业', value: `${stats?.totalEnterprises ?? 0} 家`, status: '已接入' },
    { key: '3', label: '在架商品', value: `${stats?.totalProducts ?? 0} 个`, status: '运营中' },
    { key: '4', label: '待处理事项', value: `${stats?.pendingAudits ?? 0} 项`, status: '需关注' },
  ]

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="平台注册用户"
              value={stats?.totalUsers ?? 0}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="入驻企业数量"
              value={stats?.totalEnterprises ?? 0}
              prefix={<BankOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="在库商品数量"
              value={stats?.totalProducts ?? 0}
              prefix={<ShoppingOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="待处理事项"
              value={stats?.pendingAudits ?? 0}
              prefix={<TeamOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={15}>
          <Card title="运营观察">
            <Paragraph style={{ marginBottom: 20, color: '#60708d' }}>
              管理端已接入共享业务数据。你可以从左侧导航进入用户、企业、商品与权限配置页面继续处理具体事项。
            </Paragraph>
            <Table
              pagination={false}
              rowKey="key"
              dataSource={quickRows}
              columns={[
                { title: '指标', dataIndex: 'label', key: 'label' },
                { title: '当前值', dataIndex: 'value', key: 'value' },
                {
                  title: '状态',
                  dataIndex: 'status',
                  key: 'status',
                  render: (value: string) => <Tag color="blue">{value}</Tag>,
                },
              ]}
            />
          </Card>
        </Col>
        <Col xs={24} lg={9}>
          <Card title="今日建议">
            <Paragraph style={{ color: '#31425f', lineHeight: 1.9 }}>
              1. 优先检查待处理事项，避免企业审核与商品审核积压。
            </Paragraph>
            <Paragraph style={{ color: '#31425f', lineHeight: 1.9 }}>
              2. 商品管理页已接入真实商品数据，可继续补充审核动作与详情页。
            </Paragraph>
            <Paragraph style={{ marginBottom: 0, color: '#31425f', lineHeight: 1.9 }}>
              3. 若要继续开发，建议下一步先完善企业详情或商品审核流。
            </Paragraph>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
