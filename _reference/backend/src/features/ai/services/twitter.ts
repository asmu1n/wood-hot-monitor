import { TWITTER_CONFIG } from '@env';
import { QualityManager } from '../../../features/hotspot/quality/manager';

interface Tweet {
    type: string;
    id: string;
    url: string;
    text: string;
    retweetCount: number;
    replyCount: number;
    likeCount: number;
    quoteCount: number;
    viewCount: number;
    createdAt: string;
    lang: string;
    author: {
        userName: string;
        name: string;
        isBlueVerified: boolean;
        profilePicture: string;
        followers: number;
    };
}

const TWITTER_API_BASE = 'https://api.twitterapi.io';

// ============================================================
// 质量过滤 & 排序
//   1. 排除回复推文（type 包含 reply 或以 @ 开头）
//   2. 调用通用质量分析引擎进行多维过滤与评分
// ============================================================
function filterAndRankTweets(tweets: Tweet[]) {
    // 基础过滤：排除回复推文
    const baseFiltered = tweets.filter(tweet => {
        if (tweet.type && tweet.type.toLowerCase().includes('reply')) {
            return false;
        }

        if (/^@\w+\s/.test(tweet.text.trim())) {
            return false;
        }

        return true;
    });

    // 转换为通用指标输入
    const metricInputs = baseFiltered.map(tweet => ({
        platform: 'twitter',
        likes: tweet.likeCount,
        shares: tweet.retweetCount + (tweet.quoteCount || 0),
        comments: tweet.replyCount,
        views: tweet.viewCount,
        followers: tweet.author.followers,
        isVerified: tweet.author.isBlueVerified, // 将蓝V视为已认证
        publishedAt: tweet.createdAt,
        _original: tweet
    }));

    // 调用通用质量管理器进行多维评分和过滤
    const ranked = QualityManager.process(metricInputs);

    // 映射回搜索结果格式
    return ranked.map(item => {
        const tweet = item._original;

        return {
            title: tweet.text.slice(0, 100),
            content: tweet.text,
            url: tweet.url,
            source: 'twitter' as const,
            sourceId: tweet.id,
            publishedAt: new Date(tweet.createdAt),
            viewCount: tweet.viewCount,
            likeCount: tweet.likeCount,
            retweetCount: tweet.retweetCount,
            replyCount: tweet.replyCount,
            quoteCount: tweet.quoteCount,
            score: item.finalScore, // 暴露质量分
            author: {
                name: tweet.author.name,
                username: tweet.author.userName,
                avatar: tweet.author.profilePicture,
                followers: tweet.author.followers,
                verified: tweet.author.isBlueVerified
            }
        };
    });
}

// ============================================================
// 日期工具
// ============================================================
function formatSinceDate(daysAgo: number) {
    const d = new Date(Date.now() - daysAgo * 86400000);

    return `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, '0')}-${String(d.getUTCDate()).padStart(2, '0')}`;
}

// ============================================================
// 构建高级搜索 query
//   - Top 搜索：近 7 天，min_faves:10，排除 RT 和纯回复
//   - Latest 搜索：近 3 天，排除 RT 和纯回复
// 参考语法：https://github.com/igorbrigadir/twitter-advanced-search
// ============================================================
function buildAdvancedQuery(keyword: string, type: 'Top' | 'Latest') {
    const parts: string[] = [keyword];

    // 排除 RT 和纯回复
    parts.push('-filter:retweets');
    parts.push('-filter:replies');

    // 时间范围：Top 看 7 天，Latest 看 3 天
    const daysAgo = type === 'Top' ? 7 : 3;

    parts.push(`since:${formatSinceDate(daysAgo)}`);

    // Top 搜索额外加 min_faves 保证质量
    if (type === 'Top') {
        parts.push('min_faves:10');
    }

    return parts.join(' ');
}

// ============================================================
// HTTP 请求
// ============================================================
async function makeTwitterRequest(endpoint: string, params: Record<string, string> = {}) {
    const apiKey = TWITTER_CONFIG.apiKey;

    if (!apiKey) {
        console.warn('Twitter API key not configured');

        return { tweets: [] };
    }

    const url = new URL(`${TWITTER_API_BASE}${endpoint}`);

    Object.entries(params).forEach(([key, value]) => {
        url.searchParams.append(key, value);
    });

    const response = await fetch(url.toString(), {
        headers: {
            'X-API-Key': apiKey,
            'Content-Type': 'application/json'
        }
    });

    if (!response.ok) {
        throw new Error(`Twitter API error: ${response.status} ${response.statusText}`);
    }

    return response.json();
}

// ============================================================
// 获取单页推文（支持分页）
// ============================================================
async function fetchTweetPage(query: string, queryType: 'Top' | 'Latest', cursor?: string): Promise<{ tweets: Tweet[]; nextCursor?: string }> {
    const data = await makeTwitterRequest('/twitter/tweet/advanced_search', {
        query,
        queryType,
        ...(cursor ? { cursor } : {})
    });

    return {
        tweets: data.tweets && Array.isArray(data.tweets) ? data.tweets : [],
        nextCursor: data.has_next_page ? data.next_cursor : undefined
    };
}

// ============================================================
// 主搜索函数
//   Top: 拉取 2 页（≤40 条高质量热门推文）
//   Latest: 拉取 1 页（≤20 条最新推文）
// ============================================================
export async function searchTwitter(query: string) {
    try {
        const topQuery = buildAdvancedQuery(query, 'Top');
        const latestQuery = buildAdvancedQuery(query, 'Latest');

        console.log(`Twitter advanced queries:\n  Top: ${topQuery}\n  Latest: ${latestQuery}`);

        // 第 1 批：Top 第 1 页 + Latest 第 1 页（并行）
        const [topPage1, latestPage1] = await Promise.allSettled([fetchTweetPage(topQuery, 'Top'), fetchTweetPage(latestQuery, 'Latest')]);

        const allTweets: Tweet[] = [];
        const seenIds = new Set<string>();

        const addTweets = (tweets: Tweet[]) => {
            for (const tweet of tweets) {
                if (!seenIds.has(tweet.id)) {
                    seenIds.add(tweet.id);
                    allTweets.push(tweet);
                }
            }
        };

        let topNextCursor: string | undefined;

        if (topPage1.status === 'fulfilled') {
            addTweets(topPage1.value.tweets);
            topNextCursor = topPage1.value.nextCursor;
        }

        if (latestPage1.status === 'fulfilled') {
            addTweets(latestPage1.value.tweets);
        }

        // 第 2 批：如果 Top 有下一页，再拉一页（多拿一些热门内容）
        if (topNextCursor) {
            try {
                const topPage2 = await fetchTweetPage(topQuery, 'Top', topNextCursor);

                addTweets(topPage2.tweets);
            } catch (e) {
                console.warn('Twitter Top page 2 failed:', e);
            }
        }

        console.log(`Twitter: ${allTweets.length} unique tweets fetched (Top 2 pages + Latest 1 page)`);

        // 本地质量过滤 & 排序
        const qualityResults: SearchResult[] = filterAndRankTweets(allTweets);

        console.log(`Twitter: ${allTweets.length} → ${qualityResults.length} after quality filter (using unified QualityManager)`);

        return qualityResults;
    } catch (error) {
        console.error('Twitter search error:', error);

        return [];
    }
}

export async function getTrends(woeid: number = 1): Promise<any[]> {
    try {
        const data = await makeTwitterRequest('/twitter/trends', { woeid: String(woeid) });

        return data.trends || [];
    } catch (error) {
        console.error('Error fetching trends:', error);

        return [];
    }
}

export async function getUserTweets(username: string): Promise<SearchResult[]> {
    try {
        const data = await makeTwitterRequest('/twitter/user/last_tweets', {
            userName: username
        });

        if (!data.tweets || !Array.isArray(data.tweets)) {
            return [];
        }

        return data.tweets.map((tweet: Tweet) => ({
            title: tweet.text.slice(0, 100),
            content: tweet.text,
            url: tweet.url,
            source: 'twitter' as const,
            sourceId: tweet.id,
            publishedAt: new Date(tweet.createdAt),
            viewCount: tweet.viewCount,
            likeCount: tweet.likeCount,
            retweetCount: tweet.retweetCount,
            replyCount: tweet.replyCount,
            quoteCount: tweet.quoteCount,
            author: {
                name: tweet.author.name,
                username: tweet.author.userName,
                avatar: tweet.author.profilePicture,
                followers: tweet.author.followers,
                verified: tweet.author.isBlueVerified
            }
        }));
    } catch (error) {
        console.error('Error fetching user tweets:', error);

        return [];
    }
}
