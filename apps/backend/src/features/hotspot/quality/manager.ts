import { QualityScorer, MetricInput, ScoreSnapshot } from './scoring';

export interface RatedHotspot extends MetricInput {
    scoreDetail: ScoreSnapshot;
    finalScore: number;
}

export class QualityManager {
    /**
     * 批量过滤并评分
     * @param items 原始抓取项
     */
    static process<T extends MetricInput>(items: T[]): (T & { scoreDetail: ScoreSnapshot; finalScore: number })[] {
        const processed = items
            .map(item => {
                const filterResult = QualityScorer.shouldFilter(item);

                if (!filterResult.pass) {
                    return null;
                }

                const scoreDetail = QualityScorer.calculateScore(item);

                return {
                    ...item,
                    scoreDetail,
                    finalScore: scoreDetail.total
                } as T & { scoreDetail: ScoreSnapshot; finalScore: number };
            })
            .filter((item): item is T & { scoreDetail: ScoreSnapshot; finalScore: number } => item !== null);

        // 按得分从高到低排序
        return processed.sort((a, b) => b.finalScore - a.finalScore);
    }
}
