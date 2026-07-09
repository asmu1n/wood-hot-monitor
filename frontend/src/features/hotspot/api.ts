import { GetAll, GetByID, GetStatus, Search, Delete, GetNotifications, UnreadCount, MarkRead, MarkAllRead } from '@wails/biz/hotspot/service';
import type { GetAllParams, SearchParams } from '@wails/biz/hotspot';

export const hotspotApi = {
    getAll: (params: GetAllParams) =>
        GetAll({
            page: params.page ?? 1,
            limit: params.limit ?? 20,
            source: params.source ?? null,
            importance: params.importance ?? null,
            keywordId: params.keywordId ?? null,
            isReal: params.isReal != null ? params.isReal === true : null,
            timeRange: params.timeRange ?? null,
            timeFrom: params.timeFrom ?? null,
            timeTo: params.timeTo ?? null,
            sortBy: params.sortBy ?? null,
            sortOrder: params.sortOrder ?? null
        }),

    getById: (id: string) => GetByID(id),

    getStatus: () => GetStatus(),

    search: (params: SearchParams) => Search(params),

    delete: (id: string) => Delete(id),

    getNotifications: (limit: number = 10) => GetNotifications(limit),

    unreadCount: () => UnreadCount(),

    markRead: (id: string) => MarkRead(id),

    markAllRead: () => MarkAllRead()
};
