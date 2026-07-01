import { GetAll, MarkAsRead, MarkAllAsRead, Delete, ClearAll } from '@wails/notification/service.js';
import type { Notification } from '@/types';

export interface PaginatedNotifications {
    data: Notification[];
    total: number;
    page: number;
    limit: number;
}

export const notificationApi = {
    getAll: (params: { page?: number; limit?: number; unreadOnly?: boolean } = {}) =>
        GetAll({
            page: params.page ?? 1,
            limit: params.limit ?? 20,
            unreadOnly: params.unreadOnly ?? false
        }) as Promise<PaginatedNotifications | null>,

    markAsRead: (id: string) => MarkAsRead(id),

    markAllAsRead: () => MarkAllAsRead(),

    delete: (id: string) => Delete(id),

    clearAll: () => ClearAll()
};
