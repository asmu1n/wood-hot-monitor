import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    Flame,
    ExternalLink,
    ChevronDown,
    ChevronUp,
    FileText,
    User,
    Zap,
    Repeat2,
    MessageCircle,
    Quote,
    Eye,
    Target,
    Activity,
    Clock,
    Shield,
    ShieldAlert,
    ThermometerSun,
    AlertTriangle,
    X,
    Globe,
    Search,
    TrendingUp
} from 'lucide-react';
import { relativeTime, formatDateTime } from '@/utils/relativeTime';
import { calcHeatScore, getHeatLevel } from '../utils';
import { cn } from '@repo/ui';

interface HotSpotCardProps {
    hotspot: Hotspot;
    index: number;
    isExpandedReason: boolean;
    isExpandedContent: boolean;
    onToggleReason: (id: string) => void;
    onToggleContent: (id: string) => void;
}

export const getImportanceIcon = (importance: string) => {
    switch (importance) {
        case 'urgent':
            return <AlertTriangle className="h-4 w-4" />;
        case 'high':
            return <Flame className="h-4 w-4" />;
        case 'medium':
            return <Zap className="h-4 w-4" />;
        default:
            return <TrendingUp className="h-4 w-4" />;
    }
};

const getSourceIcon = (source: string) => {
    switch (source.toLowerCase()) {
        case 'twitter':
            return <X className="h-4 w-4" />;
        case 'bilibili':
            return <Eye className="h-4 w-4" />;
        case 'weibo':
            return <Activity className="h-4 w-4" />;
        case 'google':
            return <Search className="h-4 w-4" />;
        case 'hackernews':
            return <Zap className="h-4 w-4" />;
        case 'X':
            return <X className="h-4 w-4" />;
        default:
            return <Globe className="h-4 w-4" />;
    }
};

export const getSourceLabel = (source: string) => {
    const labels: Record<string, string> = {
        twitter: 'Twitter',
        bing: 'Bing',
        google: 'Google',
        sogou: '搜狗',
        bilibili: 'Bilibili',
        weibo: '微博热搜',
        hackernews: 'HackerNews',
        duckduckgo: 'DuckDuckGo'
    };

    return labels[source.toLowerCase()] || source;
};

const HotSpotCard: React.FC<HotSpotCardProps> = ({ hotspot, index, isExpandedReason, isExpandedContent, onToggleReason, onToggleContent }) => {
    const heatScore = calcHeatScore(hotspot);
    const heat = getHeatLevel(heatScore);

    return (
        <motion.div
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: index * 0.03 }}
            className="group border-border bg-card/40 hover:border-primary/20 hover:bg-card/60 rounded-2xl border p-5 shadow-sm transition-all hover:shadow-md">
            <div className="flex items-start justify-between gap-4">
                <div className="min-w-0 flex-1">
                    {/* Row 1: Meta badges */}
                    <div className="mb-3 flex flex-wrap items-center gap-2">
                        <span
                            className={cn(
                                'flex items-center rounded-lg px-2.5 py-1 text-[10px] font-semibold tracking-wider uppercase',
                                hotspot.importance === 'urgent' && 'border-destructive/20 bg-destructive/15 text-destructive border',
                                hotspot.importance === 'high' && 'border-chart-3/20 bg-chart-3/15 text-chart-3 border',
                                hotspot.importance === 'medium' && 'border-chart-5/20 bg-chart-5/15 text-chart-5 border',
                                hotspot.importance === 'low' && 'border-chart-2/20 bg-chart-2/15 text-chart-2 border'
                            )}>
                            {getImportanceIcon(hotspot.importance)}
                            <span className="ml-1">{hotspot.importance}</span>
                        </span>
                        <span className="text-muted-foreground flex items-center gap-1 text-xs">
                            {getSourceIcon(hotspot.source)}
                            {getSourceLabel(hotspot.source)}
                        </span>
                        {hotspot.keyword && (
                            <span className="border-primary/20 bg-primary/10 text-primary rounded-md border px-2 py-0.5 text-[10px]">
                                {hotspot.keyword.text}
                            </span>
                        )}
                        {/* 真实性标记 */}
                        {!hotspot.isReal && (
                            <span className="border-destructive/20 bg-destructive/10 text-destructive flex items-center gap-1 rounded-md border px-2 py-0.5 text-[10px]">
                                <ShieldAlert className="h-3 w-3" />
                                可疑
                            </span>
                        )}
                        {hotspot.isReal && hotspot.relevance >= 80 && (
                            <span className="border-chart-2/20 bg-chart-2/10 text-chart-2 flex items-center gap-1 rounded-md border px-2 py-0.5 text-[10px]">
                                <Shield className="h-3 w-3" />
                                可信
                            </span>
                        )}
                        {hotspot.keywordMentioned === true && (
                            <span className="border-chart-4/20 bg-chart-4/10 text-chart-4 flex items-center gap-1 rounded-md border px-2 py-0.5 text-[10px]">
                                <Target className="h-3 w-3" />
                                直接提及
                            </span>
                        )}
                        {hotspot.keywordMentioned === false && (
                            <span className="border-chart-3/20 bg-chart-3/10 text-chart-3 flex items-center gap-1 rounded-md border px-2 py-0.5 text-[10px]">
                                <Target className="h-3 w-3" />
                                间接相关
                            </span>
                        )}
                        {/* 热度综合指标 */}
                        <span
                            className={cn(
                                'border-border bg-muted/20 flex items-center gap-1 rounded-md border px-2 py-0.5 text-[10px] font-medium',
                                heat.color.includes('red')
                                    ? 'text-destructive'
                                    : heat.color.includes('orange')
                                      ? 'text-chart-3'
                                      : heat.color.includes('amber')
                                        ? 'text-chart-5'
                                        : heat.color.includes('blue')
                                          ? 'text-primary'
                                          : 'text-muted-foreground'
                            )}>
                            <ThermometerSun className="h-3 w-3" />
                            {heat.label} {heatScore}
                        </span>
                    </div>

                    {/* Title */}
                    <h3 className="text-foreground group-hover:text-primary mb-2 line-clamp-2 font-medium transition-colors">{hotspot.title}</h3>

                    {/* AI Summary - 标注 */}
                    {hotspot.summary && (
                        <div className="mb-3">
                            <span className="text-primary/60 mr-1.5 text-[10px] font-medium">AI 摘要</span>
                            <span className="text-muted-foreground text-sm">{hotspot.summary}</span>
                        </div>
                    )}

                    {/* 作者信息 */}
                    {hotspot.authorName && (
                        <div className="mb-3 flex items-center gap-2">
                            {hotspot.authorAvatar ? (
                                <img src={hotspot.authorAvatar} alt="" className="border-border h-5 w-5 rounded-full border object-cover" />
                            ) : (
                                <User className="text-muted-foreground h-4 w-4" />
                            )}
                            <span className="text-muted-foreground text-xs">
                                {hotspot.authorName}
                                {hotspot.authorUsername && <span className="text-muted-foreground/60 ml-1">@{hotspot.authorUsername}</span>}
                            </span>
                            {hotspot.authorVerified && <span className="bg-primary/15 text-primary rounded px-1.5 py-0.5 text-[10px]">✓ 认证</span>}
                            {hotspot.authorFollowers != null && hotspot.authorFollowers > 0 && (
                                <span className="text-muted-foreground/60 text-[10px]">{hotspot.authorFollowers.toLocaleString()} 粉丝</span>
                            )}
                        </div>
                    )}

                    {/* 互动数据 */}
                    <div className="text-muted-foreground mb-2 flex flex-wrap items-center gap-3 text-xs">
                        <span className="hover:text-primary flex cursor-default items-center gap-1 transition-colors">
                            <Target className="h-3.5 w-3.5" />
                            相关性 {hotspot.relevance}%
                        </span>
                        {hotspot.likeCount != null && hotspot.likeCount > 0 && (
                            <span className="hover:text-primary flex cursor-default items-center gap-1 transition-colors" title="点赞">
                                <Zap className="h-3.5 w-3.5" />
                                {hotspot.likeCount.toLocaleString()}
                            </span>
                        )}
                        {hotspot.retweetCount != null && hotspot.retweetCount > 0 && (
                            <span className="hover:text-primary flex cursor-default items-center gap-1 transition-colors" title="转发">
                                <Repeat2 className="h-3.5 w-3.5" />
                                {hotspot.retweetCount.toLocaleString()}
                            </span>
                        )}
                        {hotspot.replyCount != null && hotspot.replyCount > 0 && (
                            <span className="hover:text-primary flex cursor-default items-center gap-1 transition-colors" title="回复">
                                <MessageCircle className="h-3.5 w-3.5" />
                                {hotspot.replyCount.toLocaleString()}
                            </span>
                        )}
                        {hotspot.commentCount != null && hotspot.commentCount > 0 && (
                            <span className="hover:text-primary flex cursor-default items-center gap-1 transition-colors" title="评论">
                                <MessageCircle className="h-3.5 w-3.5" />
                                {hotspot.commentCount.toLocaleString()}
                            </span>
                        )}
                        {hotspot.quoteCount != null && hotspot.quoteCount > 0 && (
                            <span className="hover:text-primary flex cursor-default items-center gap-1 transition-colors" title="引用">
                                <Quote className="h-3.5 w-3.5" />
                                {hotspot.quoteCount.toLocaleString()}
                            </span>
                        )}
                        {hotspot.viewCount != null && hotspot.viewCount > 0 && (
                            <span className="hover:text-primary flex cursor-default items-center gap-1 transition-colors" title="浏览量">
                                <Eye className="h-3.5 w-3.5" />
                                {hotspot.viewCount.toLocaleString()}
                            </span>
                        )}
                        {hotspot.danmakuCount != null && hotspot.danmakuCount > 0 && (
                            <span className="hover:text-primary flex cursor-default items-center gap-1 transition-colors" title="弹幕">
                                💬 {hotspot.danmakuCount.toLocaleString()}
                            </span>
                        )}
                    </div>

                    {/* 时间信息 */}
                    <div className="text-muted-foreground/70 flex flex-wrap items-center gap-3 text-[11px]">
                        {hotspot.publishedAt && (
                            <span className="flex items-center gap-1" title={`发布于 ${formatDateTime(hotspot.publishedAt)}`}>
                                <Clock className="h-3 w-3" />
                                发布 {relativeTime(hotspot.publishedAt)}
                            </span>
                        )}
                        <span className="flex items-center gap-1" title={`抓取于 ${formatDateTime(hotspot.createdAt)}`}>
                            <Activity className="h-3 w-3" />
                            抓取 {relativeTime(hotspot.createdAt)}
                        </span>
                    </div>

                    {/* AI 相关性理由 - 可折叠 */}
                    {hotspot.relevanceReason && (
                        <div className="text-primary mt-2 font-medium">
                            <button
                                onClick={() => onToggleReason(hotspot.id)}
                                className="text-primary/70 hover:text-primary flex items-center gap-1 text-[11px] transition-colors">
                                {isExpandedReason ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
                                AI 分析理由
                            </button>
                            <AnimatePresence>
                                {isExpandedReason && (
                                    <motion.div
                                        initial={{ height: 0, opacity: 0 }}
                                        animate={{ height: 'auto', opacity: 1 }}
                                        exit={{ height: 0, opacity: 0 }}
                                        className="overflow-hidden">
                                        <p className="border-primary/20 bg-primary/5 text-muted-foreground mt-1 rounded-r-lg border-l-2 p-2 text-xs whitespace-pre-wrap">
                                            {hotspot.relevanceReason}
                                        </p>
                                    </motion.div>
                                )}
                            </AnimatePresence>
                        </div>
                    )}

                    {/* 原始内容 - 可折叠 */}
                    {hotspot.content && hotspot.content !== hotspot.summary && (
                        <div className="mt-2">
                            <button
                                onClick={() => onToggleContent(hotspot.id)}
                                className="text-muted-foreground/60 hover:text-foreground flex items-center gap-1 text-[11px] transition-colors">
                                {isExpandedContent ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
                                <FileText className="h-3 w-3" />
                                原始内容
                            </button>
                            <AnimatePresence>
                                {isExpandedContent && (
                                    <motion.div
                                        initial={{ height: 0, opacity: 0 }}
                                        animate={{ height: 'auto', opacity: 1 }}
                                        exit={{ height: 0, opacity: 0 }}
                                        className="overflow-hidden">
                                        <p className="border-border bg-muted/10 text-muted-foreground mt-1 max-h-40 overflow-y-auto rounded-r-lg border-l-2 p-2 text-xs wrap-break-word whitespace-pre-wrap">
                                            {hotspot.content}
                                        </p>
                                    </motion.div>
                                )}
                            </AnimatePresence>
                        </div>
                    )}
                </div>

                {/* Link */}
                <a
                    href={hotspot.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={e => e.stopPropagation()}
                    className="bg-muted text-muted-foreground hover:bg-primary hover:text-primary-foreground rounded-xl p-2.5 opacity-40 shadow-sm transition-all group-hover:opacity-100">
                    <ExternalLink className="h-4 w-4" />
                </a>
            </div>
        </motion.div>
    );
};

export default HotSpotCard;
