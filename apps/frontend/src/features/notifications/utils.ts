import { getSocket } from '@/lib/socket';

interface NotificationEvent {
    type: string;
    title: string;
    content: string;
    hotSpotId?: string;
    importance?: string;
}

function onNotification(callback: (notification: NotificationEvent) => void): () => void {
    const s = getSocket();

    s.on('notification', callback);

    return () => s.off('notification', callback);
}

export { onNotification };
