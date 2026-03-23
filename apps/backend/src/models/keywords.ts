import { sqliteTable, text, integer } from 'drizzle-orm/sqlite-core';
import { relations, sql } from 'drizzle-orm';
import { hotspots } from './hotspots';

export const keywords = sqliteTable('keywords', {
    id: text('id')
        .primaryKey()
        .$defaultFn(() => crypto.randomUUID()),
    text: text('text').notNull().unique(),
    category: text('category'),
    isActive: integer('is_active', { mode: 'boolean' }).notNull().default(true),
    createdAt: integer('created_at', { mode: 'timestamp' })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
    updatedAt: integer('updated_at', { mode: 'timestamp' })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`)
});

export const keywordsRelations = relations(keywords, ({ many }) => ({
    hotspots: many(hotspots)
}));
