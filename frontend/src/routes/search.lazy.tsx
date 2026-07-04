import { createLazyFileRoute } from '@tanstack/react-router';
import { useState, useMemo } from 'react';
import { motion } from 'framer-motion';
import { Search, ChevronsUpDown } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';

import { useHotspotExpand } from '@/features/hotspot/hooks/useHotspotExpand';
import HotSpotCard from '@/features/hotspot/components/HotSpotCard';
import HotSpotPagination from '@/features/hotspot/components/HotSpotPagination';
import { hotspotApi } from '@/features/hotspot/api';

const LIMIT_COUNT = 20;

function SearchPage() {
    const { expandedReasons, expandedContents, allReasonsExpanded, toggleReason, toggleContent, toggleAllReasons } = useHotspotExpand();

    const [searchInput, setSearchInput] = useState('');
    const [submittedQuery, setSubmittedQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);

    const {
        data: searchRes,
        isLoading,
        isFetching
    } = useQuery({
        queryKey: ['search', submittedQuery, currentPage],
        queryFn: () => hotspotApi.search({ query: submittedQuery, page: currentPage, limit: LIMIT_COUNT, sources: [] }),
        enabled: !!submittedQuery
    });

    const searchResults = useMemo(() => searchRes?.data ?? [], [searchRes]);
    const totalPages = useMemo(() => {
        if (!searchRes) return 1;
        return Math.ceil(searchRes.total / searchRes.limit) || 1;
    }, [searchRes]);

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        const trimmed = searchInput.trim();

        if (!trimmed) return;

        setSubmittedQuery(trimmed);
        setCurrentPage(1);
    };

    return (
        <div className="flex h-full flex-col space-y-6">
            <form onSubmit={handleSearch} className="border-border bg-muted/30 flex-none rounded-2xl border p-5 shadow-sm">
                <div className="flex gap-3">
                    <div className="relative flex-1">
                        <Search className="text-muted-foreground/60 absolute top-1/2 left-4 h-5 w-5 -translate-y-1/2" />
                        <input
                            type="text"
                            value={searchInput}
                            onChange={e => setSearchInput(e.target.value)}
                            placeholder="搜索热点内容..."
                            className="border-border bg-background text-foreground placeholder-muted-foreground/50 focus:border-primary/50 focus:ring-primary/20 w-full rounded-xl border py-3 pr-4 pl-12 transition-all focus:ring-2 focus:outline-none"
                        />
                    </div>
                    <motion.button
                        type="submit"
                        disabled={isFetching}
                        whileHover={{ scale: 1.02 }}
                        whileTap={{ scale: 0.98 }}
                        className="bg-primary text-primary-foreground shadow-primary/25 flex items-center gap-2 rounded-xl px-6 py-3 font-medium shadow-lg disabled:opacity-50">
                        {isFetching ? (
                            <div className="border-primary-foreground/30 border-t-primary-foreground h-4 w-4 animate-spin rounded-full border-2" />
                        ) : (
                            <Search className="h-4 w-4" />
                        )}
                        搜索
                    </motion.button>
                </div>
            </form>

            <div className="flex-1 overflow-y-auto">
                {isLoading ? (
                    <div className="flex items-center justify-center py-16">
                        <div className="border-primary/30 border-t-primary h-8 w-8 animate-spin rounded-full border-2" />
                    </div>
                ) : !submittedQuery ? null : searchResults.length === 0 ? (
                    <div className="border-border bg-muted/20 rounded-2xl border border-dashed py-16 text-center">
                        <div className="bg-muted mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full">
                            <Search className="text-muted-foreground h-8 w-8" />
                        </div>
                        <p className="text-foreground font-medium">未找到相关热点</p>
                        <p className="text-muted-foreground mt-1 text-sm">尝试使用其他关键词搜索</p>
                    </div>
                ) : (
                    <div className="space-y-3">
                        {searchResults.some(h => h.relevanceReason) && (
                            <div className="flex justify-end">
                                <button
                                    onClick={() => toggleAllReasons(searchResults)}
                                    className="text-muted-foreground hover:bg-muted hover:text-primary flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs transition-colors">
                                    <ChevronsUpDown className="h-3.5 w-3.5" />
                                    {allReasonsExpanded ? '折叠所有理由' : '展开所有理由'}
                                </button>
                            </div>
                        )}
                        {searchResults.map((hotspot, i) => (
                            <HotSpotCard
                                key={hotspot.id}
                                hotspot={hotspot}
                                index={i}
                                isExpandedReason={expandedReasons.has(hotspot.id)}
                                isExpandedContent={expandedContents.has(hotspot.id)}
                                onToggleReason={toggleReason}
                                onToggleContent={toggleContent}
                            />
                        ))}
                    </div>
                )}
            </div>

            <HotSpotPagination
                className="flex-none"
                currentPage={currentPage}
                totalPages={totalPages}
                totalItems={searchRes?.total ?? 0}
                onPageChange={setCurrentPage}
            />
        </div>
    );
}

export const Route = createLazyFileRoute('/search')({
    component: SearchPage
});
