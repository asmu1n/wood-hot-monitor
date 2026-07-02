interface Keyword {
    id: string;
    text: string;
    category: string | null;
    isActive: boolean;
    createdAt: string;
    updatedAt: string;
    hotspotCount?: number;
}

interface Hotspot {
    id: string;
    title: string;
    content: string;
    url: string;
    source: string;
    sourceId: string | null;
    isReal: boolean;
    relevance: number;
    relevanceReason: string | null;
    keywordMentioned: boolean | null;
    importance: 'low' | 'medium' | 'high' | 'urgent';
    summary: string | null;
    viewCount: number | null;
    likeCount: number | null;
    retweetCount: number | null;
    replyCount: number | null;
    commentCount: number | null;
    quoteCount: number | null;
    danmakuCount: number | null;
    authorName: string | null;
    authorUsername: string | null;
    authorAvatar: string | null;
    authorFollowers: number | null;
    authorVerified: boolean | null;
    publishedAt: string | null;
    createdAt: string;
    keyword: { id: string; text: string; category: string | null } | null;
    isNotified: boolean;
    notifiedAt: string | null;
    isRead: boolean;
}

interface Status {
    total: number;
    today: number;
    urgent: number;
    bySource: Record<string, number>;
}

interface AppConfig {
    emailAddress: string;
    llmProvider: string;
    llmModel: string;
    llmApiKey: string;
    llmBaseUrl: string;
    checkInterval: number;
    settings: Record<string, any>;
}

export { Keyword, Hotspot, Status, AppConfig };
