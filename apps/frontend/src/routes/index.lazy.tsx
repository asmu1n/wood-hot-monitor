import { createLazyFileRoute } from '@tanstack/react-router';
import { Flame, Search, ChevronsUpDown } from 'lucide-react';

import { useApp } from '@/context/AppContext';
import StatusCards from '@/features/hotspot/components/StatusCards';
import FilterSortBar from '@/components/FilterSortBar';
import HotSpotCard from '@/features/hotspot/components/HotSpotCard';
import HotSpotPagination from '@/features/hotspot/components/HotSpotPagination';

function Dashboard() {
    const {
        status,
        keywords,
        hotSpots,
        dashboardFilters,
        setDashboardFilters,
        isLoading,
        currentPage,
        setCurrentPage,
        totalPages,
        expandedReasons,
        expandedContents,
        toggleReason,
        toggleContent,
        toggleAllReasons,
        allReasonsExpanded
    } = useApp();

    const activeKeywordsCount = keywords.filter(k => k.isActive).length;

    return (
        <div className="space-y-8">
            {/* Hero Stats */}
            <StatusCards stats={status} activeKeywordsCount={activeKeywordsCount} />

            {/* Hotspots Feed */}
            <div>
                <div className="mb-4 flex items-center justify-between">
                    <h2 className="text-foreground flex items-center gap-2 text-lg font-semibold">
                        <Flame className="text-primary h-5 w-5" />
                        实时热点流
                    </h2>
                    <span className="text-muted-foreground text-xs opacity-60">每 30 分钟自动更新</span>
                </div>

                {/* Filter & Sort Bar */}
                <div className="mb-5">
                    <FilterSortBar filters={dashboardFilters} onChange={setDashboardFilters} keywords={keywords} />
                </div>

                {isLoading ? (
                    <div className="flex items-center justify-center py-16">
                        <div className="border-primary/30 border-t-primary h-8 w-8 animate-spin rounded-full border-2" />
                    </div>
                ) : hotSpots.length === 0 ? (
                    <div className="border-border bg-muted/20 rounded-2xl border border-dashed py-16 text-center">
                        <div className="bg-muted mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full">
                            <Search className="text-muted-foreground h-8 w-8" />
                        </div>
                        <p className="text-foreground font-medium">尚未发现热点</p>
                        <p className="text-muted-foreground mt-1 text-sm">添加监控关键词开始追踪</p>
                    </div>
                ) : (
                    <div className="space-y-3">
                        {/* 一键展开/折叠所有理由 */}
                        {hotSpots.some(h => h.relevanceReason) && (
                            <div className="flex justify-end">
                                <button
                                    onClick={() => toggleAllReasons(hotSpots)}
                                    className="text-muted-foreground hover:bg-muted hover:text-primary flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs transition-colors">
                                    <ChevronsUpDown className="h-3.5 w-3.5" />
                                    {allReasonsExpanded ? '折叠所有理由' : '展开所有理由'}
                                </button>
                            </div>
                        )}

                        {hotSpots.map((hotspot, index) => (
                            <HotSpotCard
                                key={hotspot.id}
                                hotspot={hotspot}
                                index={index}
                                isExpandedReason={expandedReasons.has(hotspot.id)}
                                isExpandedContent={expandedContents.has(hotspot.id)}
                                onToggleReason={toggleReason}
                                onToggleContent={toggleContent}
                            />
                        ))}
                    </div>
                )}

                {/* Pagination */}
                <HotSpotPagination currentPage={currentPage} totalPages={totalPages} totalItems={status?.total || 0} onPageChange={setCurrentPage} />
            </div>
        </div>
    );
}

export const Route = createLazyFileRoute('/')({
    component: Dashboard
});
