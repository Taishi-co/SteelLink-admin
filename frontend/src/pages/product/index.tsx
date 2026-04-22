import { useState } from 'react'
import { Card, Table, Tag, Button, Space, Modal, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { productApi, ProductItem, PageParams } from '../../services/api'

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
      message.success('商品状态已更新')
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
      render: (status: number) => {
        const colors: Record<number, string> = { 0: 'default', 1: 'orange', 2: 'green', 3: 'blue', 4: 'red' }
        const texts: Record<number, string> = { 0: 'Draft', 1: 'Pending', 2: 'On Sale', 3: 'Off Sale', 4: 'Rejected' }
        return <Tag color={colors[status] || 'default'}>{texts[status] || `Status ${status}`}</Tag>
      },
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
      width: 120,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => handleAudit(record, 2)}>Approve</Button>
          <Button type="link" danger size="small" onClick={() => handleAudit(record, 4)}>Reject</Button>
        </Space>
      ),
    },
  ]

  return (
    <Card title="Product Management">
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
