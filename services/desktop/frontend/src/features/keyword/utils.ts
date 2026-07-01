import { getSocket } from '@/lib/socket';

function subscribeToKeywords(keywords: string[]): void {
    const s = getSocket();

    s.emit('subscribe', keywords);
}

function unsubscribeFromKeywords(keywords: string[]): void {
    const s = getSocket();

    s.emit('unsubscribe', keywords);
}

export { subscribeToKeywords, unsubscribeFromKeywords };
