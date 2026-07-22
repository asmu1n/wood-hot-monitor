# Wood Hot Monitor

## 项目简介

Wood Hot Monitor 是基于 Wails v3 构建的桌面热点监控应用。该程序能够自动抓取多平台的热点信息，并使用大语言模型对内容质量进行评估。

## 技术栈

后端：
- Go 1.26
- Wails v3 (alpha2.110)
- Ent ORM v0.14
- SQLite (使用 modernc.org/sqlite，纯 Go 实现，无 CGO 依赖)
- gocron v2
- go-openai (支持 OpenAI 兼容接口)

前端：
- React 19
- TanStack Router, TanStack Query
- Tailwind CSS v4
- shadcn/ui
- Framer Motion
- bun 包管理器

## 项目结构

```
├── main.go                        # Wails 应用入口 (组合根，依赖注入)
├── ent/schema/                    # Ent ORM schema (keyword, hotspot, keyword_expansion)
├── internal/
│   ├── module/                    # 业务域模块
│   │   ├── hotspot/               #   热点监控 (实体、接口、服务、参数)
│   │   │   └── persist/           #     Ent 仓储实现
│   │   └── keyword/               #   关键词管理 (实体、接口、服务)
│   │       └── persist/           #     Ent 仓储实现
│   ├── checker/                   # 业务编排 (定时抓取 → LLM 评估 → 入库 → 通知)
│   ├── infra/                     # 基础设施
│   │   ├── database/              #   SQLite 连接初始化
│   │   ├── scraper/               #   多平台爬虫 (Bilibili, Bing, HackerNews, Twitter)
│   │   ├── llm/                   #   LLM 调用 (OpenAI 兼容接口)
│   │   └── notify/                #   通知 (Wails 事件推送 + 邮件告警)
│   ├── config/                    # 本地 JSON 配置管理
│   └── shared/                    # 跨模块共享类型 (分页)
└── frontend/
    ├── bindings/                  # Wails 自动生成的 Go ↔ JS 绑定
    └── src/
        ├── routes/                # TanStack Router 页面路由
        ├── features/              # 按功能域组织 (hotspot, keyword, settings)
        ├── components/            # 共享 UI 组件
        └── hooks/                 # 业务逻辑 hook
```

## 架构设计

### 模块化分层

项目采用按业务域划分的模块化架构：

- **业务域** (`module/`) — 每个模块自包含实体、接口、服务和持久化实现
- **基础设施** (`infra/`) — 为业务模块提供的技术能力 (爬虫、LLM、通知、数据库)
- **编排层** (`checker/`) — 串联业务模块与基础设施，定时执行完整流程

### 依赖方向

```
main.go (组合根)
   │
   ├─→ module/hotspot    (业务接口 + 服务)
   ├─→ module/keyword    (业务接口 + 服务)
   ├─→ checker           (编排，依赖业务接口)
   └─→ infra/*           (实现业务接口)
        │
        └─→ module/*     (引用实体和接口类型)
```

业务模块定义接口，基础设施实现接口，main.go 完成注入。

### 关键设计决策

- **后端与客户端一体化**：Go 后端通过 Wails 绑定机制直接暴露给前端，不走 HTTP API。
- **仓储跟随业务**：每个业务模块自带 `persist/` 子包存放 Ent 仓储实现，开发时在同一目录树下查看全貌。
- **配置存储**：程序配置保存在 `~/.wood-hot-monitor/config.json`，未存入数据库。
- **通知模型**：通知状态 (`is_read`) 直接记录在 Hotspot 表中，无独立通知表。
- **数据库**：SQLite 纯 Go 驱动 (modernc.org/sqlite)，schema 通过 Ent 自动迁移。
- **LLM 接口**：通用 OpenAI 兼容接口，支持用户配置自定义 BaseURL。

## 开发

前置条件：
- Go 1.26 或更高版本
- bun
- Wails CLI v3 (安装命令：go install github.com/wailsapp/wails/v3/cmd/wails3@latest)
- Task (安装命令：go install github.com/go-task/task/v3/cmd/task@latest)

常用命令：

```bash
task dev                      # 开发模式
task build                    # 构建
task package                  # 打包

wails3 generate bindings      # 重新生成前端绑定
go generate ./ent             # 重新生成 Ent 代码

cd frontend && bun dev        # 前端独立开发
```

## 许可证

GPL-3.0
