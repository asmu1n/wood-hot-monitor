# Hotspot 模块业务逻辑 (热点监控)

该模块是系统的数据生产与消费中心，处理异步的数据抓取与精准推送。

## 1. 业务协同流程

### A. 后台数据生产 (异步引擎)

* **触发源**：

    1. **定时任务 (Cron)**：每 30 分钟自动运行一次。
    2. **手动触发**：用户点击“立即检查”按钮（通过 HTTP 请求启动，但不等待结果返回）。

* **前端实现片段 (`useAppLogic.ts`)**:

```typescript
// 手动触发检查请求
const handleManualCheck = async () => {
    await attempt(() => {
        setIsChecking(true);
        return checkHotSpot.request(); // 发送异步任务启动指令
    });
    setIsChecking(false);
    showToast('热点检查已触发', 'success');
};
```

* **执行逻辑片段 (`server.ts`)**:

```typescript
// 定时任务配置
cron.schedule('*/30 * * * *', async () => {
    await runHotspotCheck(io);
});
```

* **执行过程**：

    1. 从数据库加载所有 `isActive` 的关键词。
    2. 并行抓取（Twitter, 微博, B站, Bing 等多源数据）。
    3. **AI 分析**：对内容进行去重、真实性校验、相关性评分、智能总结。
    4. **持久化**：将高质量热点存入 `hotspots` 表。

### B. 实时分发 (WebSocket 推送)

* **后端实现片段 (`hotspotCheck.ts`)**:

```typescript
// 精准推送到关键词 Room
io.to(`keyword:${keyword.text}`).emit('hotspot:new', hotspot);
```

* **特性**：

  * **定向性**：热点只会被推送到关注了该词的“房间”里。
  * **非阻塞**：后端任务无需等待前端响应，分析完一个就推一个，实现“流式”交互。

## 2. 前端展示逻辑

* **实时监听 (`useAppLogic.ts`)**:

```typescript
useEffect(() => {
    // 注册新热点回调
    const unSubHotSpot = onNewHotSpot(async (hotspot) => {
        // 将新接收的热点推入列表顶部
        setHotSpots(prev => [hotspot as Hotspot, ...prev.slice(0, 19)]);
        showToast('发现新热点: ' + (hotspot as Hotspot).title, 'success');
        await loadData(); // 重新加载数据以保持统计同步
    });
    return () => unSubHotSpot(); // 卸载时取消监听
}, []);
```

* **实时监听工具 (`hotspot/utils.ts`)**:

```typescript
export function onNewHotSpot(callback: (hotspot: HotSpotEvent) => void) {
    const s = getSocket();
    s.on('hotspot:new', callback); // 监听后端推送
    return () => s.off('hotspot:new', callback);
}
```

* **动态更新**：收到新热点后，前端直接 `unshift` 到当前列表顶部，并弹出 Toast 成功提醒。

* **无需刷新**：利用 WebSocket 弥补了异步任务耗时过长（AI 分析通常需 10s+）导致的交互断层。
