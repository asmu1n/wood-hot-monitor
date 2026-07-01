import { getSocket } from '@/lib/socket';
import type { Hotspot } from '@repo/types';

/** 计算热度综合指标（归一化 0-100） */
function calcHeatScore(h: Hotspot): number {
    const likes = h.likeCount ?? 0;
    const retweets = h.retweetCount ?? 0;
    const replies = h.replyCount ?? 0;
    const comments = h.commentCount ?? 0;
    const quotes = h.quoteCount ?? 0;
    const views = h.viewCount ?? 0;
    // 加权公式：转发最重、其次点赞、然后评论/回复
    const raw = likes * 2 + retweets * 3 + replies * 1.5 + comments * 1.5 + quotes * 2 + views / 100;

    // log 压缩到 0-100
    if (raw <= 0) {
        return 0;
    }

    return Math.min(100, Math.round(Math.log10(raw + 1) * 25));
}

function getHeatLevel(score: number): { label: string; color: string } {
    if (score >= 80) {
        return { label: '爆', color: 'text-red-400' };
    }

    if (score >= 60) {
        return { label: '热', color: 'text-orange-400' };
    }

    if (score >= 40) {
        return { label: '温', color: 'text-amber-400' };
    }

    if (score >= 20) {
        return { label: '凉', color: 'text-blue-400' };
    }

    return { label: '冷', color: 'text-slate-500' };
}
/**
 * 热点排序工具函数（前端版本，与 server/src/utils/sortHotspots.ts 逻辑一致）
 */

interface SortableHotSpot {
    likeCount: number | null;
    retweetCount: number | null;
    viewCount: number | null;
    importance: string;
    relevance: number;
    publishedAt: Date | string | null;
    createdAt: Date | string;
}

export const IMPORTANCE_ORDER: Record<string, number> = {
    urgent: 0,
    high: 1,
    medium: 2,
    low: 3
};

function calcHotScore(item: SortableHotSpot): number {
    const likes = item.likeCount || 0;
    const retweets = item.retweetCount || 0;
    const views = item.viewCount || 0;

    return likes * 10 + retweets * 5 + Math.log10(Math.max(views, 1)) * 2;
}

function compareImportance(a: SortableHotSpot, b: SortableHotSpot): number {
    return (IMPORTANCE_ORDER[a.importance] ?? 4) - (IMPORTANCE_ORDER[b.importance] ?? 4);
}

function toTimestamp(d: Date | string | null): number {
    if (!d) {
        return 0;
    }

    return typeof d === 'string' ? new Date(d).getTime() : d.getTime();
}

function sortHotSpots<T extends SortableHotSpot>(items: T[], sortBy: string, sortOrder: 'asc' | 'desc' = 'desc'): T[] {
    const sorted = [...items];
    const desc = sortOrder === 'desc';

    sorted.sort((a, b) => {
        let result: number;

        switch (sortBy) {
            case 'publishedAt': {
                const ta = toTimestamp(a.publishedAt);
                const tb = toTimestamp(b.publishedAt);

                result = ta - tb;

                if (result === 0) {
                    result = toTimestamp(a.createdAt) - toTimestamp(b.createdAt);
                }

                break;
            }

            case 'importance': {
                result = compareImportance(a, b);

                if (result === 0) {
                    result = toTimestamp(a.createdAt) - toTimestamp(b.createdAt);

                    return desc ? -result : result;
                }

                return desc ? result : -result;
            }

            case 'relevance':
                result = a.relevance - b.relevance;
                break;

            case 'hot':
                result = calcHotScore(a) - calcHotScore(b);
                break;

            default: // createdAt
                result = toTimestamp(a.createdAt) - toTimestamp(b.createdAt);
                break;
        }

        return desc ? -result : result;
    });

    return sorted;
}

interface HotSpotEvent {
    id: string;
    title: string;
    content: string;
    url: string;
    source: string;
    importance: string;
    summary: string | null;
    keyword?: { text: string } | null;
}

function onNewHotSpot(callback: (hotspot: HotSpotEvent) => void): () => void {
    const s = getSocket();

    s.on('hotspot:new', callback);

    return () => s.off('hotspot:new', callback);
}

export { calcHeatScore, getHeatLevel, sortHotSpots, onNewHotSpot };
