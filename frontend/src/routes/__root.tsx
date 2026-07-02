import { createRootRoute, Outlet } from '@tanstack/react-router';
import { Sidebar } from '@/components/Sidebar';
import { ContentHeader } from '@/components/ContentHeader';
import Toast from '@/components/Toast';
import { AppProvider, useApp } from '@/context/AppContext';

function RootComponent() {
    const {
        isChecking,
        handleManualCheck,
        unreadCount,
        notifications,
        showNotifications,
        setShowNotifications,
        handleMarkAllRead,
        toast,
        status
    } = useApp();

    return (
        <div className="bg-background flex h-screen overflow-hidden">
            <Sidebar status={status} />

            <div className="flex min-w-0 flex-1 flex-col">
                <ContentHeader
                    isChecking={isChecking}
                    onManualCheck={handleManualCheck}
                    unreadCount={unreadCount}
                    notifications={notifications}
                    showNotifications={showNotifications}
                    setShowNotifications={setShowNotifications}
                    onMarkAllRead={handleMarkAllRead}
                />

                <main className="flex-1 overflow-y-auto p-6">
                    <Outlet />
                </main>
            </div>

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
