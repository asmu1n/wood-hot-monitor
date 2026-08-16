import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hotspotApi } from '@/features/hotspot/api';

export function useNotifications() {
    const queryClient = useQueryClient();
    const [showNotifications, setShowNotifications] = useState(false);

    const { data: unreadCount = 0 } = useQuery({
        queryKey: ['unreadCount'],
        queryFn: hotspotApi.unreadCount
    });

    const { data: notifications = [] } = useQuery({
        queryKey: ['notifications'],
        queryFn: () => hotspotApi.getNotifications(10)
    });

    const markAllReadMutation = useMutation({
        mutationFn: () => hotspotApi.markAllRead(),
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: ['notifications'] });
            void queryClient.invalidateQueries({ queryKey: ['unreadCount'] });
        },
        onError: error => {
            console.error('Failed to mark as read:', error);
        }
    });

    const handleMarkAllRead = () => {
        markAllReadMutation.mutate();
    };

    return {
        notifications,
        unreadCount,
        showNotifications,
        setShowNotifications,
        handleMarkAllRead
    };
}
