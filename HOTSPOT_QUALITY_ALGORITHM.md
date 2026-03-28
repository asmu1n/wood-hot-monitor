# 热点质量分析与过滤算法文档

本文档详细说明了系统中用于跨平台热点质量评估、过滤与评分的统一算法逻辑。该算法旨在通过多维度指标对各个源（Twitter、Bilibili、HackerNews 等）提取的内容进行归一化评估。

## 1. 核心设计理念

系统采用**三级漏斗模式**对抓取内容进行“去噪”：

1. **平台侧预过滤 (Platform-side Pre-filtering)**：利用搜索引擎高级语法（如 Twitter 的 `min_faves`）在获取阶段排除大部分低质量内容。
2. **指标性过滤与评分 (Heuristic Metrics)**：基于点赞、转发、粉丝数等硬性指标进行脱水，并转化为归一化的质量分。
3. **语义深度分析 (Semantic AI Analysis)**：由 AI 对内容的相关性、重要性和真实性进行终审。

---

## 2. 统一属性映射 (Unified Attribute Mapping)

为了实现跨平台评分，系统将不同平台的原始属性映射到 6 个统一属性中：

| 统一属性 | Twitter 对应 | Bilibili 对应 | HackerNews 对应 | 搜索引擎对应 |
| :--- | :--- | :--- | :--- | :--- |
| `likes` | 点赞数 (Likes) | 点赞数 (Likes) | 得分 (Points) | - |
| `shares` | 转发数 + 引用数 | **收藏数 (Favorites)** | - | - |
| `comments` | 回复数 (Replies) | 评论数 + 弹幕数 | 评论数 | - |
| `views` | 查看数 (Views) | 播放数 (Plays) | - | - |
| `followers` | 账号粉丝数 | - | - | - |
| `isVerified` | 蓝V / 认证账号 | 认证账号 | - | - |

> [!NOTE]
> 对于 Bilibili，我们将“收藏数”映射为 `shares`。在 B站生态中，收藏是比点赞更高权重的质量信号，代表了内容的长期价值和传播潜力。

---

## 3. 启发式评分逻辑 (Scoring Logic)

评分引擎 `QualityScorer` 会根据统一后的指标计算三个维度的分值：

### 3.1 互动量评分 (Engagement Score)

基于平台特性应用不同的权重公式（详见第 4 节）。
**计算公式**：`EngagementScore = Min(100, (归一化互动值 / 平台门槛) * 50)`

### 3.2 权威度评分 (Authority Score)

基于作者的社交资产和身份：

- **身份加成**：通过认证的账号（如 Twitter 蓝V）自动获得 **+20** 分，且其过滤阈值减半。
- **粉丝加权**：根据粉丝数阶梯式增加分值。

### 3.3 时效性衰减 (Recency Decay)

**计算公式**：`RecencyScore = Max(0, 100 - 发布时长(小时) * 0.6)`
确保 24 小时内的内容具有极高的展示优先级。

---

## 4. 平台评分策略描述

不同平台的互动行为具有不同的含金量，系统为此设定了差异化的计算权重：

### 4.1 Twitter 策略：侧重“传播性”

Twitter 的核心价值在于信息的病毒式扩散。

- **公式**：`likes + shares * 2 + comments * 1.5`
- **逻辑**：转发（Shares）权重最高，因为转发带动了二阶传播；回复（Comments）权重次之，代表了讨论密度。

### 4.2 Bilibili 策略：侧重“内容质量与收藏”

B站作为一个长视频社区，用户的收藏行为是极强的质量背书。

- **公式**：`likes + shares * 3 + comments * 2`
- **逻辑**：**收藏 (映射为 shares) 的权重最高 (x3)**。评论与弹幕合并计入互动，权重为 2。这能有效筛选出高含金量的硬核内容。

### 4.3 HackerNews 策略：侧重“讨论活跃度”

HN 只有 Points 和评论两个维度。

- **公式**：`likes + comments * 2`
- **逻辑**：由于 HN 用户更看重讨论质量，评论数的权重设为 2，以提拔那些讨论激烈的深度话题。

### 4.4 搜索引擎策略 (Bing, Google, Sogou 等)

搜索引擎爬取的结果通常缺乏点赞、评论等实时互动指标。

- **采集过滤**：排除了广告、侧边栏推荐位以及类似“大家还在搜”等噪音数据。
- **去重处理**：系统采用 **URL 标准化去重 (Deduplication)**，通过协议头统一、路径标准化（移除末尾 `/` 等）来确保跨源结果的唯一性。
- **AI 动态提权**：对于缺乏互动的网页，系统的 AI 深度分析阶段起着关键作用。如果 AI 判定其相关性（Relevance）或重要性（Importance）很高，其在列表中的位置会显著上升。

---

## 5. 质量过滤门槛

| 平台 | 最低点赞 | 最低综合互动量 | 最低展示/播放量 | 最低粉丝数 |
| :--- | :--- | :--- | :--- | :--- |
| **Twitter** | 10 | 15 | 500 | 100 |
| **Bilibili** | 50 | 100 | 1000 | 500 |
| **HackerNews** | 5 | 2 | - | - |
| **SearchEngines**| 0 | 0 | - | - |

---

## 6. 技术实现参考

核心代码位于 `apps/backend/src/features/hotspot/quality/` 目录下：

- `constants.ts`: 存储配置。
- `scoring.ts`: 核心评分计算。
- `manager.ts`: 批量处理入口。

### 接入示例

```typescript
// 转换时进行对齐
const metricInput = {
    platform: 'bilibili',
    likes: video.like,
    shares: video.favorites, // 映射到 shares
    comments: video.review + video.danmaku,
    // ...
};

const ranked = QualityManager.process([metricInput]);
```
