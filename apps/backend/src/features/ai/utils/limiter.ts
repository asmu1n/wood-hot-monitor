/**
 * 频率限制器：基于时间的限流，保证两次请求之间至少间隔指定的时间。
 */
class RateLimiter {
    private lastRequestTime = 0;
    private minInterval: number;

    constructor(minIntervalMs: number = 5000) {
        this.minInterval = minIntervalMs;
    }

    async wait(): Promise<void> {
        const elapsed = Date.now() - this.lastRequestTime;

        if (elapsed < this.minInterval) {
            await new Promise(resolve => setTimeout(resolve, this.minInterval - elapsed));
        }

        this.lastRequestTime = Date.now();
    }
}

/**
 * 并发限制器：使用队列机制限制最大并发请求数，避免瞬间请求过多导致 429。
 */
class ConcurrencyLimiter {
    private maxConcurrent: number;
    private currentConcurrent: number = 0;
    private queue: (() => void)[] = [];

    constructor(maxConcurrent: number = 5) {
        this.maxConcurrent = maxConcurrent;
    }

    async run<T>(task: () => Promise<T>): Promise<T> {
        if (this.currentConcurrent >= this.maxConcurrent) {
            // 达到并发上限，进入队列等待
            await new Promise<void>(resolve => this.queue.push(resolve));
        }

        this.currentConcurrent++;

        try {
            return await task();
        } finally {
            this.currentConcurrent--;

            if (this.queue.length > 0) {
                // 唤醒队列中的下一个任务
                const next = this.queue.shift();

                if (next) {
                    next();
                }
            }
        }
    }
}
export { RateLimiter, ConcurrencyLimiter };
