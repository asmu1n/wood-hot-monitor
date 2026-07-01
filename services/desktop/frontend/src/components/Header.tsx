import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Flame, RefreshCw, Bell } from 'lucide-react';
import { cn } from '@repo/ui';
import { type Notification } from '@repo/types';

interface HeaderProps {
    isChecking: boolean;
    onManualCheck: () => void;
    unreadCount: number;
    notifications: Notification[];
    showNotifications: boolean;
    setShowNotifications: (show: boolean) => void;
    onMarkAllRead: () => void;
}

const Header: React.FC<HeaderProps> = ({
    isChecking,
    onManualCheck,
    unreadCount,
    notifications,
    showNotifications,
    setShowNotifications,
    onMarkAllRead
}) => {
    return (
        <header className="border-border bg-background/70 sticky top-0 z-40 border-b backdrop-blur-2xl">
            <div className="mx-auto max-w-6xl px-6 py-4">
                <div className="flex items-center justify-between">
                    {/* Logo */}
                    <div className="flex items-center gap-4">
                        <div className="relative">
                            <div className="bg-primary shadow-primary/20 flex h-10 w-10 items-center justify-center rounded-xl shadow-lg">
                                <Flame className="text-primary-foreground h-5 w-5" />
                            </div>
                            <div className="border-background absolute -right-1 -bottom-1 h-3 w-3 animate-pulse rounded-full border-2 bg-emerald-400" />
                        </div>
                        <div>
                            <h1 className="text-foreground text-lg font-semibold tracking-tight">HotMonitor</h1>
                            <p className="text-muted-foreground text-xs">AI 热点雷达</p>
                        </div>
                    </div>

                    {/* Actions */}
                    <div className="flex items-center gap-3">
                        <motion.button
                            onClick={onManualCheck}
                            disabled={isChecking}
                            whileHover={{ scale: 1.02 }}
                            whileTap={{ scale: 0.98 }}
                            className={cn(
                                'flex items-center gap-2 rounded-xl px-4 py-2.5 text-sm font-medium transition-all',
                                isChecking
                                    ? 'bg-primary/20 text-primary cursor-wait'
                                    : 'bg-primary text-primary-foreground shadow-primary/25 hover:shadow-primary/40 shadow-lg'
                            )}>
                            <RefreshCw className={cn('h-4 w-4', isChecking && 'animate-spin')} />
                            {isChecking ? '扫描中' : '立即扫描'}
                        </motion.button>

                        {/* Notifications */}
                        <div className="relative">
                            <button
                                onClick={() => setShowNotifications(!showNotifications)}
                                className="border-border bg-muted/50 hover:bg-muted relative rounded-xl border p-2.5 transition-all">
                                <Bell className="text-muted-foreground h-5 w-5" />
                                {unreadCount > 0 && (
                                    <span className="bg-destructive text-destructive-foreground absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full text-[10px] font-bold">
                                        {unreadCount > 9 ? '9+' : unreadCount}
                                    </span>
                                )}
                            </button>

                            <AnimatePresence>
                                {showNotifications && (
                                    <>
                                        {/* Backdrop for closing popover */}
                                        <div className="fixed inset-0 z-[-1]" onClick={() => setShowNotifications(false)} />
                                        <motion.div
                                            initial={{ opacity: 0, y: 8, scale: 0.96 }}
                                            animate={{ opacity: 1, y: 0, scale: 1 }}
                                            exit={{ opacity: 0, y: 8, scale: 0.96 }}
                                            className="border-border bg-popover/95 absolute top-14 right-0 w-80 overflow-hidden rounded-2xl border shadow-2xl backdrop-blur-2xl">
                                            <div className="border-border flex items-center justify-between border-b p-4">
                                                <h3 className="text-popover-foreground font-medium">通知</h3>
                                                {unreadCount > 0 && (
                                                    <button onClick={onMarkAllRead} className="text-primary hover:text-primary/80 text-xs">
                                                        全部已读
                                                    </button>
                                                )}
                                            </div>
                                            <div className="max-h-80 overflow-y-auto">
                                                {notifications.length === 0 ? (
                                                    <p className="text-muted-foreground py-8 text-center text-sm">暂无通知</p>
                                                ) : (
                                                    <div className="divide-border divide-y">
                                                        {notifications.slice(0, 10).map(n => (
                                                            <div
                                                                key={n.id}
                                                                className={cn(
                                                                    'p-4 transition-colors',
                                                                    n.isRead ? 'opacity-50' : 'hover:bg-muted/50'
                                                                )}>
                                                                <p className="text-popover-foreground text-sm font-medium">{n.title}</p>
                                                                <p className="text-muted-foreground mt-1 line-clamp-2 text-xs">{n.content}</p>
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
                </div>
            </div>
        </header>
    );
};

export default Header;
