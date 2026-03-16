import axios from 'axios';
import * as cheerio from 'cheerio';
import { RateLimiter } from '../utils/limiter';
import { getRandomUserAgent } from '../utils/http';
import type { Search } from 'types';

const bingLimiter = new RateLimiter(5000);
const googleLimiter = new RateLimiter(10000);
const duckduckgoLimiter = new RateLimiter(3000);
const hackernewsLimiter = new RateLimiter(1000); // HN API 更宽松

const searchBing: Search = async query => {
    await bingLimiter.wait();

    try {
        const response = await axios.get('https://www.bing.com/search', {
            params: {
                q: query,
                count: 20
            },
            headers: {
                'User-Agent': getRandomUserAgent(),
                Accept: 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
                'Accept-Language': 'en-US,en;q=0.5',
                'Accept-Encoding': 'gzip, deflate, br'
            },
            timeout: 15000
        });

        const $ = cheerio.load(response.data);
        const results: SearchResult[] = [];

        $('li.b_algo').each((_, element) => {
            const titleElement = $(element).find('h2 a');
            const title = titleElement.text().trim();
            const url = titleElement.attr('href');
            const snippet = $(element).find('.b_caption p').text().trim();

            if (title && url && url.startsWith('http')) {
                results.push({
                    title,
                    content: snippet,
                    url,
                    source: 'bing'
                });
            }
        });

        console.log(`Bing search for "${query}": found ${results.length} results`);

        return results;
    } catch (error) {
        console.error('Bing search error:', error);

        return [];
    }
};

const searchGoogle: Search = async query => {
    await googleLimiter.wait();

    try {
        const response = await axios.get('https://www.google.com/search', {
            params: {
                q: query,
                num: 20,
                hl: 'en'
            },
            headers: {
                'User-Agent': getRandomUserAgent(),
                Accept: 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
                'Accept-Language': 'en-US,en;q=0.5'
            },
            timeout: 15000
        });

        const $ = cheerio.load(response.data);
        const results: SearchResult[] = [];

        $('div.g').each((_, element) => {
            const titleElement = $(element).find('h3').first();
            const title = titleElement.text().trim();
            const linkElement = $(element).find('a').first();
            const url = linkElement.attr('href');
            const snippet = $(element).find('.VwiC3b').text().trim();

            if (title && url && url.startsWith('http')) {
                results.push({
                    title,
                    content: snippet,
                    url,
                    source: 'google'
                });
            }
        });

        console.log(`Google search for "${query}": found ${results.length} results`);

        return results;
    } catch (error) {
        console.error('Google search error:', error);

        return [];
    }
};

// DuckDuckGo 搜索（使用 HTML 版本）
const searchDuckDuckGo: Search = async query => {
    await duckduckgoLimiter.wait();

    try {
        const response = await axios.get('https://html.duckduckgo.com/html/', {
            params: {
                q: query
            },
            headers: {
                'User-Agent': getRandomUserAgent(),
                Accept: 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
                'Accept-Language': 'en-US,en;q=0.5'
            },
            timeout: 15000
        });

        const $ = cheerio.load(response.data);
        const results: SearchResult[] = [];

        $('.result').each((_, element) => {
            const titleElement = $(element).find('.result__title a');
            const title = titleElement.text().trim();
            const rawUrl = titleElement.attr('href');
            const snippet = $(element).find('.result__snippet').text().trim();

            // DuckDuckGo 使用重定向 URL，需要提取实际 URL
            let url = rawUrl;

            if (rawUrl && rawUrl.includes('uddg=')) {
                try {
                    const urlParams = new URLSearchParams(rawUrl.split('?')[1]);

                    url = decodeURIComponent(urlParams.get('uddg') || rawUrl);
                } catch {
                    url = rawUrl;
                }
            }

            if (title && url && url.startsWith('http')) {
                results.push({
                    title,
                    content: snippet,
                    url,
                    source: 'duckduckgo'
                });
            }
        });

        console.log(`DuckDuckGo search for "${query}": found ${results.length} results`);

        return results;
    } catch (error) {
        console.error('DuckDuckGo search error:', error);

        return [];
    }
};

// Hacker News API（官方免费 API）
interface HNSearchResult {
    hits: Array<{
        objectID: string;
        title: string;
        url: string | null;
        story_text: string | null;
        author: string;
        points: number;
        num_comments: number;
        created_at: string;
    }>;
}

const searchHackerNews: Search = async query => {
    await hackernewsLimiter.wait();

    try {
        // 使用 Algolia 提供的 HN 搜索 API
        const oneDayAgo = Math.floor((Date.now() - 24 * 3600 * 1000) / 1000);
        const response = await axios.get<HNSearchResult>('https://hn.algolia.com/api/v1/search', {
            params: {
                query: query,
                tags: 'story', // 只搜索故事，排除评论
                hitsPerPage: 20,
                numericFilters: `created_at_i>${oneDayAgo}` // 只搜最近24小时
            },
            timeout: 15000
        });

        const results: SearchResult[] = response.data.hits
            .filter(hit => hit.url || hit.story_text) // 确保有内容
            .map(hit => ({
                title: hit.title,
                content: hit.story_text || hit.title,
                url: hit.url || `https://news.ycombinator.com/item?id=${hit.objectID}`,
                source: 'hackernews' as const,
                sourceId: hit.objectID,
                publishedAt: new Date(hit.created_at),
                score: hit.points,
                commentCount: hit.num_comments,
                author: {
                    name: hit.author,
                    username: hit.author
                }
            }));

        console.log(`Hacker News search for "${query}": found ${results.length} results`);

        return results;
    } catch (error) {
        console.error('Hacker News search error:', error);

        return [];
    }
};

// 聚合搜索（国际搜索引擎，仅保留可用的）
const searchAll: Search = async query => {
    const results = await Promise.allSettled([searchBing(query), searchHackerNews(query), searchDuckDuckGo(query), searchGoogle(query)]);

    const allResults: SearchResult[] = [];
    const sourceNames = ['Bing', 'HackerNews', 'DuckDuckGo', 'Google'];

    results.forEach((result, index) => {
        if (result.status === 'fulfilled') {
            allResults.push(...result.value);
        } else {
            console.warn(`${sourceNames[index]} search failed:`, result.reason);
        }
    });

    const uniqueResults = deduplicateResults(allResults);

    console.log(`Search aggregation for "${query}": ${allResults.length} total, ${uniqueResults.length} unique`);

    return uniqueResults;
};

// 去重工具函数
function deduplicateResults(allResults: SearchResult[]): SearchResult[] {
    const uniqueUrls = new Set<string>();

    return allResults.filter(item => {
        // 标准化 URL 用于去重
        const normalizedUrl = item.url.replace(/\/$/, '').replace(/^https?:\/\/www\./, 'https://');

        if (uniqueUrls.has(normalizedUrl)) {
            return false;
        }

        uniqueUrls.add(normalizedUrl);

        return true;
    });
}

export { searchBing, searchGoogle, searchDuckDuckGo, searchHackerNews, searchAll, deduplicateResults };
