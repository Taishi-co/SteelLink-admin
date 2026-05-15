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

const matchingModeOptions = [
  { label: '混合召回', value: 'hybrid' },
  { label: '向量优先', value: 'vector_first' },
  { label: '关键词优先', value: 'keyword_first' },
  { label: '规则兜底', value: 'rule_based' },
]

const modelProviderOptions = [
  { label: 'MiniMax', value: 'minimax' },
  { label: 'OpenAI兼容', value: 'openai_compatible' },
  { label: '平台内置规则', value: 'rule_based' },
  { label: '私有模型服务', value: 'private_endpoint' },
]

const visionProviderOptions = [
  { label: 'MiniMax Token Plan', value: 'minimax_token_plan' },
  { label: 'Chat Completions图片模型', value: 'chat_completions' },
]

const modelNameOptions = [
  { label: 'MiniMax-M2.7', value: 'MiniMax-M2.7' },
  { label: 'MiniMax-M2.7-highspeed', value: 'MiniMax-M2.7-highspeed' },
  { label: 'MiniMax-M2.5', value: 'MiniMax-M2.5' },
  { label: 'MiniMax-M2.1', value: 'MiniMax-M2.1' },
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
  const modelConfig = Form.useWatch('model', form) ?? config?.model

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
              <Form.Item name={['model', 'parsingEnabled']} label="启用模型解析" valuePropName="checked">
                <Switch checkedChildren="启用" unCheckedChildren="规则兜底" />
              </Form.Item>
              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item name={['model', 'provider']} label="模型供应方" rules={[{ required: true }]}>
                    <Select options={modelProviderOptions} />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item name={['model', 'model']} label="模型名称" rules={[{ required: true }]}>
                    <Select options={modelNameOptions} showSearch />
                  </Form.Item>
                </Col>
              </Row>
              <Form.Item name={['model', 'baseUrl']} label="Chat Completions Endpoint" rules={[{ required: true }]}>
                <Input placeholder="https://api.minimax.io/v1/chat/completions" />
              </Form.Item>
              <Row gutter={12}>
                <Col span={16}>
                  <Form.Item name={['model', 'apiKeyEnv']} label="API Key 环境变量" rules={[{ required: true }]}>
                    <Input placeholder="MINIMAX_API_KEY" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item label="密钥状态">
                    <Tag color={modelConfig?.apiKeyConfigured ? 'green' : 'orange'}>
                      {modelConfig?.apiKeyConfigured ? '已配置' : '未检测到'}
                    </Tag>
                  </Form.Item>
                </Col>
              </Row>
              <Alert
                type="warning"
                showIcon
                style={{ marginBottom: 16 }}
                message="API Key 不在管理端落库。请在后端运行环境配置对应密钥，页面只保存环境变量名。"
              />
              <div style={{ fontWeight: 600, marginBottom: 12 }}>商品截图识别</div>
              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item name={['model', 'visionProvider']} label="图片模型供应方" rules={[{ required: true }]}>
                    <Select options={visionProviderOptions} />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name={['model', 'visionBaseUrl']}
                    label="MiniMax Token Plan VLM Endpoint"
                    rules={[{ required: true }]}
                  >
                    <Input placeholder="https://api.minimax.io/v1/coding_plan/vlm" />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={16}>
                  <Form.Item
                    name={['model', 'visionApiKeyEnv']}
                    label="Token Plan Key 环境变量"
                    rules={[{ required: true }]}
                  >
                    <Input placeholder="MINIMAX_TOKEN_PLAN_API_KEY" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item label="图片密钥状态">
                    <Tag color={modelConfig?.visionApiKeyConfigured ? 'green' : 'orange'}>
                      {modelConfig?.visionApiKeyConfigured ? '已配置' : '未检测到'}
                    </Tag>
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={8}>
                  <Form.Item name={['model', 'temperature']} label="温度">
                    <InputNumber min={0} max={1} step={0.1} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['model', 'topP']} label="Top P">
                    <InputNumber min={0} max={1} step={0.05} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['model', 'confidenceThreshold']} label="置信阈值">
                    <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={8}>
                  <Form.Item name={['model', 'maxTokens']} label="最大Token">
                    <InputNumber min={256} max={2048} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['model', 'timeoutSec']} label="超时秒数">
                    <InputNumber min={3} max={120} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['model', 'retryCount']} label="失败重试">
                    <InputNumber min={0} max={3} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
            </Card>
          </Col>
          <Col span={12}>
            <Card title="商品匹配策略">
              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item name={['matching', 'mode']} label="匹配模式">
                    <Select options={matchingModeOptions} />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item name={['matching', 'vectorEnabled']} label="启用向量检索" valuePropName="checked">
                    <Switch checkedChildren="启用" unCheckedChildren="关闭" />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item name={['matching', 'vectorProvider']} label="向量数据库">
                    <Input placeholder="pgvector" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item name={['matching', 'vectorConnectionEnv']} label="pgvector DSN环境变量">
                    <Input placeholder="PGVECTOR_DSN" />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={8}>
                  <Form.Item name={['matching', 'vectorEndpoint']} label="向量库连接串">
                    <Input placeholder="可选，推荐使用 PGVECTOR_DSN 环境变量" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['matching', 'vectorDatabase']} label="Postgres Database">
                    <Input placeholder="steellink_vectors" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['matching', 'vectorUsername']} label="Postgres 用户名">
                    <Input placeholder="可选" />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={8}>
                  <Form.Item name={['matching', 'vectorPasswordEnv']} label="Postgres密码环境变量">
                    <Input placeholder="PGVECTOR_PASSWORD" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['matching', 'productCollection']} label="商品向量表">
                    <Input placeholder="ai_product_embeddings" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['matching', 'supplierCollection']} label="供应商向量表">
                    <Input placeholder="ai_supplier_embeddings" />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={8}>
                  <Form.Item name={['matching', 'financeProductCollection']} label="融资产品向量表">
                    <Input placeholder="ai_finance_product_embeddings" />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={8}>
                  <Form.Item name={['matching', 'embeddingProvider']} label="Embedding供应方">
                    <Input placeholder="minimax" />
                  </Form.Item>
                </Col>
                <Col span={16}>
                  <Form.Item name={['matching', 'embeddingBaseUrl']} label="Embedding Endpoint">
                    <Input placeholder="https://api.minimax.io/v1/embeddings" />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={8}>
                  <Form.Item name={['matching', 'embeddingApiKeyEnv']} label="Embedding Key环境变量">
                    <Input placeholder="MINIMAX_API_KEY" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['matching', 'embeddingModel']} label="Embedding模型">
                    <Input placeholder="embo-01" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['matching', 'embeddingDimension']} label="向量维度">
                    <InputNumber min={128} max={8192} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={8}>
                  <Form.Item name={['matching', 'vectorTopK']} label="向量TopK">
                    <InputNumber min={1} max={100} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['matching', 'vectorMinScore']} label="最低相似度">
                    <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['matching', 'hybridKeywordWeight']} label="关键词权重">
                    <InputNumber min={0} max={1} step={0.05} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={['matching', 'hybridVectorWeight']} label="向量权重">
                    <InputNumber min={0} max={1} step={0.05} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
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
