import { sqliteTable, text, integer } from 'drizzle-orm/sqlite-core';
import { sql } from 'drizzle-orm';

export const notifications = sqliteTable('notifications', {
    id: text('id')
        .primaryKey()
        .$defaultFn(() => crypto.randomUUID()),
    type: text('type').notNull(), // hotspot, alert
    title: text('title').notNull(),
    content: text('content').notNull(),
    isRead: integer('is_read', { mode: 'boolean' }).notNull().default(false),
    hotSpotId: text('hotspot_id'),
    createdAt: integer('created_at', { mode: 'timestamp' })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`)
});
