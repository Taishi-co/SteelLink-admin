import { useState } from 'react'
import { Card, Table, Button, Space, Modal, Select, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { productApi, ProductItem, PageParams } from '../../services/api'
import { productStatusMeta, renderStatusTag } from '../../utils/status'

export default function ProductList() {
  const [params, setParams] = useState<PageParams>({ page: 1, pageSize: 10 })
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['products', params],
    queryFn: () => productApi.list(params),
  })

  const auditMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: number }) =>
      productApi.audit(id, { status }),
    onSuccess: () => {
      message.success('商品审核状态已更新')
      queryClient.invalidateQueries({ queryKey: ['products'] })
    },
  })

  const handleAudit = (record: ProductItem, status: number) => {
    Modal.confirm({
      title: status === 2 ? '确认上架该商品？' : '确认驳回该商品？',
      content: record.name,
      onOk: () => auditMutation.mutate({ id: record.id, status }),
    })
  }

  const columns: ColumnsType<ProductItem> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Spec', dataIndex: 'spec', key: 'spec' },
    { title: 'Unit', dataIndex: 'unit', key: 'unit' },
    {
      title: 'Price',
      dataIndex: 'price',
      key: 'price',
      render: (price: number) => `¥${price.toFixed(2)}`,
    },
    { title: 'Stock', dataIndex: 'stock', key: 'stock' },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: number) => renderStatusTag(status, productStatusMeta),
    },
    {
      title: '审核备注',
      dataIndex: 'auditRemark',
      key: 'auditRemark',
      width: 220,
      render: (value?: string) => value || '-',
    },
    {
      title: 'Created At',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 180,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" disabled={record.status === 2} onClick={() => handleAudit(record, 2)}>
            通过
          </Button>
          <Button type="link" danger size="small" disabled={record.status === 4} onClick={() => handleAudit(record, 4)}>
            驳回
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <Card title="商品审核">
      <Space className="mb-4" wrap>
        <Select
          value={params.status}
          placeholder="筛选审核状态"
          allowClear
          style={{ width: 180 }}
          onChange={(status) => setParams({ ...params, page: 1, status })}
          options={[
            { label: '草稿', value: 0 },
            { label: '待审核', value: 1 },
            { label: '已上架', value: 2 },
            { label: '已下架', value: 3 },
            { label: '已驳回', value: 4 },
          ]}
        />
      </Space>
      <Table
        columns={columns}
        dataSource={data?.data.data.list || []}
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: params.page,
          pageSize: params.pageSize,
          total: data?.data.data.total || 0,
          onChange: (page, pageSize) => setParams({ ...params, page, pageSize }),
        }}
      />
    </Card>
  )
}
