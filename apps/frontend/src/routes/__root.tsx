import { createRootRoute, Outlet, Link } from '@tanstack/react-router';
import { BackgroundBeams } from '@/components/BackgroundBeams';
import Header from '@/components/Header';
import Toast from '@/components/Toast';
import { AppProvider, useApp } from '@/context/AppContext';
import { Activity, Target, Search } from 'lucide-react';
import { cn, Spotlight } from '@repo/ui';

function RootComponent() {
    const { isChecking, handleManualCheck, unreadCount, notifications, showNotifications, setShowNotifications, handleMarkAllRead, toast } = useApp();

    return (
        <div className="bg-background relative min-h-screen overflow-hidden">
            {/* Background Effects */}
            <BackgroundBeams className="z-0" />
            <Spotlight className="-top-40 left-0 opacity-60 md:-top-20 md:left-60" fill="var(--primary)" />

            {/* Subtle gradient orbs */}
            <div className="bg-primary/5 pointer-events-none fixed top-0 right-0 h-[600px] w-[600px] rounded-full blur-3xl" />
            <div className="bg-accent/5 pointer-events-none fixed bottom-0 left-0 h-[400px] w-[400px] rounded-full blur-3xl" />

            {/* Header - Minimal & Clean */}
            <Header
                isChecking={isChecking}
                onManualCheck={handleManualCheck}
                unreadCount={unreadCount}
                notifications={notifications}
                showNotifications={showNotifications}
                setShowNotifications={setShowNotifications}
                onMarkAllRead={handleMarkAllRead}
            />

            {/* Main Content */}
            <main className="relative z-10 mx-auto max-w-6xl px-6 py-8">
                {/* Navigation Tabs */}
                <div className="mb-8 flex gap-2">
                    {(
                        [
                            { to: '/', label: '热点雷达', icon: Activity },
                            { to: '/keywords', label: '监控词', icon: Target },
                            { to: '/search', label: '搜索', icon: Search }
                        ] as const
                    ).map(({ to, label, icon: Icon }) => (
                        <Link
                            key={to}
                            to={to}
                            activeProps={{
                                className: 'bg-primary text-primary-foreground'
                            }}
                            inactiveProps={{
                                className: 'text-muted-foreground hover:bg-muted hover:text-foreground'
                            }}
                            className={cn('flex items-center gap-2 rounded-xl px-5 py-2.5 text-sm font-medium shadow-sm transition-all')}>
                            <Icon className="h-4 w-4" />
                            {label}
                        </Link>
                    ))}
                </div>

                <Outlet />
            </main>

            <Toast toast={toast} />
        </div>
    );
}

export const Route = createRootRoute({
    component: () => (
        <AppProvider>
            <RootComponent />
        </AppProvider>
    )
});
