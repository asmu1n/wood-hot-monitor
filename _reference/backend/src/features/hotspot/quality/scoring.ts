import { DEFAULT_QUALITY_CONFIG } from './constants';

export interface ScoreSnapshot {
    engagement: number;
    authority: number;
    recency: number;
    isBlueVerified: boolean;
    total: number;
}

export interface MetricInput {
    platform: string;
    likes: number; // 点赞 / 喜欢 / 赞同 / HN Points
    shares: number; // 转发 / 分享 / 引用 (Twitter Retweet+Quote, Bili Share)
    comments: number; // 评论 / 回复 / 弹幕 (Twitter Reply, Bili Comment+Danmaku)
    views: number; // 阅读 / 播放
    followers: number; // 作者粉丝数
    isVerified: boolean; // 是否认证 (包含蓝V)
    publishedAt?: Date | string;
}

export class QualityScorer {
    /**
     * 计算单平台综合评分
     * @param input 原始指标
     */
    static calculateScore(input: MetricInput) {
        const config = DEFAULT_QUALITY_CONFIG[input.platform] || DEFAULT_QUALITY_CONFIG.general;

        // 1. 参与度分数 (Engagement)
        // 归一化处理：根据平台特性的权重计算聚合指标
        const engagementVal = this.getNormalizedEngagement(input);

        const engagementScore = Math.min(100, (engagementVal / (config.minEngagement || 10)) * 50);

        // 2. 权威度分数 (Authority)
        const followers = input.followers || 0;

        let authorityScore = Math.min(100, (followers / (config.minFollowers || 500)) * 50);

        if (input.isVerified) {
            authorityScore = Math.min(100, authorityScore + 20);
        }

        // 3. 时效性分数 (Recency)
        let recencyScore = 100;

        if (input.publishedAt) {
            const pubDate = new Date(input.publishedAt);
            const hoursDiff = (Date.now() - pubDate.getTime()) / (1000 * 3600);

            // 衰减系数: 每 24 小时衰减一些，超过 7 天基本归零
            recencyScore = Math.max(0, 100 - hoursDiff * 0.6);
        }

        // 4. 总分计算
        const total = engagementScore * config.weightEngagement + authorityScore * config.weightAuthority + recencyScore * config.weightRecency;

        const score: ScoreSnapshot = {
            engagement: engagementScore,
            authority: authorityScore,
            recency: recencyScore,
            isBlueVerified: input.platform === 'twitter' && input.isVerified,
            total: Math.round(total * 10) / 10
        };

        return score;
    }

    /**
     * 判断是否通过基本质量过滤
     */
    static shouldFilter(input: MetricInput) {
        const config = DEFAULT_QUALITY_CONFIG[input.platform] || DEFAULT_QUALITY_CONFIG.general;

        // 蓝V / 认证 宽限
        const factor = input.isVerified ? 0.5 : 1.0;

        if ((input.likes || 0) < config.minLikes * factor) {
            return { pass: false, reason: 'low_likes' };
        }

        const engagement = this.getNormalizedEngagement(input);

        if (engagement < config.minEngagement * factor) {
            return { pass: false, reason: 'low_engagement' };
        }

        if ((input.views || 0) < config.minViews * factor && config.minViews > 0) {
            return { pass: false, reason: 'low_views' };
        }

        return { pass: true };
    }

    private static getNormalizedEngagement(input: MetricInput): number {
        const { platform, likes, shares, comments } = input;

        switch (platform) {
            case 'twitter':
                // Twitter: 转发最重，回复次之
                return likes + shares * 2 + comments * 1.5;
            case 'bilibili':
                // Bilibili: 分享（转发）最重
                return likes + shares * 3 + comments * 2;
            case 'hackernews':
                // HN: Points 是基础，评论增加讨论权重
                return likes + comments * 2;
            default:
                return likes + shares + comments;
        }
    }
}
