export interface PlatformQualityConfig {
    minLikes: number;
    minEngagement: number; // 综合互动数 (likes + retweets/comments)
    minViews: number;
    minFollowers: number;
    weightEngagement: number;
    weightAuthority: number;
    weightRecency: number;
}

export const DEFAULT_QUALITY_CONFIG: Record<string, PlatformQualityConfig> = {
    twitter: {
        minLikes: 10,
        minEngagement: 15,
        minViews: 500,
        minFollowers: 100,
        weightEngagement: 0.6,
        weightAuthority: 0.3,
        weightRecency: 0.1
    },
    bilibili: {
        minLikes: 50,
        minEngagement: 100, // 点赞 + 投币 + 收藏 + 评论
        minViews: 1000,
        minFollowers: 500,
        weightEngagement: 0.5,
        weightAuthority: 0.4,
        weightRecency: 0.1
    },
    hackernews: {
        minLikes: 5, // Points
        minEngagement: 2, // Comments
        minViews: 0,
        minFollowers: 0,
        weightEngagement: 0.8,
        weightAuthority: 0.1,
        weightRecency: 0.1
    },
    general: {
        minLikes: 0,
        minEngagement: 0,
        minViews: 0,
        minFollowers: 0,
        weightEngagement: 0.5,
        weightAuthority: 0.3,
        weightRecency: 0.2
    }
};
