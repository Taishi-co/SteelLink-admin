# 实现方案

## 文件变更清单

### 1. 买家端订单详情页
**文件路径**：`apps/buyer-app/src/pages/OrderDetail/index.tsx`

**修改内容**：
- 修复状态映射：状态15改为"待买方确认订单"
- 添加状态15的确认按钮逻辑
- 修改物流报价按钮显示条件为状态16
- 修复物流竞价Modal中的状态判断（15/16 → 16/17）
- 更新时间线组件，添加"选择目的地仓储"步骤

### 2. 买家端订单列表页
**文件路径**：`apps/buyer-app/src/pages/Orders/index.tsx`

**修改内容**：
- 修复状态映射表：
  - 状态15："物流竞价中" → "待买方确认订单"
  - 状态16："请确认物流资源" → "物流竞价中"

### 3. 卖家端订单列表页
**文件路径**：`apps/seller-app/src/pages/Orders/index.tsx`

**修改内容**：
- 添加状态15的guidance提示（待买方最终确认）
- 修复物流竞价Modal中的状态判断

### 4. 后端订单服务
**文件路径**：`internal/modules/order/application/service/order_service.go`

**修改内容**：
- 在`buildOrderTimeline`函数中添加"选择目的地仓储"步骤
- 优化时间戳记录机制，从`OrderEvent`表获取准确操作时间

## 实现步骤

### 步骤1：修改买家端订单列表页状态映射
1. 打开 `apps/buyer-app/src/pages/Orders/index.tsx`
2. 更新statusMeta对象中的状态15和16的显示文本

### 步骤2：修改买家端订单详情页
1. 打开 `apps/buyer-app/src/pages/OrderDetail/index.tsx`
2. 修复getStatusNotice函数中的状态15判断
3. 添加状态15的确认按钮
4. 修改物流报价按钮显示条件
5. 修复物流竞价Modal的状态判断
6. 添加"选择目的地仓储"时间线步骤

### 步骤3：修改卖家端订单列表页
1. 打开 `apps/seller-app/src/pages/Orders/index.tsx`
2. 添加状态15的guidance提示
3. 修复物流竞价Modal的状态判断

### 步骤4：修改后端订单服务
1. 打开 `internal/modules/order/application/service/order_service.go`
2. 在`buildOrderTimeline`函数中添加"选择目的地仓储"步骤
3. 确保时间戳从`OrderEvent`表正确获取

## 技术要点

### 时间戳记录机制
- 后端使用`OrderEvent`表记录每个状态变更事件
- 每个步骤完成时记录准确的操作时间（CreatedAt）
- 时间线展示时从事件表获取对应时间戳
- 无事件记录时回退使用订单的`UpdatedAt`

### 状态流转顺序
```
12 待选择目的地仓储 → 13 待仓库确认 → 14 待卖方确认 → 15 待买方确认 → 2 待签合同 → 3 待支付 → 4 已支付 → 16 物流竞价中 → 17 待确认物流
```

## 影响范围

- **买家端**：订单详情页、订单列表页
- **卖家端**：订单列表页
- **后端**：订单时间线构建逻辑