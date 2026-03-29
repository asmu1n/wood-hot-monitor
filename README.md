# 🌲 Wood Hot Monitor

<p align="center">
  <strong>基于 AI 驱动的多源热点实时监控与分析引擎</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Frontend-React_19-61DAFB?style=for-the-badge&logo=react" alt="React" />
  <img src="https://img.shields.io/badge/Backend-Node.js-339933?style=for-the-badge&logo=nodedotjs" alt="Node.js" />
  <img src="https://img.shields.io/badge/Database-SQLite-003B57?style=for-the-badge&logo=sqlite" alt="SQLite" />
  <img src="https://img.shields.io/badge/Styles-Tailwind_CSS_4-06B6D4?style=for-the-badge&logo=tailwindcss" alt="Tailwind" />
  <img src="https://img.shields.io/badge/Tools-Turborepo-EF4444?style=for-the-badge&logo=turborepo" alt="Turborepo" />
</p>

---

## 📖 项目简介

**Wood Hot Monitor** 是一款高效的实时热点监控工具。它通过爬虫和 AI 技术，从多个社交平台（如 Twitter、Bilibili 等）自动抓取、分析并过滤高价值内容，并利用 WebSocket 技术实现数据的实时推送到客户端。

本项目采用了现代化的微服务（Monorepo）架构，旨在提供极低延迟的热点感知能力，并通过自研的质量评分算法（Quality Scoring）确保信息的准确性与时效性。

## ✨ 核心特性

- 🚀 **实时监控**：利用 Socket.IO 实现全双工通信，热点更新秒级触达。
- 🤖 **AI 驱动分析**：集成 AI 引擎进行多维度内容过滤与分类。
- 📊 **质量评分机制**：基于时间衰减、互动率等指标的动态评分系统。
- 🔍 **关键词定制**：支持用户自定义监控关键词，并提供精确的 Room 级消息路由。
- 🔔 **全方位通知**：结合全局感知系统，通过 Toast 与导航提醒确保不遗漏重要信息。

## 🛠️ 技术栈

### 前端 (apps/frontend)

- **框架**: [React 19](https://react.dev/) + [Vite](https://vitejs.dev/)
- **路由**: [TanStack Router](https://tanstack.com/router)
- **状态管理**: [TanStack Query](https://tanstack.com/query)
- **动画**: [Framer Motion](https://www.framer.com/motion/)
- **UI 组件**: [Radix UI](https://www.radix-ui.com/) + [Lucide Icons](https://lucide.dev/)
- **通信**: [Socket.IO Client](https://socket.io/docs/v4/client-api/)

### 后端 (apps/backend)

- **运行时**: [Node.js](https://nodejs.org/) + [Express](https://expressjs.com/)
- **ORM**: [Drizzle ORM](https://orm.drizzle.team/)
- **数据库**: [Better-SQLite3](https://github.com/WiseLibs/better-sqlite3)
- **任务调度**: [Node-cron](https://github.com/node-cron/node-cron)
- **通信**: [Socket.IO](https://socket.io/)

### 工程化

- **管理**: [Turborepo](https://turbo.build/) + [Yarn v4 (Berry)](https://yarnpkg.com/)
- **类型**: TypeScript (Full-stack Safety)
- **代码规范**: Prettier + ESLint

## 📂 项目结构

```text
.
├── apps/
│   ├── frontend/          # React 前端应用
│   └── backend/           # Express 后端应用
├── packages/
│   ├── types/             # 跨项目共享的 TypeScript 类型定义
│   └── ui/                # 共享 UI 组件库
├── ARCHITECTURE_COMMUNICATION.md  # 详细架构与通信协议说明
├── HOTSPOT_QUALITY_ALGORITHM.md   # 热点质量评分算法说明
└── turbo.json             # Turborepo 配置文件
```

## 🚀 快速开始

### 前提条件

- Node.js (建议 v20+)
- Yarn v4

### 安装依赖

```bash
yarn install
```

### 运行开发环境

```bash
# 同时启动前端和后端
yarn dev

# 仅启动后端服务
yarn serve
```

### 构建项目

```bash
yarn build
```

## 📝 详细文档

为了深入了解系统设计，请参阅以下专业文档：

- 🔌 [通信架构解析](./ARCHITECTURE_COMMUNICATION.md) - 深入了解 HTTP 与 WebSocket 的协作模式。
- 📈 [质量评分算法](./HOTSPOT_QUALITY_ALGORITHM.md) - 详解如何通过数学模型进行内容过滤。
- 🧱 [模块实现细节](./LOGIC_KEYWORD.md) - 开发指南与内部逻辑说明。

---

## 📄 开源协议

本项目采用 [GPL-3.0](./LICENSE) 开源协议。

---

<p align="center">
  Made with ❤️ by Antigravity
</p>
