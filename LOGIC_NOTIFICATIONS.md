# Notifications 模块业务逻辑 (通知系统)

该模块负责全局的系统感知，确保用户在任何页面都能实时获知新动态。

## 1. 业务通信流程

### A. 广播发布 (WebSocket)

* **后端实现片段 (`hotspotCheck.ts`)**:

```typescript
// 全局广播通知
io.emit('notification', {
    type: 'hotspot',
    title: '发现新热点',
    content: hotspot.title,
    hotspotId: hotspot.id,
    importance: hotspot.importance
});
```

* **触发**：每当 Hotspot 模块确认生成一个新热点并入库后。

* **特性**：**全局广播**。不带 Room 参数，发送给所有在线的连接。

* **内容**：包含通知类型、标题、热点 ID 以及重要程度。

### B. 状态管理 (HTTP)

* **前端消费逻辑 (`useAppLogic.ts`)**:

```typescript
useEffect(() => {
    const unSubNotification = onNotification(() => {
        // 收到全局广播后，本地未读计数加一
        setUnreadCount(prev => prev + 1);
    });
    return () => unSubNotification();
}, []);
```

* **前端工具实现 (`notifications/utils.ts`)**:

```typescript
export function onNotification(callback: (notification: NotificationEvent) => void) {
    const s = getSocket();
    s.on('notification', callback);
    return () => s.off('notification', callback);
}
```

* **后端逻辑设计**:

```typescript
// 标记所有通知为已读
router.post('/mark-all-read', RequestHandler(async (req, res) => {
    await NotificationService.markAllAsRead();
    res.json(responseBody(true, '已全部标记为已读'));
}));
```

* **加载**：页面初始化时通过 HTTP 加载历史通知列表。

* **已读处理**：用户点击“清除”或打开通知面板时，通过 HTTP 请求标记已读。

* **同步逻辑**：已读操作会触发数据库变更，并让前端同步更新 `unreadCount`（未读计数）。

## 2. 核心价值

* **横向感知**：即使用户当前停留在“设置”或“历史搜索”等没有订阅特定关键词房间的页面，依然能感知到新热点的产生。

* **轻量交互**：WebSocket 事件负责“触发提醒”（计数加一），具体的“内容拉取”可在需要时通过 HTTP 请求完成，平衡了实时性与数据开销。
