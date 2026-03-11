/**
 * 根据优先级或热度对热点进行内存排序
 */
function sortHotspots(hotspots: any[], sortBy: string, sortOrder: 'asc' | 'desc') {
    const sorted = [...hotspots];

    sorted.sort((a, b) => {
        let valA, valB;

        if (sortBy === 'importance') {
            const importanceMap: Record<string, number> = { urgent: 3, high: 2, medium: 1, low: 0 };

            valA = importanceMap[a.importance] || 0;
            valB = importanceMap[b.importance] || 0;
        } else if (sortBy === 'hot') {
            // 热度计算逻辑：(浏览数*0.2 + 点赞数*0.5 + 评论数*0.3)
            valA = (a.viewCount || 0) * 0.2 + (a.likeCount || 0) * 0.5 + (a.commentCount || 0) * 0.3;
            valB = (b.viewCount || 0) * 0.2 + (b.likeCount || 0) * 0.5 + (b.commentCount || 0) * 0.3;
        } else {
            valA = a[sortBy];
            valB = b[sortBy];
        }

        if (valA < valB) {
            return sortOrder === 'asc' ? -1 : 1;
        }

        if (valA > valB) {
            return sortOrder === 'asc' ? 1 : -1;
        }

        return 0;
    });

    return sorted;
}

export { sortHotspots };
