import { useEffect } from 'react';
import { createRootRoute, Outlet } from '@tanstack/react-router';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Sidebar } from '@/components/Sidebar';
import { ContentHeader } from '@/components/ContentHeader';
import { ToastProvider } from '@/hooks/useToast';
import { useCheckerStatus } from '@/features/hotspot/hooks/useCheckerStatus';
import { useNotifications } from '@/features/hotspot/hooks/useNotifications';
import { hotspotApi } from '@/features/hotspot/api';
import { onCheckComplete } from '@/features/hotspot/utils';
// import Toast from '@/components/Toast';

function RootComponent() {
    const queryClient = useQueryClient();
    // const { toast } = useToast();
    const { isChecking, handleManualCheck } = useCheckerStatus();
    const { unreadCount, notifications, showNotifications, setShowNotifications, handleMarkAllRead } = useNotifications();

    const { data: status = null } = useQuery({
        queryKey: ['status'],
        queryFn: hotspotApi.getStatus
    });

    // useEffect(() => {
    //     const unSubHotSpot = onNewHotSpot(async hotspot => {
    //         // await queryClient.invalidateQueries({ queryKey: ['hotspots'] });
    //         void queryClient.invalidateQueries({ queryKey: ['notifications'] });
    //         void queryClient.invalidateQueries({ queryKey: ['unreadCount'] });
    //         void queryClient.invalidateQueries({ queryKey: ['status'] });
    //         showToast('发现新热点: ' + hotspot.title.slice(0, 30), 'success');
    //     });

    //     return () => {
    //         unSubHotSpot();
    //     };
    // }, [queryClient, showToast]);

    useEffect(() => {
        const unSubCheckComplete = onCheckComplete(() => {
            void queryClient.invalidateQueries({ queryKey: ['hotspots'] });
            void queryClient.invalidateQueries({ queryKey: ['notifications'] });
            void queryClient.invalidateQueries({ queryKey: ['unreadCount'] });
            void queryClient.invalidateQueries({ queryKey: ['status'] });
        });

        return () => {
            unSubCheckComplete();
        };
    }, [queryClient]);

    return (
        <div className="bg-background flex h-screen">
            <Sidebar status={status} />

            <div className="flex h-full min-w-0 flex-1 flex-col">
                <ContentHeader
                    isChecking={isChecking}
                    onManualCheck={handleManualCheck}
                    unreadCount={unreadCount}
                    notifications={notifications || []}
                    showNotifications={showNotifications}
                    setShowNotifications={setShowNotifications}
                    onMarkAllRead={handleMarkAllRead}
                />

                <main className="min-h-0 px-6 py-2">
                    <Outlet />
                </main>
            </div>
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
