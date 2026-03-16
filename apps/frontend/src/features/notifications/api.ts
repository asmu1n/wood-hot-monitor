import { BaseRequest } from '@/lib/request';
import type { Notification } from '@repo/types';

export const BASE_URL = '/notifications';

const getAllNotifications = new BaseRequest<
    {
        page?: number;
        limit?: number;
        unreadOnly?: boolean;
    },
    Notification[],
    true
>({ method: 'get', url: BASE_URL });

const markAsRead = new BaseRequest<{ id: string }, Notification>(params => ({ method: 'patch', url: `${BASE_URL}/${params?.id}/read` }));

const markAllAsRead = new BaseRequest<undefined, Notification[], true>({ method: 'patch', url: `${BASE_URL}/read-all` });

const deleteNotification = new BaseRequest<{ id: string }, void>(params => ({ method: 'delete', url: `${BASE_URL}/${params?.id}` }));

const clearNotification = new BaseRequest<undefined, Notification>({ method: 'delete', url: `${BASE_URL}` });

export { getAllNotifications, markAsRead as getNotificationById, markAllAsRead, deleteNotification, clearNotification };
