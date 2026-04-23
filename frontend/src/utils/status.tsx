import { Tag } from 'antd'

export const enterpriseStatusMeta: Record<number, { color: string; text: string }> = {
  0: { color: 'orange', text: '待审核' },
  1: { color: 'green', text: '已通过' },
  2: { color: 'red', text: '已驳回' },
}

export const productStatusMeta: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '草稿' },
  1: { color: 'orange', text: '待审核' },
  2: { color: 'green', text: '已上架' },
  3: { color: 'blue', text: '已下架' },
  4: { color: 'red', text: '已驳回' },
}

export function renderStatusTag(
  status: number,
  metaMap: Record<number, { color: string; text: string }>,
) {
  const meta = metaMap[status] || { color: 'default', text: `状态${status}` }
  return <Tag color={meta.color}>{meta.text}</Tag>
}
