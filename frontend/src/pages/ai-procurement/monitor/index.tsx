import { useState } from 'react'
import { Alert, Button, Card, Col, Descriptions, Form, Input, Row, Select, Space, Statistic, Table, Tag, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { ReloadOutlined, SearchOutlined, ThunderboltOutlined } from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import {
  aiProcurementApi,
  AIProcurementMonitorCount,
  AIProcurementMonitorItem,
  AIProcurementMonitorParams,
} from '../../../services/api'
import { formatDateTime } from '../../../utils/format'

const { Text } = Typography

const defaultStats = {
  total: 0,
  today: 0,
  active: 0,
  reviewSubmitted: 0,
  failed: 0,
  stale: 0,
  needsIntervention: 0,
}

const stageOptions = [
  { label: '全部阶段', value: '' },
  { label: '指令解析', value: 'instruction_parsing' },
  { label: '商品匹配', value: 'product_matching' },
  { label: '物流/仓储匹配', value: 'logistics_warehouse_matching' },
  { label: '订单生成', value: 'order_generation' },
  { label: '融资匹配', value: 'financing_matching' },
  { label: '审核提交', value: 'review_submission' },
]

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '草稿', value: 'draft' },
  { label: '已解析', value: 'parsed' },
  { label: '已匹配', value: 'matched' },
  { label: '已生成订单', value: 'order_generated' },
  { label: '融资方案就绪', value: 'financing_ready' },
  { label: '已提交审核', value: 'review_submitted' },
  { label: '失败', value: 'failed' },
  { label: '异常', value: 'error' },
  { label: '已取消', value: 'cancelled' },
]

interface FilterValues {
  keyword?: string
  stage?: string
  status?: string
  intervention?: string
}

function stageColor(stage: string) {
  const colors: Record<string, string> = {
    instruction_parsing: 'blue',
    product_matching: 'cyan',
    logistics_warehouse_matching: 'purple',
    order_generation: 'geekblue',
    financing_matching: 'gold',
    review_submission: 'orange',
  }
  return colors[stage] || 'default'
}

function statusColor(status: string) {
  const colors: Record<string, string> = {
    draft: 'default',
    parsed: 'blue',
    matched: 'cyan',
    order_generated: 'geekblue',
    financing_ready: 'purple',
    review_submitted: 'gold',
    failed: 'red',
    error: 'red',
    cancelled: 'default',
  }
  return colors[status] || 'default'
}

function runtimeStatusTag(status: string) {
  if (status === 'healthy') return <Tag color="green">正常</Tag>
  if (status === 'warning') return <Tag color="orange">需关注</Tag>
  if (status === 'disabled') return <Tag>已停用</Tag>
  return <Tag>{status || '未知'}</Tag>
}

function percent(value?: number) {
  if (!value) return '-'
  return `${Math.round(value * 100)}%`
}

function booleanTag(value: boolean, trueText = '是', falseText = '否') {
  return <Tag color={value ? 'green' : 'default'}>{value ? trueText : falseText}</Tag>
}

export default function AIProcurementMonitorPage() {
  const [form] = Form.useForm<FilterValues>()
  const [params, setParams] = useState<AIProcurementMonitorParams>({ page: 1, pageSize: 10 })

  const { data, isLoading, isFetching, refetch } = useQuery({
    queryKey: ['ai-procurement-monitor', params],
    queryFn: () => aiProcurementApi.monitor(params),
  })

  const monitor = data?.data.data
  const stats = monitor?.stats ?? defaultStats
  const records = monitor?.records.list ?? []
  const total = monitor?.records.total ?? 0

  const applyFilters = () => {
    const values = form.getFieldsValue()
    setParams({
      page: 1,
      pageSize: params.pageSize,
      keyword: values.keyword?.trim() || undefined,
      stage: values.stage || undefined,
      status: values.status || undefined,
      intervention: values.intervention === '1' ? 1 : undefined,
    })
  }

  const resetFilters = () => {
    form.resetFields()
    setParams({ page: 1, pageSize: params.pageSize })
  }

  const countColumns: ColumnsType<AIProcurementMonitorCount> = [
    {
      title: '名称',
      dataIndex: 'label',
      key: 'label',
      render: (value: string, record) => <Tag color={record.key ? 'blue' : 'default'}>{value}</Tag>,
    },
    { title: '数量', dataIndex: 'count', key: 'count', width: 100, align: 'right' },
  ]

  const columns: ColumnsType<AIProcurementMonitorItem> = [
    {
      title: '采购单',
      dataIndex: 'procurementNo',
      key: 'procurementNo',
      width: 150,
      fixed: 'left',
      render: (value: string, record) => (
        <Space direction="vertical" size={0}>
          <Text strong>{value || `#${record.id}`}</Text>
          <Text type="secondary">企业 {record.enterpriseId || '-'}</Text>
        </Space>
      ),
    },
    {
      title: '需求',
      key: 'request',
      width: 260,
      render: (_, record) => (
        <Space direction="vertical" size={0} style={{ maxWidth: 240 }}>
          <Text strong ellipsis={{ tooltip: record.productName || record.category || '-' }}>
            {record.productName || record.category || '-'}
          </Text>
          <Text type="secondary" ellipsis={{ tooltip: record.prompt || '-' }}>
            {record.prompt || '-'}
          </Text>
        </Space>
      ),
    },
    {
      title: '数量',
      key: 'quantity',
      width: 120,
      render: (_, record) => `${record.quantity || 0}${record.unit || ''}`,
    },
    {
      title: '阶段',
      dataIndex: 'currentStageLabel',
      key: 'currentStage',
      width: 140,
      render: (value: string, record) => <Tag color={stageColor(record.currentStage)}>{value}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'statusLabel',
      key: 'status',
      width: 130,
      render: (value: string, record) => <Tag color={statusColor(record.status)}>{value}</Tag>,
    },
    {
      title: '解析',
      dataIndex: 'parseConfidence',
      key: 'parseConfidence',
      width: 90,
      render: (value?: number) => percent(value),
    },
    {
      title: '产物',
      key: 'artifacts',
      width: 230,
      render: (_, record) => (
        <Space wrap size={[4, 4]}>
          <Tag>商品 {record.matchCount}</Tag>
          <Tag>订单 {record.orderCount}</Tag>
          <Tag>融资 {record.financePlanCount}</Tag>
          <Tag>审核 {record.auditItemCount}</Tag>
        </Space>
      ),
    },
    {
      title: '处置',
      key: 'intervention',
      width: 230,
      render: (_, record) => (
        record.needsIntervention
          ? <Tag color="red">{record.interventionReason || '需人工介入'}</Tag>
          : <Tag color="green">正常推进</Tag>
      ),
    },
    {
      title: '配置影响',
      key: 'flags',
      width: 170,
      render: (_, record) => (
        <Space wrap size={[4, 4]}>
          {booleanTag(record.financingRequired, '需融资', '无融资')}
          {booleanTag(record.historySaved, '已留痕', '未留痕')}
        </Space>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      key: 'updatedAt',
      width: 160,
      render: (value: string) => formatDateTime(value),
    },
  ]

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      {monitor && !monitor.tableReady && (
        <Alert
          type="warning"
          showIcon
          message="AI采购运行表尚未初始化"
          description="当前只展示配置状态；主流程落库表 ai_procurements 建立后会自动显示运行数据。"
        />
      )}

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic title="总采购指令" value={stats.total} prefix={<ThunderboltOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic title="今日新增" value={stats.today} valueStyle={{ color: '#1677ff' }} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic title="推进中" value={stats.active} valueStyle={{ color: '#389e0d' }} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic title="待人工处理" value={stats.needsIntervention} valueStyle={{ color: '#cf1322' }} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic title="已提交审核" value={stats.reviewSubmitted} valueStyle={{ color: '#d48806' }} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card>
            <Statistic title="异常/停滞" value={stats.failed + stats.stale} valueStyle={{ color: '#d46b08' }} />
          </Card>
        </Col>
      </Row>

      <Card title="运行配置摘要" loading={isLoading && !monitor}>
        {monitor?.config.warnings.map((warning) => (
          <Alert key={warning} type="warning" showIcon message={warning} style={{ marginBottom: 12 }} />
        ))}
        <Descriptions column={{ xs: 1, sm: 2, lg: 4 }} size="small">
          <Descriptions.Item label="运行状态">{runtimeStatusTag(monitor?.config.runtimeStatus || '')}</Descriptions.Item>
          <Descriptions.Item label="总开关">{booleanTag(Boolean(monitor?.config.enabled), '启用', '停用')}</Descriptions.Item>
          <Descriptions.Item label="模型解析">{booleanTag(Boolean(monitor?.config.parsingEnabled), '启用', '停用')}</Descriptions.Item>
          <Descriptions.Item label="模型">{monitor?.config.model || '-'}</Descriptions.Item>
          <Descriptions.Item label="模型密钥">{booleanTag(Boolean(monitor?.config.modelApiKeyConfigured), '已配置', '未配置')}</Descriptions.Item>
          <Descriptions.Item label="图片密钥">{booleanTag(Boolean(monitor?.config.visionApiKeyConfigured), '已配置', '未配置')}</Descriptions.Item>
          <Descriptions.Item label="向量召回">{booleanTag(Boolean(monitor?.config.vectorEnabled), '启用', '停用')}</Descriptions.Item>
          <Descriptions.Item label="向量连接">{booleanTag(Boolean(monitor?.config.vectorConnectionReady), '就绪', '未就绪')}</Descriptions.Item>
          <Descriptions.Item label="召回 TopK">{monitor?.config.vectorTopK ?? '-'}</Descriptions.Item>
          <Descriptions.Item label="最大推荐">{monitor?.config.maxRecommendations ?? '-'}</Descriptions.Item>
          <Descriptions.Item label="融资匹配">{booleanTag(Boolean(monitor?.config.financeEnabled), '启用', '停用')}</Descriptions.Item>
          <Descriptions.Item label="自动提交审核">{booleanTag(Boolean(monitor?.config.autoSubmitReview), '启用', '关闭')}</Descriptions.Item>
          <Descriptions.Item label="更新时间">{formatDateTime(monitor?.config.updatedAt)}</Descriptions.Item>
          <Descriptions.Item label="生成时间">{formatDateTime(monitor?.generatedAt)}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="阶段分布">
            <Table
              rowKey={(record) => record.key || 'empty-stage'}
              columns={countColumns}
              dataSource={monitor?.stageCounts ?? []}
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="状态分布">
            <Table
              rowKey={(record) => record.key || 'empty-status'}
              columns={countColumns}
              dataSource={monitor?.statusCounts ?? []}
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
      </Row>

      <Card
        title="采购流程明细"
        extra={
          <Button icon={<ReloadOutlined />} loading={isFetching} onClick={() => void refetch()}>
            刷新
          </Button>
        }
      >
        <Form form={form} layout="inline" onFinish={applyFilters} style={{ marginBottom: 16 }}>
          <Form.Item name="keyword">
            <Input
              allowClear
              prefix={<SearchOutlined />}
              placeholder="采购单号、需求、品类"
              style={{ width: 220 }}
            />
          </Form.Item>
          <Form.Item name="stage">
            <Select options={stageOptions} style={{ width: 160 }} />
          </Form.Item>
          <Form.Item name="status">
            <Select options={statusOptions} style={{ width: 150 }} />
          </Form.Item>
          <Form.Item name="intervention">
            <Select
              options={[
                { label: '全部记录', value: '' },
                { label: '仅需人工处理', value: '1' },
              ]}
              style={{ width: 150 }}
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">查询</Button>
              <Button onClick={resetFilters}>重置</Button>
            </Space>
          </Form.Item>
        </Form>

        <Table
          columns={columns}
          dataSource={records}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1490 }}
          pagination={{
            current: params.page,
            pageSize: params.pageSize,
            total,
            showTotal: (value) => `共 ${value} 条`,
            onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
          }}
        />
      </Card>
    </Space>
  )
}
