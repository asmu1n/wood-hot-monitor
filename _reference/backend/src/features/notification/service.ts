import db from '@/config/database';
import { notifications as notificationsTable } from '@/models/notifications';
import { eq, and, sql, desc, type SQL } from 'drizzle-orm';
import { NotFoundError } from '@/config/error';
import { getPagination } from '@/db/utils/query';
import type { Notification } from '@repo/types';

export class NotificationService {
    /**
     * 获取通知列表
     */
    static async getNotifications({ page, limit, unreadOnly = 'false' }: { page?: number; limit?: number; unreadOnly?: string }): Promise<{
        data: Notification[];
        page: number;
        limit: number;
        total: number;
    }> {
        const filters: SQL[] = [];

        const { pageNum, limitNum, offset } = getPagination(page, limit);

        if (unreadOnly === 'true') {
            filters.push(eq(notificationsTable.isRead, false));
        }

        const where = filters.length > 0 ? and(...filters) : undefined;

        const [notifications, total] = await Promise.all([
            db.query.notifications.findMany({
                where,
                orderBy: [desc(notificationsTable.createdAt)],
                limit: limitNum,
                offset
            }),
            db
                .select({ count: sql<number>`count(*)` })
                .from(notificationsTable)
                .where(where)
        ]);

        const data = notifications.map(item => ({
            ...item,
            createdAt: item.createdAt.toISOString() || ''
        }));

        return {
            data,
            page: pageNum,
            limit: limitNum,
            total: total[0]?.count || 0
        };
    }

    /**
     * 标记为已读
     */
    static async markAsRead(id: string) {
        const result = await db.update(notificationsTable).set({ isRead: true }).where(eq(notificationsTable.id, id)).returning();

        if (result.length === 0) {
            throw new NotFoundError('通知未找到');
        }

        return result[0];
    }

    /**
     * 全部标记为已读
     */
    static async markAllAsRead(): Promise<Notification[]> {
        const result = await db.update(notificationsTable).set({ isRead: true }).where(eq(notificationsTable.isRead, false)).returning();

        return result.map(item => ({
            ...item,
            createdAt: item.createdAt.toISOString() || ''
        }));
    }

    /**
     * 删除通知
     */
    static async delete(id: string) {
        const result = await db.delete(notificationsTable).where(eq(notificationsTable.id, id)).returning();

        if (result.length === 0) {
            throw new NotFoundError('通知未找到');
        }

        return true;
    }

    /**
     * 清空所有通知
     */
    static async clearAll(): Promise<Notification[]> {
        const result = await db.delete(notificationsTable).returning();

        return result.map(item => ({
            ...item,
            createdAt: item.createdAt.toISOString() || ''
        }));
    }
}
