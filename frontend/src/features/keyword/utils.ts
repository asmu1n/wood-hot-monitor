// No-op in desktop app: Go backend monitors active keywords from DB directly.

function subscribeToKeywords(_keywords: string[]): void {}

function unsubscribeFromKeywords(_keywords: string[]): void {}

export { subscribeToKeywords, unsubscribeFromKeywords };
