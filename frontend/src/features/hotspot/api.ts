import { GetAll, GetByID, GetStatus, Search, Delete, Check, GetNotifications, UnreadCount, MarkRead, MarkAllRead } from '@wails/hotspot/service.js';
import type { Hotspot, Status } from '@/types';

export interface HotspotFilters {
    page?: number;
    limit?: number;
    source?: string;
    importance?: string;
    keywordId?: string;
    isReal?: string;
    timeRange?: string;
    timeFrom?: string;
    timeTo?: string;
    sortBy?: string;
    sortOrder?: 'desc' | 'asc';
}

export interface PaginatedHotspots {
    data: Hotspot[];
    total: number;
    page: number;
    limit: number;
}

export const hotspotApi = {
    getAll: (params: HotspotFilters = {}) =>
        GetAll({
            page: params.page ?? 1,
            limit: params.limit ?? 20,
            source: params.source ?? null,
            importance: params.importance ?? null,
            keywordId: params.keywordId ?? null,
            isReal: params.isReal != null ? params.isReal === 'true' : null,
            timeRange: params.timeRange ?? null,
            timeFrom: params.timeFrom ?? null,
            timeTo: params.timeTo ?? null,
            sortBy: params.sortBy ?? null,
            sortOrder: params.sortOrder ?? null
        }) as Promise<PaginatedHotspots | null>,

    getById: (id: string) => GetByID(id) as Promise<Hotspot | null>,

    getStatus: () => GetStatus() as Promise<Status | null>,

    search: (query: string, sources?: string[]) => Search(query, sources ?? []) as Promise<Hotspot[]>,

    delete: (id: string) => Delete(id),

    check: () => Check(),

    getNotifications: (limit: number = 10) => GetNotifications(limit) as Promise<Hotspot[]>,

    unreadCount: () => UnreadCount() as Promise<number>,

    markRead: (id: string) => MarkRead(id),

    markAllRead: () => MarkAllRead()
};
