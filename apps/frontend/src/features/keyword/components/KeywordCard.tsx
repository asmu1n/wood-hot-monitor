import React from 'react';
import { motion } from 'framer-motion';
import { Trash2 } from 'lucide-react';
import { cn } from '@repo/ui';
import type { Keyword } from '@repo/types';

interface KeywordCardProps {
    keyword: Keyword;
    index: number;
    onToggle: (keyword: Keyword) => void;
    onDelete: (keyword: Keyword) => void;
}

const KeywordCard: React.FC<KeywordCardProps> = ({ keyword, index, onToggle, onDelete }) => {
    return (
        <motion.div
            layout
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.9 }}
            transition={{ delay: index * 0.02 }}
            className={cn(
                'group rounded-xl border p-4 shadow-sm transition-all',
                keyword.isActive ? 'border-primary/20 bg-muted/40 hover:border-primary/30' : 'border-border bg-muted/10 opacity-60'
            )}>
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    {/* Toggle */}
                    <button
                        onClick={() => onToggle(keyword)}
                        className={cn('relative h-6 w-11 rounded-full transition-all', keyword.isActive ? 'bg-primary' : 'bg-muted-foreground/30')}>
                        <span
                            className={cn(
                                'bg-background absolute top-1 h-4 w-4 rounded-full shadow-sm transition-all',
                                keyword.isActive ? 'left-6' : 'left-1'
                            )}
                        />
                    </button>

                    <div>
                        <span className={cn('font-medium', keyword.isActive ? 'text-foreground' : 'text-muted-foreground')}>{keyword.text}</span>
                        {keyword._count && keyword._count.hotspots > 0 && (
                            <span className="text-muted-foreground/60 ml-2 text-xs">{keyword._count.hotspots} 条热点</span>
                        )}
                    </div>
                </div>

                <button
                    onClick={() => onDelete(keyword)}
                    className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive rounded-lg p-2 opacity-0 transition-all group-hover:opacity-100">
                    <Trash2 className="h-4 w-4" />
                </button>
            </div>
        </motion.div>
    );
};

export default KeywordCard;
