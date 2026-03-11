import { createLazyFileRoute } from '@tanstack/react-router';
import { useMemo } from 'react';
import { motion } from 'framer-motion';
import { Search } from 'lucide-react';

import { useApp } from '@/context/AppContext';
import FilterSortBar from '@/components/FilterSortBar';
import HotSpotCard from '@/features/hotspot/components/HotSpotCard';
import { sortHotSpots } from '@/features/hotspot/utils';

function SearchPage() {
    const {
        searchQuery,
        setSearchQuery,
        handleSearch,
        isLoading,
        searchFilters,
        setSearchFilters,
        keywords,
        searchResults,
        expandedReasons,
        expandedContents,
        toggleReason,
        toggleContent
    } = useApp();

    // Client-side filtering/sorting for search results
    const filteredSearchResults = useMemo(() => {
        let results = [...searchResults];

        // Apply filters
        if (searchFilters.source) {
            results = results.filter(h => h.source === searchFilters.source);
        }

        if (searchFilters.importance) {
            results = results.filter(h => h.importance === searchFilters.importance);
        }

        if (searchFilters.isReal === 'true') {
            results = results.filter(h => h.isReal);
        } else if (searchFilters.isReal === 'false') {
            results = results.filter(h => !h.isReal);
        }

        if (searchFilters.keywordId) {
            results = results.filter(h => h.keyword?.id === searchFilters.keywordId);
        }

        if (searchFilters.timeRange) {
            const now = new Date();
            let dateFrom: Date | null = null;

            switch (searchFilters.timeRange) {
                case '1h':
                    dateFrom = new Date(now.getTime() - 60 * 60 * 1000);
                    break;
                case 'today':
                    dateFrom = new Date(now);
                    dateFrom.setHours(0, 0, 0, 0);
                    break;
                case '7d':
                    dateFrom = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
                    break;
                case '30d':
                    dateFrom = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
                    break;
            }

            if (dateFrom) {
                results = results.filter(h => new Date(h.createdAt) >= dateFrom!);
            }
        }

        // Apply sorting using shared utility
        results = sortHotSpots(results, searchFilters.sortBy || 'createdAt', (searchFilters.sortOrder || 'desc') as 'asc' | 'desc');

        return results;
    }, [searchResults, searchFilters]);

    return (
        <div className="space-y-6">
            {/* Search Form */}
            <form onSubmit={handleSearch} className="border-border bg-muted/30 rounded-2xl border p-5 shadow-sm">
                <div className="flex gap-3">
                    <div className="relative flex-1">
                        <Search className="text-muted-foreground/60 absolute top-1/2 left-4 h-5 w-5 -translate-y-1/2" />
                        <input
                            type="text"
                            value={searchQuery}
                            onChange={e => setSearchQuery(e.target.value)}
                            placeholder="搜索热点内容..."
                            className="border-border bg-background text-foreground placeholder-muted-foreground/50 focus:border-primary/50 focus:ring-primary/20 w-full rounded-xl border py-3 pr-4 pl-12 transition-all focus:ring-2 focus:outline-none"
                        />
                    </div>
                    <motion.button
                        type="submit"
                        disabled={isLoading}
                        whileHover={{ scale: 1.02 }}
                        whileTap={{ scale: 0.98 }}
                        className="bg-primary text-primary-foreground shadow-primary/25 flex items-center gap-2 rounded-xl px-6 py-3 font-medium shadow-lg disabled:opacity-50">
                        {isLoading ? (
                            <div className="border-primary-foreground/30 border-t-primary-foreground h-4 w-4 animate-spin rounded-full border-2" />
                        ) : (
                            <Search className="h-4 w-4" />
                        )}
                        搜索
                    </motion.button>
                </div>
            </form>

            {/* Search Filter & Sort Bar */}
            <FilterSortBar filters={searchFilters} onChange={setSearchFilters} keywords={keywords} />

            {/* Search Results */}
            <div className="space-y-3">
                {filteredSearchResults.length === 0 && searchResults.length > 0 && (
                    <div className="rounded-2xl border border-dashed border-white/10 py-12 text-center">
                        <p className="text-slate-500">当前筛选条件下无结果</p>
                        <p className="mt-1 text-sm text-slate-600">尝试调整筛选条件</p>
                    </div>
                )}
                {filteredSearchResults.map((hotspot, i) => (
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
        </div>
    );
}

export const Route = createLazyFileRoute('/search')({
    component: SearchPage
});
