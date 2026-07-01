import { Events } from '@wailsio/runtime';

interface NotificationEvent {
    type: string;
    title: string;
    content: string;
    hotSpotId?: string;
    importance?: string;
}

function onNotification(callback: (notification: NotificationEvent) => void): () => void {
    const cancel = Events.On('notification', (event: { data: NotificationEvent }) => {
        callback(event.data);
    });

    return () => cancel();
}

export { onNotification };
