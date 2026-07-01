import { sqliteTable, text, integer, uniqueIndex } from 'drizzle-orm/sqlite-core';
import { relations, sql } from 'drizzle-orm';
import { keywords } from './keywords';

export const hotspots = sqliteTable(
    'hotspots',
    {
        id: text('id')
            .primaryKey()
            .$defaultFn(() => crypto.randomUUID()),
        title: text('title').notNull(),
        content: text('content').notNull(),
        url: text('url').notNull(),
        source: text('source').notNull(), // twitter, bing, google
        sourceId: text('source_id'), // 原始推文ID等
        isReal: integer('is_real', { mode: 'boolean' }).notNull().default(true),
        relevance: integer('relevance').notNull().default(0),
        relevanceReason: text('relevance_reason'), // AI 分析相关性的理由
        keywordMentioned: integer('keyword_mentioned', { mode: 'boolean' }), // 内容中是否直接提及了关键词
        importance: text('importance', { enum: ['low', 'medium', 'high', 'urgent'] })
            .notNull()
            .default('low'),
        summary: text('summary'),
        viewCount: integer('view_count'),
        likeCount: integer('like_count'),
        retweetCount: integer('retweet_count'),
        replyCount: integer('reply_count'), // 回复数
        commentCount: integer('comment_count'), // 评论数
        quoteCount: integer('quote_count'), // 引用/转引数
        danmakuCount: integer('danmaku_count'), // 弹幕数 (Bilibili)
        authorName: text('author_name'), // 作者名称
        authorUsername: text('author_username'), // 作者用户名
        authorAvatar: text('author_avatar'), // 作者头像
        authorFollowers: integer('author_followers'), // 作者粉丝数
        authorVerified: integer('author_verified', { mode: 'boolean' }), // 作者是否认证
        publishedAt: integer('published_at', { mode: 'timestamp' }),
        createdAt: integer('created_at', { mode: 'timestamp' })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        keywordId: text('keyword_id').references(() => keywords.id, { onDelete: 'set null' })
    },
    table => [uniqueIndex('url_source_idx').on(table.url, table.source)]
);

export const hotspotsRelations = relations(hotspots, ({ one }) => ({
    keyword: one(keywords, {
        fields: [hotspots.keywordId],
        references: [keywords.id]
    })
}));
