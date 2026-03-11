import { SILICONFLOW_API_KEY } from '@env';
import { ConcurrencyLimiter } from '../utils/limiter';
import { fetchSiliconFlow } from '../utils/llm';
import { buildAnalysisPrompt, buildExpandKeywordPrompt } from '../utils/prompt';
import db from '@/config/database';
import { keywordExpansions } from '@/models/keywordExpansions';
import { eq } from 'drizzle-orm';

interface AIAnalysis {
    isReal: boolean;
    relevance: number;
    relevanceReason: string; // AI 判断相关性的理由
    keywordMentioned: boolean; // 内容中是否直接提及了关键词或其核心概念
    importance: 'low' | 'medium' | 'high' | 'urgent';
    summary: string; // 与关键词的关联说明（不是单纯的内容介绍）
}

// ========== Rate Limiter (频率限制器) ==========

// 硅基流动支持较高的并发，这里暂定最大并发为 10（可按需调整）
const aiLimiter = new ConcurrencyLimiter(10);

// ========== Query Expansion（查询扩展） ==========

/**
 * 使用 AI 将关键词扩展为多个变体，用于文本预过滤。
 * 返回扩展后的关键词列表（含原始关键词）。
 * 结果会被缓存，同一关键词不会重复调用 AI。
 */

export async function expandKeyword(keyword: string): Promise<string[]> {
    // 先从数据库读取已存储的扩展词
    const rows = await db.select({ expansion: keywordExpansions.expansion }).from(keywordExpansions).where(eq(keywordExpansions.keyword, keyword));

    if (rows.length > 0) {
        // 已有记录，返回包括原始关键词在内的数组
        const stored = rows.map(r => r.expansion);

        return [keyword, ...new Set(stored)];
    }

    // 不管 AI 是否可用，先提取基础核心词
    const coreTerms = extractCoreTerms(keyword);

    if (!SILICONFLOW_API_KEY) {
        const result = [keyword, ...coreTerms];

        // 保存到数据库
        await db.insert(keywordExpansions).values(result.slice(1).map(exp => ({ keyword, expansion: exp })));

        return result;
    }

    try {
        const expanded = await aiLimiter.run(async () => {
            const prompt = buildExpandKeywordPrompt(keyword);
            const responseText = await fetchSiliconFlow(prompt);
            const jsonMatch = responseText.match(/\[[\s\S]*\]/);

            if (jsonMatch) {
                const parsed: string[] = JSON.parse(jsonMatch[0]);

                return [...new Set([keyword, ...coreTerms, ...parsed.map(s => s.trim()).filter(Boolean)])];
            }

            return [keyword, ...coreTerms];
        });
        // 将新生成的扩展词写入数据库（去除原始关键词本身）
        const toInsert = expanded.slice(1).filter(exp => !coreTerms.includes(exp));

        if (toInsert.length > 0) {
            await db.insert(keywordExpansions).values(toInsert.map(exp => ({ keyword, expansion: exp })));
        }

        console.log(`  🔍 Query expansion for "${keyword}": ${expanded.length} variants`);

        return expanded;
    } catch (error) {
        console.error('Query expansion failed:', error);
    }

    // Fallback：使用基础核心词并写入数据库
    const fallback = [keyword, ...coreTerms];

    await db.insert(keywordExpansions).values(fallback.slice(1).map(exp => ({ keyword, expansion: exp })));

    return fallback;
}

/**
 * 从关键词中提取核心词（纯文本方式，不依赖 AI）
 */
function extractCoreTerms(keyword: string): string[] {
    const terms: string[] = [];
    // 按空格、连字符、下划线分割
    const parts = keyword.split(/[\s\-_/\\·]+/).filter(p => p.length >= 2);

    if (parts.length > 1) {
        terms.push(...parts);

        // 两两组合
        for (let i = 0; i < parts.length - 1; i++) {
            terms.push(parts[i] + ' ' + parts[i + 1]);
        }
    }

    // 去重，排除原始关键词本身
    return [...new Set(terms)].filter(t => t.toLowerCase() !== keyword.toLowerCase());
}

// ========== 关键词预匹配 ==========

/**
 * 检查文本中是否包含任一扩展关键词（不区分大小写）。
 * 返回是否匹配以及匹配到的词。
 */
export function preMatchKeyword(text: string, expandedKeywords: string[]): { matched: boolean; matchedTerms: string[] } {
    const lowerText = text.toLowerCase();
    const matchedTerms: string[] = [];

    for (const kw of expandedKeywords) {
        if (lowerText.includes(kw.toLowerCase())) {
            matchedTerms.push(kw);
        }
    }

    return { matched: matchedTerms.length > 0, matchedTerms };
}

// ========== AI 内容分析（关键词感知） ==========

export async function analyzeContent(
    content: string,
    keyword: string,
    preMatchResult?: { matched: boolean; matchedTerms: string[] }
): Promise<AIAnalysis> {
    // 默认预匹配结果
    const matchResult = preMatchResult ?? { matched: false, matchedTerms: [] };

    if (!SILICONFLOW_API_KEY) {
        console.warn('SiliconFlow API key not configured, using fallback analysis');

        return {
            isReal: true,
            relevance: matchResult.matched ? 50 : 20,
            relevanceReason: '未配置 AI 服务，使用默认分数',
            keywordMentioned: matchResult.matched,
            importance: 'low',
            summary: content.slice(0, 50) + '...'
        };
    }

    try {
        return await aiLimiter.run(async () => {
            const systemPrompt = buildAnalysisPrompt(keyword, matchResult);

            const promptContent = `System Instruction: ${systemPrompt}\n\nContent to analyze:\n${content.slice(0, 5000)}`;
            const responseText = await fetchSiliconFlow(promptContent);

            // 尝试解析 JSON
            const jsonMatch = responseText.match(/\{[\s\S]*\}/);

            if (jsonMatch) {
                const parsed = JSON.parse(jsonMatch[0]);

                return {
                    isReal: Boolean(parsed.isReal),
                    relevance: Math.min(100, Math.max(0, Number(parsed.relevance) || 0)),
                    relevanceReason: String(parsed.relevanceReason || '').slice(0, 200),
                    keywordMentioned: Boolean(parsed.keywordMentioned),
                    importance: ['low', 'medium', 'high', 'urgent'].includes(parsed.importance) ? parsed.importance : 'low',
                    summary: String(parsed.summary || '').slice(0, 150)
                };
            }

            throw new Error('Failed to parse AI response');
        });
    } catch (error) {
        console.error('AI analysis failed:', error);

        // Fallback
        return {
            isReal: true,
            relevance: matchResult.matched ? 30 : 10,
            relevanceReason: 'AI 分析失败，使用默认分数',
            keywordMentioned: matchResult.matched,
            importance: 'low',
            summary: content.slice(0, 50) + '...'
        };
    }
}

export async function batchAnalyze(contents: string[], keyword: string, expandedKeywords?: string[]): Promise<AIAnalysis[]> {
    // 配合新的并发限制器，这里可以适当放大 batchSize（AI limiter 会自己在底层控制全局并发）
    const batchSize = 10;
    const results: AIAnalysis[] = [];

    for (let i = 0; i < contents.length; i += batchSize) {
        const batch = contents.slice(i, i + batchSize);
        const batchResults = await Promise.all(
            batch.map(content => {
                const preMatch = expandedKeywords ? preMatchKeyword(content, expandedKeywords) : undefined;

                return analyzeContent(content, keyword, preMatch);
            })
        );

        results.push(...batchResults);
    }

    return results;
}
