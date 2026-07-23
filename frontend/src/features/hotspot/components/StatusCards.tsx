import { motion } from 'framer-motion';
import { Activity, Clock, AlertTriangle, Target } from 'lucide-react';
import { cn } from '@/lib/ui';
import type { Status } from '@wails/internal/module/hotspot';

interface StatsCardsProps {
    stats: Status | null;
    activeKeywordsCount: number;
    className?: string;
}

function StatusCards({ stats, activeKeywordsCount, className }: StatsCardsProps) {
    if (!stats) {
        return null;
    }

    return (
        <div className={cn('grid grid-cols-2 gap-4 lg:grid-cols-4', className)}>
            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="border-primary/10 bg-muted/30 hover:bg-muted/50 rounded-2xl border p-5 shadow-sm transition-colors">
                <div className="text-muted-foreground mb-2 flex items-center gap-2 text-sm font-medium">
                    <Activity className="text-primary h-4 w-4" />
                    总热点
                </div>
                <p className="text-foreground text-3xl font-bold tracking-tight">{stats.total}</p>
            </motion.div>

            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.05 }}
                className="border-chart-1/10 bg-muted/30 hover:bg-muted/50 rounded-2xl border p-5 shadow-sm transition-colors">
                <div className="text-muted-foreground mb-2 flex items-center gap-2 text-sm font-medium">
                    <Clock className="text-chart-1 h-4 w-4" />
                    今日新增
                </div>
                <p className="text-chart-1 text-3xl font-bold tracking-tight">{stats.today}</p>
            </motion.div>

            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 }}
                className="border-destructive/10 bg-muted/30 hover:bg-muted/50 rounded-2xl border p-5 shadow-sm transition-colors">
                <div className="text-muted-foreground mb-2 flex items-center gap-2 text-sm font-medium">
                    <AlertTriangle className="text-destructive h-4 w-4" />
                    紧急热点
                </div>
                <p className="text-destructive text-3xl font-bold tracking-tight">{stats.urgent}</p>
            </motion.div>

            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.15 }}
                className="border-chart-2/10 bg-muted/30 hover:bg-muted/50 rounded-2xl border p-5 shadow-sm transition-colors">
                <div className="text-muted-foreground mb-2 flex items-center gap-2 text-sm font-medium">
                    <Target className="text-chart-2 h-4 w-4" />
                    监控词
                </div>
                <p className="text-chart-2 text-3xl font-bold tracking-tight">{activeKeywordsCount}</p>
            </motion.div>
        </div>
    );
}

export default StatusCards;
