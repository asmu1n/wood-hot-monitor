import { Router } from 'express';
import { NotificationService } from '@/features/notification/service';
import RequestHandler from '@/config/requestHandler';
import responseBody from '@/config/response';

const router = Router();

// 获取所有通知
router.get(
    '/',
    RequestHandler(async (req, res) => {
        const result = await NotificationService.getNotifications(req.query);

        res.json(responseBody(true, '获取通知成功', result));
    })
);

// 标记为已读
router.patch(
    '/:id/read',
    RequestHandler(async (req, res) => {
        const notification = await NotificationService.markAsRead(req.params.id as string);

        res.json(responseBody(true, '标记为已读成功', { data: notification }));
    })
);

// 全部标记为已读
router.patch(
    '/read-all',
    RequestHandler(async (req, res) => {
        await NotificationService.markAllAsRead();

        res.json(responseBody(true, '全部标记为已读成功'));
    })
);

// 删除通知
router.delete(
    '/:id',
    RequestHandler(async (req, res) => {
        await NotificationService.delete(req.params.id as string);

        res.json(responseBody(true, '删除通知成功'));
    })
);

// 清空所有通知
router.delete(
    '/',
    RequestHandler(async (req, res) => {
        await NotificationService.clearAll();

        res.json(responseBody(true, '清空所有通知成功'));
    })
);

export default router;
