import { useEffect } from 'react';
import { createRootRoute, Outlet } from '@tanstack/react-router';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Sidebar } from '@/components/Sidebar';
import { ContentHeader } from '@/components/ContentHeader';
import Toast from '@/components/Toast';
import { ToastProvider, useToast } from '@/hooks/useToast';
import { useCheckerStatus } from '@/features/hotspot/hooks/useCheckerStatus';
import { useNotifications } from '@/features/hotspot/hooks/useNotifications';
import { hotspotApi } from '@/features/hotspot/api';
import { onNewHotSpot, onCheckComplete } from '@/features/hotspot/utils';

function RootComponent() {
    const queryClient = useQueryClient();
    const { toast, showToast } = useToast();
    const { isChecking, handleManualCheck } = useCheckerStatus();
    const { unreadCount, notifications, showNotifications, setShowNotifications, handleMarkAllRead } = useNotifications();

    const { data: status = null } = useQuery({
        queryKey: ['status'],
        queryFn: hotspotApi.getStatus
    });

    useEffect(() => {
        const unSubHotSpot = onNewHotSpot(async hotspot => {
            await queryClient.invalidateQueries({ queryKey: ['hotspots'] });
            void queryClient.invalidateQueries({ queryKey: ['notifications'] });
            void queryClient.invalidateQueries({ queryKey: ['unreadCount'] });
            showToast('发现新热点: ' + hotspot.title.slice(0, 30), 'success');
        });

        return () => {
            unSubHotSpot();
        };
    }, [queryClient, showToast]);

    useEffect(() => {
        const unSubCheckComplete = onCheckComplete(() => {
            void queryClient.invalidateQueries({ queryKey: ['hotspots'] });
            void queryClient.invalidateQueries({ queryKey: ['notifications'] });
            void queryClient.invalidateQueries({ queryKey: ['unreadCount'] });
        });

        return () => {
            unSubCheckComplete();
        };
    }, [queryClient]);

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
        <ToastProvider>
            <RootComponent />
        </ToastProvider>
    )
});
