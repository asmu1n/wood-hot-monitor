import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ArrowUpDown, Filter, X, Clock, Flame, TrendingUp, Target, ChevronDown, Check, RotateCcw } from 'lucide-react';
import { cn } from '@repo/ui';

export interface FilterState {
    source: string;
    importance: string;
    keywordId: string;
    timeRange: string;
    isReal: string;
    sortBy: string;
    sortOrder: string;
}

export const defaultFilterState: FilterState = {
    source: '',
    importance: '',
    keywordId: '',
    timeRange: '',
    isReal: '',
    sortBy: 'createdAt',
    sortOrder: 'desc'
};

interface FilterSortBarProps {
    filters: FilterState;
    onChange: (filters: FilterState) => void;
    keywords: Keyword[];
}

const SORT_OPTIONS = [
    { value: 'createdAt', label: '最新发现', icon: Clock },
    { value: 'publishedAt', label: '最新发布', icon: Clock },
    { value: 'importance', label: '重要程度', icon: Flame },
    { value: 'relevance', label: '相关性', icon: Target },
    { value: 'hot', label: '热度综合', icon: TrendingUp }
];

const SOURCE_OPTIONS = [
    { value: '', label: '全部来源' },
    { value: 'twitter', label: 'Twitter' },
    { value: 'bing', label: 'Bing' },
    { value: 'google', label: 'Google' },
    { value: 'sogou', label: '搜狗' },
    { value: 'bilibili', label: 'Bilibili' },
    { value: 'weibo', label: '微博热搜' },
    { value: 'hackernews', label: 'HackerNews' },
    { value: 'duckduckgo', label: 'DuckDuckGo' }
];

const IMPORTANCE_OPTIONS = [
    { value: '', label: '全部等级' },
    { value: 'urgent', label: '🔴 紧急', color: 'text-red-400' },
    { value: 'high', label: '🟠 高', color: 'text-orange-400' },
    { value: 'medium', label: '🟡 中', color: 'text-amber-400' },
    { value: 'low', label: '🟢 低', color: 'text-emerald-400' }
];

const TIME_RANGE_OPTIONS = [
    { value: '', label: '全部时间' },
    { value: '1h', label: '最近 1 小时' },
    { value: 'today', label: '今天' },
    { value: '7d', label: '最近 7 天' },
    { value: '30d', label: '最近 30 天' }
];

const REAL_OPTIONS = [
    { value: '', label: '全部' },
    { value: 'true', label: '✅ 真实' },
    { value: 'false', label: '⚠️ 疑似虚假' }
];

// Dropdown component
function Dropdown({
    label,
    value,
    options,
    onChange
}: {
    label: string;
    value: string;
    options: { value: string; label: string; color?: string }[];
    onChange: (v: string) => void;
}) {
    const [open, setOpen] = useState(false);
    const selected = options.find(o => o.value === value);
    const isActive = value !== '';

    return (
        <div className="relative">
            <button
                onClick={() => setOpen(!open)}
                className={cn(
                    'flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium whitespace-nowrap shadow-sm transition-all',
                    isActive
                        ? 'border-primary/30 bg-primary/15 text-primary border'
                        : 'border-border bg-muted/50 text-muted-foreground hover:border-border/80 hover:text-foreground border'
                )}>
                <span>{isActive ? selected?.label : label}</span>
                <ChevronDown className={cn('h-3 w-3 opacity-60 transition-transform', open && 'rotate-180')} />
            </button>

            <AnimatePresence>
                {open && (
                    <>
                        <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
                        <motion.div
                            initial={{ opacity: 0, y: 4, scale: 0.96 }}
                            animate={{ opacity: 1, y: 0, scale: 1 }}
                            exit={{ opacity: 0, y: 4, scale: 0.96 }}
                            transition={{ duration: 0.15 }}
                            className="border-border bg-popover/98 absolute top-full left-0 z-50 mt-1 min-w-[160px] overflow-hidden rounded-xl border shadow-2xl backdrop-blur-xl">
                            {options.map(option => (
                                <button
                                    key={option.value}
                                    onClick={() => {
                                        onChange(option.value);
                                        setOpen(false);
                                    }}
                                    className={cn(
                                        'flex w-full items-center gap-2 px-3 py-2 text-left text-xs transition-colors',
                                        value === option.value
                                            ? 'bg-primary/10 text-primary'
                                            : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                                    )}>
                                    {value === option.value && <Check className="h-3 w-3 shrink-0" />}
                                    <span className={cn(option.color)}>{option.label}</span>
                                </button>
                            ))}
                        </motion.div>
                    </>
                )}
            </AnimatePresence>
        </div>
    );
}

export default function FilterSortBar({ filters, onChange, keywords }: FilterSortBarProps) {
    const [showFilters, setShowFilters] = useState(false);

    const activeFilterCount = [filters.source, filters.importance, filters.keywordId, filters.timeRange, filters.isReal].filter(v => v !== '').length;

    const hasNonDefaultSort = filters.sortBy !== 'createdAt';

    const update = (key: keyof FilterState, value: string) => {
        onChange({ ...filters, [key]: value });
    };

    const resetFilters = () => {
        onChange({ ...defaultFilterState });
    };

    const keywordOptions = [{ value: '', label: '全部关键词' }, ...keywords.filter(k => k.isActive).map(k => ({ value: k.id, label: k.text }))];

    return (
        <div className="space-y-3">
            {/* Main Bar: Sort + Filter Toggle */}
            <div className="flex flex-wrap items-center gap-2">
                {/* Sort Selector */}
                <div className="border-border bg-muted/30 flex items-center gap-1 rounded-xl border p-1">
                    <ArrowUpDown className="text-muted-foreground/60 ml-2 h-3.5 w-3.5" />
                    {SORT_OPTIONS.map(opt => {
                        const Icon = opt.icon;

                        return (
                            <button
                                key={opt.value}
                                onClick={() => update('sortBy', opt.value)}
                                className={cn(
                                    'flex items-center gap-1 rounded-lg px-2.5 py-1.5 text-xs font-medium whitespace-nowrap transition-all',
                                    filters.sortBy === opt.value
                                        ? 'bg-primary/10 text-primary shadow-sm'
                                        : 'text-muted-foreground hover:text-foreground'
                                )}>
                                <Icon className="h-3 w-3" />
                                {opt.label}
                            </button>
                        );
                    })}
                </div>

                {/* Filter Toggle */}
                <button
                    onClick={() => setShowFilters(!showFilters)}
                    className={cn(
                        'flex items-center gap-1.5 rounded-xl px-3 py-2 text-xs font-medium shadow-sm transition-all',
                        showFilters || activeFilterCount > 0
                            ? 'border-primary/30 bg-primary/15 text-primary border'
                            : 'border-border bg-muted/50 text-muted-foreground hover:border-border/80 hover:text-foreground border'
                    )}>
                    <Filter className="h-3.5 w-3.5 opacity-60" />
                    筛选
                    {activeFilterCount > 0 && (
                        <span className="bg-primary text-primary-foreground flex h-4 w-4 items-center justify-center rounded-full text-[10px] font-bold">
                            {activeFilterCount}
                        </span>
                    )}
                </button>

                {/* Reset */}
                {(activeFilterCount > 0 || hasNonDefaultSort) && (
                    <button
                        onClick={resetFilters}
                        className="text-muted-foreground hover:text-foreground flex items-center gap-1 rounded-xl px-2.5 py-2 text-xs transition-colors">
                        <RotateCcw className="h-3 w-3" />
                        重置
                    </button>
                )}

                {/* Active Filter Tags */}
                {activeFilterCount > 0 && !showFilters && (
                    <div className="flex flex-wrap items-center gap-1.5">
                        {filters.source && (
                            <FilterTag
                                label={SOURCE_OPTIONS.find(o => o.value === filters.source)?.label || filters.source}
                                onRemove={() => update('source', '')}
                            />
                        )}
                        {filters.importance && (
                            <FilterTag
                                label={IMPORTANCE_OPTIONS.find(o => o.value === filters.importance)?.label || filters.importance}
                                onRemove={() => update('importance', '')}
                            />
                        )}
                        {filters.keywordId && (
                            <FilterTag
                                label={keywords.find(k => k.id === filters.keywordId)?.text || '关键词'}
                                onRemove={() => update('keywordId', '')}
                            />
                        )}
                        {filters.timeRange && (
                            <FilterTag
                                label={TIME_RANGE_OPTIONS.find(o => o.value === filters.timeRange)?.label || filters.timeRange}
                                onRemove={() => update('timeRange', '')}
                            />
                        )}
                        {filters.isReal && (
                            <FilterTag
                                label={REAL_OPTIONS.find(o => o.value === filters.isReal)?.label || '真实性'}
                                onRemove={() => update('isReal', '')}
                            />
                        )}
                    </div>
                )}
            </div>

            {/* Expanded Filter Panel */}
            <AnimatePresence>
                {showFilters && (
                    <motion.div
                        initial={{ opacity: 0, height: 0 }}
                        animate={{ opacity: 1, height: 'auto' }}
                        exit={{ opacity: 0, height: 0 }}
                        transition={{ duration: 0.2 }}>
                        <div className="border-border bg-muted/30 flex flex-wrap items-center gap-2 rounded-xl border p-3 shadow-inner">
                            <Dropdown label="来源" value={filters.source} options={SOURCE_OPTIONS} onChange={v => update('source', v)} />
                            <Dropdown
                                label="重要程度"
                                value={filters.importance}
                                options={IMPORTANCE_OPTIONS}
                                onChange={v => update('importance', v)}
                            />
                            <Dropdown label="关键词" value={filters.keywordId} options={keywordOptions} onChange={v => update('keywordId', v)} />
                            <Dropdown label="时间" value={filters.timeRange} options={TIME_RANGE_OPTIONS} onChange={v => update('timeRange', v)} />
                            <Dropdown label="真实性" value={filters.isReal} options={REAL_OPTIONS} onChange={v => update('isReal', v)} />
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
}

function FilterTag({ label, onRemove }: { label: string; onRemove: () => void }) {
    return (
        <span className="border-primary/20 bg-primary/10 text-primary hover:bg-primary/15 inline-flex items-center gap-1 rounded-md border px-2 py-1 text-[10px] font-medium shadow-sm transition-colors">
            {label}
            <button onClick={onRemove} className="hover:text-foreground opacity-60 transition-colors hover:opacity-100">
                <X className="h-2.5 w-2.5" />
            </button>
        </span>
    );
}
