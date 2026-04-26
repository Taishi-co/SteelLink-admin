import { useEffect } from 'react'
import {
  Alert,
  Button,
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  Row,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { SaveOutlined, ThunderboltOutlined } from '@ant-design/icons'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { aiProcurementApi, AIProcurementConfig, AIProcurementStageConfig } from '../../../services/api'

const supplierPriorityOptions = [
  { label: '已认证供应商优先', value: 'approved_first' },
  { label: '价格优先', value: 'price_first' },
  { label: '库存优先', value: 'stock_first' },
]

const modelProviderOptions = [
  { label: '平台内置', value: 'platform' },
  { label: 'OpenAI兼容', value: 'openai_compatible' },
  { label: '私有模型服务', value: 'private_endpoint' },
]

export default function AIProcurementConfigPage() {
  const [form] = Form.useForm<AIProcurementConfig>()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['ai-procurement-config'],
    queryFn: aiProcurementApi.getConfig,
  })

  const config = data?.data.data

  useEffect(() => {
    if (config) {
      form.setFieldsValue(config)
    }
  }, [config, form])

  const { mutate: saveConfig, isPending } = useMutation({
    mutationFn: (values: AIProcurementConfig) => aiProcurementApi.updateConfig(values),
    onSuccess: () => {
      message.success('AI采购配置已保存')
      void queryClient.invalidateQueries({ queryKey: ['ai-procurement-config'] })
    },
  })

  const handleSave = async () => {
    const values = await form.validateFields()
    saveConfig(values)
  }

  const stageColumns: ColumnsType<AIProcurementStageConfig> = [
    {
      title: '阶段',
      dataIndex: 'name',
      key: 'name',
      width: 160,
      render: (_, __, index) => (
        <Space>
          <Tag color="blue">{index + 1}</Tag>
          <Form.Item name={['stages', index, 'name']} noStyle>
            <Input style={{ width: 120 }} />
          </Form.Item>
        </Space>
      ),
    },
    {
      title: '编码',
      dataIndex: 'code',
      key: 'code',
      width: 180,
      render: (_, __, index) => (
        <Form.Item name={['stages', index, 'code']} noStyle>
          <Input disabled />
        </Form.Item>
      ),
    },
    {
      title: '启用',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 90,
      render: (_, __, index) => (
        <Form.Item name={['stages', index, 'enabled']} valuePropName="checked" noStyle>
          <Switch />
        </Form.Item>
      ),
    },
    {
      title: '必经',
      dataIndex: 'required',
      key: 'required',
      width: 90,
      render: (_, __, index) => (
        <Form.Item name={['stages', index, 'required']} valuePropName="checked" noStyle>
          <Switch />
        </Form.Item>
      ),
    },
    {
      title: '排序',
      dataIndex: 'sort',
      key: 'sort',
      width: 110,
      render: (_, __, index) => (
        <Form.Item name={['stages', index, 'sort']} noStyle>
          <InputNumber min={1} style={{ width: 80 }} />
        </Form.Item>
      ),
    },
    {
      title: '超时秒数',
      dataIndex: 'timeoutSec',
      key: 'timeoutSec',
      width: 130,
      render: (_, __, index) => (
        <Form.Item name={['stages', index, 'timeoutSec']} noStyle>
          <InputNumber min={5} max={300} style={{ width: 100 }} />
        </Form.Item>
      ),
    },
  ]

  const stages = Form.useWatch('stages', form) ?? config?.stages ?? []

  return (
    <Form form={form} layout="vertical">
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Card
          title={
            <Space>
              <ThunderboltOutlined />
              AI采购总开关
            </Space>
          }
          extra={
            <Button type="primary" icon={<SaveOutlined />} loading={isPending} onClick={() => void handleSave()}>
              保存配置
            </Button>
          }
          loading={isLoading}
        >
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item name="enabled" label="启用AI采购" valuePropName="checked">
                <Switch checkedChildren="启用" unCheckedChildren="停用" />
              </Form.Item>
            </Col>
            <Col span={16}>
              <Alert
                type="info"
                showIcon
                message="AI采购配置覆盖采购指令解析、商品匹配、订单生成、融资匹配与审核提交。"
              />
            </Col>
          </Row>
        </Card>

        <Card title="流程阶段配置">
          <Table
            columns={stageColumns}
            dataSource={stages}
            rowKey="code"
            pagination={false}
            scroll={{ x: 760 }}
          />
        </Card>

        <Row gutter={16}>
          <Col span={12}>
            <Card title="模型与解析策略">
              <Form.Item name={['model', 'provider']} label="模型供应方" rules={[{ required: true }]}>
                <Select options={modelProviderOptions} />
              </Form.Item>
              <Form.Item name={['model', 'model']} label="模型名称" rules={[{ required: true }]}>
                <Input placeholder="steel-procurement-router" />
              </Form.Item>
              <Row gutter={12}>
                <Col span={8}>
                  <Form.Item name={['model', 'temperature']} label="温度">
                    <InputNumber min={0} max={1} step={0.1} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['model', 'maxTokens']} label="最大Token">
                    <InputNumber min={256} max={8000} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['model', 'confidenceThreshold']} label="置信阈值">
                    <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
            </Card>
          </Col>
          <Col span={12}>
            <Card title="商品匹配策略">
              <Form.Item name={['matching', 'supplierPriority']} label="供应商优先级">
                <Select options={supplierPriorityOptions} />
              </Form.Item>
              <Row gutter={12}>
                <Col span={8}>
                  <Form.Item name={['matching', 'categoryBoost']} label="品类权重">
                    <InputNumber min={0} max={100} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['matching', 'priceTolerancePct']} label="价差容忍%">
                    <InputNumber min={0} max={50} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['matching', 'maxRecommendations']} label="推荐数量">
                    <InputNumber min={1} max={20} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
            </Card>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col span={12}>
            <Card title="融资匹配策略">
              <Form.Item name={['finance', 'enabled']} label="启用融资匹配" valuePropName="checked">
                <Switch checkedChildren="启用" unCheckedChildren="停用" />
              </Form.Item>
              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item name={['finance', 'minAmount']} label="融资最低金额">
                    <InputNumber min={0} precision={2} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item name={['finance', 'maxAmount']} label="融资最高金额">
                    <InputNumber min={0} precision={2} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
              <Form.Item name={['finance', 'requiredDocs']} label="必需材料">
                <Select mode="tags" tokenSeparators={[',', '，']} />
              </Form.Item>
              <Form.Item name={['finance', 'autoSubmitReview']} label="融资命中后自动提交审核" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Card>
          </Col>
          <Col span={12}>
            <Card title="审核策略">
              <Form.Item name={['review', 'reviewerRole']} label="审核角色">
                <Input placeholder="finance_reviewer" />
              </Form.Item>
              <Form.Item name={['review', 'manualReviewThreshold']} label="人工复核金额阈值">
                <InputNumber min={0} precision={2} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item name={['review', 'riskKeywords']} label="风险关键词">
                <Select mode="tags" tokenSeparators={[',', '，']} />
              </Form.Item>
            </Card>
          </Col>
        </Row>
      </Space>
    </Form>
  )
}
