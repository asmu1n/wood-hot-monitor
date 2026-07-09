import type { Hotspot } from '@wails/core/models';
import { useState } from 'react';

// 展开热点信息卡片逻辑
export function useHotspotExpand() {
    const [expandedReasons, setExpandedReasons] = useState<Set<string>>(new Set());
    const [expandedContents, setExpandedContents] = useState<Set<string>>(new Set());
    const [allReasonsExpanded, setAllReasonsExpanded] = useState(false);

    const toggleReason = (id: string) => {
        setExpandedReasons(prev => {
            const next = new Set(prev);

            if (next.has(id)) {
                next.delete(id);
            } else {
                next.add(id);
            }

            return next;
        });
    };

    const toggleContent = (id: string) => {
        setExpandedContents(prev => {
            const next = new Set(prev);

            if (next.has(id)) {
                next.delete(id);
            } else {
                next.add(id);
            }

            return next;
        });
    };

    const toggleAllReasons = (list: Hotspot[]) => {
        if (allReasonsExpanded) {
            setExpandedReasons(new Set());
        } else {
            setExpandedReasons(new Set(list.filter(h => h.relevanceReason).map(h => h.id)));
        }

        setAllReasonsExpanded(!allReasonsExpanded);
    };

    return {
        expandedReasons,
        expandedContents,
        allReasonsExpanded,
        toggleReason,
        toggleContent,
        toggleAllReasons
    };
}
