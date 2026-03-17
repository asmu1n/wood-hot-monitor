import express from 'express';
import cookieParser from 'cookie-parser';
import hotspotRouter from '@/features/hotspot/route';
import keywordRouter from '@/features/keyword/route';
import notificationRouter from '@/features/notification/route';
import settingRouter from '@/features/setting/route';
import errorHandler from '@/middleware/errorHandler';
import { BASE_URL, CLIENT_URL, PORT } from 'env.config';
import { createServer } from 'http';
import { Server } from 'socket.io';
import { runHotspotCheck } from '@/tasks/hotspotCheck';
import cron from 'node-cron';
import responseBody from '@/config/response';
import RequestHandler from '@/config/requestHandler';
import { clearKeywordExpansions } from '@/tasks/keywordCleanup';

//服务配置
const app: express.Application = express();
const baseUrl = BASE_URL || '/api';

const httpServer = createServer(app);
const io = new Server(httpServer, {
    cors: {
        origin: CLIENT_URL || 'http://localhost:5173',
        methods: ['GET', 'POST']
    }
});

// 中件
app.use(express.urlencoded({ extended: true }));
app.use(express.json());
app.use(cookieParser());
//路由
app.use(`${baseUrl}/hotspots`, hotspotRouter);
app.use(`${baseUrl}/keywords`, keywordRouter);
app.use(`${baseUrl}/notifications`, notificationRouter);
app.use(`${baseUrl}/settings`, settingRouter);

// Health check
app.get('/api/health', (req, res) => {
    res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

// Manual trigger for hotspot check
app.post(
    '/api/check',
    RequestHandler(async (req, res) => {
        await runHotspotCheck(io);
        res.json(responseBody(true, 'Hotspot check completed'));
    })
);

// 错误处理(必须放在所有路由的最后)
app.use(errorHandler);

// WebSocket connection handling
io.on('connection', socket => {
    console.log('Client connected:', socket.id);

    socket.on('subscribe', (keywords: string[]) => {
        keywords.forEach(kw => socket.join(`keyword:${kw}`));
        console.log(`Socket ${socket.id} subscribed to:`, keywords);
    });

    socket.on('unsubscribe', (keywords: string[]) => {
        keywords.forEach(kw => socket.leave(`keyword:${kw}`));
    });

    socket.on('disconnect', () => {
        console.log('Client disconnected:', socket.id);
    });
});

// Scheduled job: Run hotspot check every 30 minutes
cron.schedule('*/30 * * * *', async () => {
    console.log('🔄 Running scheduled hotspot check...');

    try {
        await runHotspotCheck(io);
        console.log('✅ Scheduled hotspot check completed');
    } catch (error) {
        console.error('❌ Scheduled hotspot check failed:', error);
    }
});

// 每周日凌晨 2 点清理 keyword_expansions 表
cron.schedule('0 2 * * 0', async () => {
    try {
        await clearKeywordExpansions();
        console.log('[Cron] Weekly keyword expansions cleanup completed');
    } catch (err) {
        console.error('[Cron] Failed to clean keyword expansions:', err);
    }
});

httpServer.listen(PORT, () => {
    console.log(`
  🔥 热点监控服务启动成功!
  📡 Server running on http://localhost:${PORT}
  🔌 WebSocket ready
  ⏰ Hotspot check scheduled every 30 minutes
  `);
});

// app.listen(port, () => console.log(`Server running on  http://${getLocalIP()}:${port} 🎉🎉🎉🎉`));
