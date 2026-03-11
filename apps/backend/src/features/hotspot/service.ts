import db from '@/config/database';
import { hotspots as hotspotsTable } from '@/models/hotspots';
import { eq, and, gte, lte, sql, desc, asc, type SQL } from 'drizzle-orm';
import { sortHotspots } from './utils';
import { searchTwitter } from '@/features/ai/services/twitter';
import { searchBing } from '@/features/ai/services/search';
import { batchAnalyze } from '@/features/ai/services/aiAnalysis';
import { NotFoundError } from '@/config/error';
import { queryFilter, getPagination, getTimeRangeFilter } from '@/db/utils/query';

export class HotspotService {
    /**
     * 获取热点列表
     */
    static async getHotspots({
        page = 1,
        limit = 20,
        source,
        importance,
        keywordId,
        isReal,
        timeRange,
        timeFrom,
        timeTo,
        sortBy = 'createdAt',
        sortOrder = 'desc'
    }: {
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
    }) {
        const { pageNum, limitNum, offset } = getPagination(page, limit);

        const { finalTimeFrom, finalTimeTo } = getTimeRangeFilter({ timeRange, timeFrom, timeTo });

        const filters = queryFilter(
            {
                source: (val: string) => eq(hotspotsTable.source, val),
                importance: (val: string) => eq(hotspotsTable.importance, val),
                keywordId: (val: string) => eq(hotspotsTable.keywordId, val),
                isReal: (val: string) => eq(hotspotsTable.isReal, String(val) === 'true'),
                finalTimeFrom: (val: Date) => gte(hotspotsTable.createdAt, val),
                finalTimeTo: (val: Date) => lte(hotspotsTable.createdAt, val)
            },
            { source, importance, keywordId, isReal, finalTimeFrom, finalTimeTo }
        );

        const where = filters.length > 0 ? and(...filters) : undefined;

        // 排序逻辑
        let orderBy: SQL[] = [];
        const needsMemorySort = sortBy === 'importance' || sortBy === 'hot';

        if (!needsMemorySort) {
            const orderFn = sortOrder === 'asc' ? asc : desc;

            if (sortBy === 'publishedAt') {
                orderBy = [orderFn(hotspotsTable.publishedAt), desc(hotspotsTable.createdAt)];
            } else if (sortBy === 'relevance') {
                orderBy = [orderFn(hotspotsTable.relevance)];
            } else {
                orderBy = [orderFn(hotspotsTable.createdAt)];
            }
        }

        // 执行查询
        const fetchAll = async () => {
            return await db.query.hotspots.findMany({
                where,
                orderBy,
                ...(needsMemorySort ? {} : { limit: limitNum, offset }),
                with: {
                    keyword: {
                        columns: { id: true, text: true, category: true }
                    }
                }
            });
        };

        const countTotal = async () => {
            const result = await db
                .select({ count: sql<number>`count(*)` })
                .from(hotspotsTable)
                .where(where);

            return result[0]?.count || 0;
        };

        const [rawHotspots, total] = await Promise.all([fetchAll(), countTotal()]);

        let hotspots = rawHotspots;

        if (needsMemorySort) {
            const sorted = sortHotspots(rawHotspots, sortBy, sortOrder);

            hotspots = sorted.slice(offset, offset + limitNum);
        }

        return {
            data: hotspots,
            page: pageNum,
            limit: limitNum,
            total: total
        };
    }

    /**
     * 获取统计数据
     */
    static async getStatus() {
        const today = new Date();

        today.setHours(0, 0, 0, 0);

        const [total, todayCount, urgentCount, sourceData] = await Promise.all([
            db.select({ count: sql<number>`count(*)` }).from(hotspotsTable),
            db
                .select({ count: sql<number>`count(*)` })
                .from(hotspotsTable)
                .where(gte(hotspotsTable.createdAt, today)),
            db
                .select({ count: sql<number>`count(*)` })
                .from(hotspotsTable)
                .where(eq(hotspotsTable.importance, 'urgent')),
            db
                .select({ source: hotspotsTable.source, count: sql<number>`count(*)` })
                .from(hotspotsTable)
                .groupBy(hotspotsTable.source)
        ]);

        return {
            total: total[0]?.count || 0,
            today: todayCount[0]?.count || 0,
            urgent: urgentCount[0]?.count || 0,
            bySource: sourceData.reduce(
                (acc, item) => {
                    acc[item.source] = item.count;

                    return acc;
                },
                {} as Record<string, number>
            )
        };
    }

    /**
     * 获取单个热点
     */
    static async getById(id: string) {
        const hotspot = await db.query.hotspots.findFirst({
            where: eq(hotspotsTable.id, id),
            with: { keyword: true }
        });

        if (!hotspot) {
            throw new NotFoundError('热点未找到');
        }

        return hotspot;
    }

    /**
     * 删除热点
     */
    static async delete(id: string) {
        const result = await db.delete(hotspotsTable).where(eq(hotspotsTable.id, id)).returning();

        if (result.length === 0) {
            throw new NotFoundError('热点未找到');
        }

        return true;
    }

    /**
     * 实时跨平台搜索热点
     */
    static async searchExternal(query: string, sources: string[] = ['twitter', 'bing']) {
        const results: any[] = [];

        // Twitter 搜索
        if (sources.includes('twitter')) {
            try {
                const tweets = await searchTwitter(query);

                results.push(...tweets);
            } catch (error) {
                console.error('Twitter search failed:', error);
            }
        }

        // Bing 搜索
        if (sources.includes('bing')) {
            try {
                const webResults = await searchBing(query);

                results.push(...webResults);
            } catch (error) {
                console.error('Bing search failed:', error);
            }
        }

        // AI 分析前几个结果 (减少至前 5 个以便节省额度)
        const itemsToAnalyze = results.slice(0, 5);
        const analysisList = await batchAnalyze(
            itemsToAnalyze.map(item => item.title + ' ' + item.content),
            query
        );
        // results.slice(0, 5).map(async item => {
        //     try {
        //         const analysis = await analyzeContent(item.title + ' ' + item.content, query);

        //         return { ...item, analysis };
        //     } catch {
        //         return { ...item, analysis: null };
        //     }
        // })

        return itemsToAnalyze.map((item, index) => ({
            ...item,
            analysis: analysisList[index] || null
        }));
    }
}
