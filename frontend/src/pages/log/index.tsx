import { useState } from 'react'
import { Card, Table, Tag, Input, Select, Space, Button, DatePicker } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useQuery } from '@tanstack/react-query'
import { operationLogApi, OperationLogItem, OperationLogParams, PageParams } from '../../services/api'
import { formatDateTime } from '../../utils/format'
import { SearchOutlined } from '@ant-design/icons'

const { RangePicker } = DatePicker

const moduleOptions = [
  { label: '全部模块', value: '' },
  { label: '认证', value: 'auth' },
  { label: '管理员', value: 'admin' },
  { label: '用户', value: 'user' },
  { label: '企业', value: 'enterprise' },
  { label: '商品', value: 'product' },
]

const actionColorMap: Record<string, string> = {
  login: 'blue',
  create: 'green',
  update: 'orange',
  delete: 'red',
  audit: 'purple',
  update_status: 'cyan',
}

export default function OperationLogList() {
  const [params, setParams] = useState<OperationLogParams>({ page: 1, pageSize: 20 })
  const [keyword, setKeyword] = useState('')
  const [moduleFilter, setModuleFilter] = useState('')
  const [dateRange, setDateRange] = useState<[string, string] | null>(null)

  const queryParams: OperationLogParams = {
    ...params,
    adminName: keyword || undefined,
    module: moduleFilter || undefined,
    startDate: dateRange?.[0] || undefined,
    endDate: dateRange?.[1] || undefined,
  }

  const { data, isLoading } = useQuery({
    queryKey: ['operation-logs', queryParams],
    queryFn: () => operationLogApi.list(queryParams),
  })

  const handleSearch = () => {
    setParams({ ...params, page: 1 })
  }

  const columns: ColumnsType<OperationLogItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: '操作人', dataIndex: 'adminName', key: 'adminName', width: 120 },
    {
      title: '模块', dataIndex: 'module', key: 'module', width: 100,
      render: (v: string) => <Tag>{v}</Tag>,
    },
    {
      title: '动作', dataIndex: 'action', key: 'action', width: 120,
      render: (v: string) => <Tag color={actionColorMap[v] || 'default'}>{v}</Tag>,
    },
    {
      title: '方法', dataIndex: 'method', key: 'method', width: 70,
      render: (v: string) => <Tag color={v === 'GET' ? 'blue' : v === 'POST' ? 'green' : v === 'PUT' ? 'orange' : 'red'}>{v}</Tag>,
    },
    { title: '路径', dataIndex: 'path', key: 'path', width: 200, ellipsis: true },
    { title: 'IP', dataIndex: 'ip', key: 'ip', width: 130 },
    {
      title: '请求体', dataIndex: 'requestBody', key: 'requestBody', width: 200,
      ellipsis: true,
      render: (v: string) => v || '-',
    },
    {
      title: '状态码', dataIndex: 'responseCode', key: 'responseCode', width: 80,
      render: (v: number) => <Tag color={v === 0 ? 'green' : 'red'}>{v === 0 ? '成功' : v}</Tag>,
    },
    {
      title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 170,
      render: (v: string) => formatDateTime(v),
    },
  ]

  return (
    <Card title="操作日志">
      <Space className="mb-4" wrap>
        <Input
          placeholder="搜索操作人"
          prefix={<SearchOutlined />}
          allowClear
          style={{ width: 200 }}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onPressEnter={handleSearch}
        />
        <Select
          value={moduleFilter}
          onChange={(v) => { setModuleFilter(v); handleSearch() }}
          style={{ width: 140 }}
          options={moduleOptions}
        />
        <RangePicker
          onChange={(_, dates) => {
            if (dates && dates[0] && dates[1]) {
              setDateRange([dates[0], dates[1]])
            } else {
              setDateRange(null)
            }
            handleSearch()
          }}
        />
        <Button type="primary" onClick={handleSearch}>搜索</Button>
      </Space>

      <Table
        columns={columns}
        dataSource={data?.data.data.list || []}
        rowKey="id"
        loading={isLoading}
        scroll={{ x: 1400 }}
        pagination={{
          current: params.page,
          pageSize: params.pageSize,
          total: data?.data.data.total || 0,
          showTotal: (total) => `共 ${total} 条`,
          onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
        }}
      />
    </Card>
  )
}
