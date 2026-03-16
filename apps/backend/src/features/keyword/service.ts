import db from '@/config/database';
import { keywords as keywordsTable } from '@/models/keywords';
import { hotspots as hotspotsTable } from '@/models/hotspots';
import { eq, sql, desc } from 'drizzle-orm';
import { NotFoundError, ConflictError } from '@/config/error';
import type { Keyword } from '@repo/types';

export class KeywordService {
    /**
     * 获取所有关键词及热点计数
     */
    static async getAll(): Promise<Keyword[]> {
        const result = await db
            .select({
                id: keywordsTable.id,
                text: keywordsTable.text,
                category: keywordsTable.category,
                isActive: keywordsTable.isActive,
                createdAt: keywordsTable.createdAt,
                updatedAt: keywordsTable.updatedAt,
                hotspotsCount: sql<number>`count(${hotspotsTable.id})`.mapWith(Number)
            })
            .from(keywordsTable)
            .leftJoin(hotspotsTable, eq(keywordsTable.id, hotspotsTable.keywordId))
            .groupBy(keywordsTable.id)
            .orderBy(desc(keywordsTable.createdAt));

        return result.map(item => ({
            ...item,
            createdAt: item.createdAt.toISOString() || '',
            updatedAt: item.updatedAt.toISOString() || ''
        }));
    }

    /**
     * 获取单个关键词及其关联的热点 (前20个)
     */
    static async getById(id: string): Promise<Keyword> {
        const keyword = await db.query.keywords.findFirst({
            where: eq(keywordsTable.id, id),
            with: {
                hotspots: {
                    limit: 20,
                    orderBy: [desc(hotspotsTable.createdAt)]
                }
            }
        });

        if (!keyword) {
            throw new NotFoundError('关键词未找到');
        }

        return {
            ...keyword,
            createdAt: keyword.createdAt.toISOString() || '',
            updatedAt: keyword.updatedAt.toISOString() || ''
        };
    }

    /**
     * 创建关键词
     */
    static async create(data: { text: string; category?: string | null }) {
        if (!data.text || typeof data.text !== 'string' || data.text.trim().length === 0) {
            throw new ConflictError('关键词不能为空'); // Actually BadRequestError would be better, but we handle it in route too. Wait, let's keep BadRequestError check in route or move it here. Let's move it to Route via another edit for service side constraints.
        }

        try {
            const result = await db
                .insert(keywordsTable)
                .values({
                    text: data.text.trim(),
                    category: data.category?.trim() || null
                })
                .returning();

            return result[0];
        } catch (error: any) {
            if (error.message?.includes('UNIQUE constraint failed')) {
                throw new ConflictError('关键词已存在');
            }

            throw error;
        }
    }

    /**
     * 更新关键词
     */
    static async update(id: string, data: any): Promise<Keyword> {
        const updateData: any = {};

        if (data.text !== undefined) {
            updateData.text = data.text.trim();
        }

        if (data.category !== undefined) {
            updateData.category = data.category?.trim() || null;
        }

        if (data.isActive !== undefined) {
            updateData.isActive = data.isActive;
        }

        updateData.updatedAt = new Date();

        try {
            const result = await db.update(keywordsTable).set(updateData).where(eq(keywordsTable.id, id)).returning();

            if (result.length === 0) {
                throw new NotFoundError('关键词未找到');
            }

            const data = result[0];

            return {
                ...data,
                createdAt: data.createdAt.toISOString() || '',
                updatedAt: data.updatedAt.toISOString() || ''
            };
        } catch (error: any) {
            if (error.message?.includes('UNIQUE constraint failed')) {
                throw new ConflictError('关键词已存在');
            }

            throw error;
        }
    }

    /**
     * 删除关键词
     */
    static async delete(id: string) {
        const result = await db.delete(keywordsTable).where(eq(keywordsTable.id, id)).returning();

        if (result.length === 0) {
            throw new NotFoundError('关键词未找到');
        }

        return true;
    }

    /**
     * 切换关键词启用状态
     */
    static async toggle(id: string): Promise<Keyword> {
        const keyword = await this.getById(id); // will throw NotFoundError if not found

        const result = await db
            .update(keywordsTable)
            .set({
                isActive: !keyword.isActive,
                updatedAt: new Date()
            })
            .where(eq(keywordsTable.id, id))
            .returning();

        const data = result[0];

        return {
            ...data,
            createdAt: data.createdAt.toISOString() || '',
            updatedAt: data.updatedAt.toISOString() || ''
        };
    }
}
