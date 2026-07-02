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
├── main.go              # Wails 应用入口
├── ent/schema/          # Ent ORM schema (keyword, hotspot, keyword_expansion)
├── internal/
│   ├── checker/         # 定时检查调度，处理热点抓取、LLM 评估与入库
│   ├── config/          # 本地 JSON 配置管理，路径为 ~/.wood-hot-monitor/config.json
│   ├── database/        # SQLite 数据库初始化
│   ├── email/           # 邮件通知模块
│   ├── hotspot/         # 热点 CRUD 与通知查询 (GetNotifications, UnreadCount, MarkRead, MarkAllRead)
│   ├── keyword/         # 关键词管理
│   ├── llm/             # LLM 调用实现，支持自定义 BaseURL
│   ├── models/          # Wails 绑定的数据模型，定义 API 契约层
│   ├── quality/         # 热点质量评估逻辑
│   └── scraper/         # 多平台爬虫 (Bilibili, Bing, HackerNews, Twitter)
└── frontend/
    ├── bindings/        # Wails 自动生成的 Go 与 JS 交互绑定
    └── src/
        ├── routes/      # TanStack Router 页面路由 (热点雷达、监控词、搜索、设置)
        ├── features/    # 按功能域组织的代码 (hotspot, keyword, settings)
        ├── components/  # 共享 UI 组件 (Sidebar, ContentHeader, StatusCards 等)
        └── hooks/       # 业务逻辑 hook (useAppLogic)
```

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

## 架构说明

- 后端与客户端一体化：应用不采用传统的 C/S 架构，Go 后端逻辑通过 Wails 绑定机制直接暴露给前端使用。
- 配置存储：程序配置保存在本地 JSON 文件中，未存储于数据库。
- 通知模型：通知状态 (is_read) 直接记录在 Hotspot 数据表中，未设立独立的通知表。
- 数据库设计：选用 SQLite 纯 Go 驱动以规避 CGO 编译问题，schema 变更通过 Ent 自动迁移完成。
- LLM 接口：提供通用的 OpenAI 兼容接口，支持用户配置自定义 BaseURL。

## 许可证

GPL-3.0
