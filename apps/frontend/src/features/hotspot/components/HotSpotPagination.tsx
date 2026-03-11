import React from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { cn } from '@/utils/common';

interface HotSpotPaginationProps {
    currentPage: number;
    totalPages: number;
    totalItems: number;
    onPageChange: (page: number) => void;
}

const HotSpotPagination: React.FC<HotSpotPaginationProps> = ({ currentPage, totalPages, totalItems, onPageChange }) => {
    if (totalPages <= 1) {
        return null;
    }

    return (
        <div className="mt-6 flex items-center justify-center gap-3">
            <button
                onClick={() => onPageChange(Math.max(1, currentPage - 1))}
                disabled={currentPage <= 1}
                className="border-border bg-muted/50 text-muted-foreground hover:border-border/80 hover:text-foreground rounded-xl border p-2 transition-all disabled:cursor-not-allowed disabled:opacity-30">
                <ChevronLeft className="h-4 w-4" />
            </button>
            <div className="flex items-center gap-1.5">
                {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
                    let page: number;

                    if (totalPages <= 7) {
                        page = i + 1;
                    } else if (currentPage <= 4) {
                        page = i + 1;
                    } else if (currentPage >= totalPages - 3) {
                        page = totalPages - 6 + i;
                    } else {
                        page = currentPage - 3 + i;
                    }

                    return (
                        <button
                            key={page}
                            onClick={() => onPageChange(page)}
                            className={cn(
                                'h-8 w-8 rounded-lg text-xs font-medium shadow-sm transition-all',
                                currentPage === page
                                    ? 'border-primary/30 bg-primary/20 text-primary border'
                                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                            )}>
                            {page}
                        </button>
                    );
                })}
            </div>
            <button
                onClick={() => onPageChange(Math.min(totalPages, currentPage + 1))}
                disabled={currentPage >= totalPages}
                className="border-border bg-muted/50 text-muted-foreground hover:border-border/80 hover:text-foreground rounded-xl border p-2 transition-all disabled:cursor-not-allowed disabled:opacity-30">
                <ChevronRight className="h-4 w-4" />
            </button>
            <span className="text-muted-foreground/60 ml-2 text-xs">共 {totalItems} 条</span>
        </div>
    );
};

export default HotSpotPagination;
