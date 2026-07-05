import { useRouterState } from '@tanstack/react-router';
import { AnimatePresence, motion } from 'framer-motion';
import { Bell, RefreshCw } from 'lucide-react';
import { cn } from '@/lib/ui';
import type { Hotspot } from '@/types';

const pageTitles: Record<string, string> = {
    '/': '热点雷达',
    '/keywords': '监控词',
    '/search': '搜索',
    '/settings': '设置'
};

interface ContentHeaderProps {
    isChecking: boolean;
    unreadCount: number;
    notifications: Hotspot[];
    showNotifications: boolean;
    setShowNotifications: (show: boolean) => void;
    onMarkAllRead: () => void;
    onManualCheck: () => void;
}

export function ContentHeader({
    isChecking,
    onManualCheck,
    unreadCount,
    notifications,
    showNotifications,
    setShowNotifications,
    onMarkAllRead
}: ContentHeaderProps) {
    const { location } = useRouterState();
    const title = pageTitles[location.pathname] ?? '';

    return (
        <header className="border-border bg-background/80 flex shrink-0 items-center justify-between border-b px-6 py-3 backdrop-blur-sm">
            <h1 className="text-foreground text-lg font-semibold">{title}</h1>

            <div className="flex items-center gap-2">
                <button
                    onClick={onManualCheck}
                    disabled={isChecking}
                    className={cn(
                        'flex items-center gap-2 rounded-lg px-3.5 py-2 text-sm font-medium transition-colors',
                        isChecking ? 'bg-primary/10 text-primary cursor-wait' : 'bg-primary text-primary-foreground hover:bg-primary/90'
                    )}>
                    <RefreshCw className={cn('h-3.5 w-3.5', isChecking && 'animate-spin')} />
                    {isChecking ? '扫描中' : '立即扫描'}
                </button>

                <div className="relative">
                    <button
                        onClick={() => setShowNotifications(!showNotifications)}
                        className="border-border hover:bg-muted relative rounded-lg border p-2 transition-colors">
                        <Bell className="text-muted-foreground h-[18px] w-[18px]" />
                        {unreadCount > 0 && (
                            <span className="bg-destructive text-destructive-foreground absolute -top-1 -right-1 flex h-4.5 w-4.5 items-center justify-center rounded-full text-[9px] font-bold">
                                {unreadCount > 99 ? '99+' : unreadCount}
                            </span>
                        )}
                    </button>

                    <AnimatePresence>
                        {showNotifications && (
                            <>
                                <div className="fixed inset-0 z-[-1]" onClick={() => setShowNotifications(false)} />
                                <motion.div
                                    initial={{ opacity: 0, y: 4, scale: 0.97 }}
                                    animate={{ opacity: 1, y: 0, scale: 1 }}
                                    exit={{ opacity: 0, y: 4, scale: 0.97 }}
                                    className="border-border bg-popover absolute top-12 right-0 z-50 w-80 overflow-hidden rounded-xl border shadow-xl">
                                    <div className="border-border flex items-center justify-between border-b px-4 py-3">
                                        <h3 className="text-popover-foreground text-sm font-medium">通知</h3>
                                        {unreadCount > 0 && (
                                            <button onClick={onMarkAllRead} className="text-primary text-xs hover:underline">
                                                全部已读
                                            </button>
                                        )}
                                    </div>
                                    <div className="max-h-72 overflow-y-auto">
                                        {notifications.length === 0 ? (
                                            <p className="text-muted-foreground py-8 text-center text-sm">暂无通知</p>
                                        ) : (
                                            <div className="divide-border divide-y">
                                                {notifications.slice(0, 10).map(n => (
                                                    <div
                                                        key={n.id}
                                                        className={cn('px-4 py-3 transition-colors', n.isRead ? 'opacity-50' : 'hover:bg-muted/50')}>
                                                        <p className="text-popover-foreground text-sm font-medium">{n.title}</p>
                                                        <p className="text-muted-foreground mt-0.5 line-clamp-2 text-xs">{n.summary ?? n.content}</p>
                                                    </div>
                                                ))}
                                            </div>
                                        )}
                                    </div>
                                </motion.div>
                            </>
                        )}
                    </AnimatePresence>
                </div>
            </div>
        </header>
    );
}
