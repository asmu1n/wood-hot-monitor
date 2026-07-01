import { sqliteTable, text, integer } from 'drizzle-orm/sqlite-core';

export const keywordExpansions = sqliteTable('keyword_expansions', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    keyword: text('keyword').notNull(), // 原始关键词
    expansion: text('expansion').notNull() // 对应的扩展词条
});
