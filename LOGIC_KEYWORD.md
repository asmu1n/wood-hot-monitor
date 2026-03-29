# Keyword 模块业务逻辑 (关键词管理)

该模块是整个系统的“配置中心”，负责定义监控的目标意图。

## 1. 前后端通信流程

### A. 状态持久化 (HTTP)

* **操作**：用户在前端执行“添加”、“删除”或“切换激活状态”操作。

* **动作**：前端触发 `keywordApi` 请求，后端修改数据库中的 `isActive` 字段。

* **前端实现片段 (`useAppLogic.ts`)**:

```typescript
// 切换关键词状态逻辑 (using useMutation)
const toggleKeywordMutation = useMutation({
    mutationFn: (keyword: Keyword) => toggleKeyword.request({ id: keyword.id }),
    onSuccess: async ({ data: updatedKeyword }) => {
        // 动态同步 WebSocket 订阅状态
        if (updatedKeyword.isActive) {
            subscribeToKeywords([updatedKeyword.text]);
        } else {
            unsubscribeFromKeywords([updatedKeyword.text]);
        }

        // 利用 TanStack Query 自动刷新缓存
        await queryClient.invalidateQueries({ queryKey: ['keywords'] });
    },
    onError: () => {
        showToast('操作失败', 'error');
    }
});

const handleToggleKeyword = (keyword: Keyword) => {
    toggleKeywordMutation.mutate(keyword);
}
```

* **后端实现片段 (`route.ts`)**:

```typescript
// 切换关键词状态逻辑
router.patch('/:id/toggle', RequestHandler(async (req, res) => {
    const updated = await KeywordService.toggle(req.params.id);
    res.json(responseBody(true, '切换关键词成功', { data: updated }));
}));
```

* **目的**：确保配置在服务器端永久生效，供后台爬虫任务读取。

### B. 通信频道同步 (WebSocket)

* **操作**：在 HTTP 请求成功后，前端立即触发 `subscribeToKeywords` 或 `unsubscribeFromKeywords`。

* **前端工具实现 (`keyword/utils.ts`)**:

```typescript
export function subscribeToKeywords(keywords: string[]) {
    const s = getSocket();
    s.emit('subscribe', keywords); // 发送订阅指令到后端
}
```

* **机制**：利用 Socket.IO 的 **Room (房间)** 功能。每个关键词对应一个房间。

* **后端处理逻辑 (`server.ts`)**:

```typescript
io.on('connection', (socket) => {
    // 处理订阅事件
    socket.on('subscribe', (keywords: string[]) => {
        keywords.forEach(kw => socket.join(`keyword:${kw}`));
        console.log(`Socket ${socket.id} subscribed to:`, keywords);
    });

    // 处理取消订阅事件
    socket.on('unsubscribe', (keywords: string[]) => {
        keywords.forEach(kw => socket.leave(`keyword:${kw}`));
    });
});
```

## 2. 核心价值

* **意图对齐**：确保“我想看什么（DB）”与“我正在接收什么（Socket）”实时同步。

* **流量节省**：通过在服务端进行房间过滤，避免将无关的抓取消息分发给未关注该词的用户。

* **精确控制**：在单用户多标签页场景下，通过动态取消订阅，可以精确控制不同页面接收的消息流。
