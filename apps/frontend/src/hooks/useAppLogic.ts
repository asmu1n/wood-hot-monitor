import { useState, useEffect, useCallback } from 'react';
import { onNewHotSpot } from '@/features/hotspot/utils';
import { createKeyword, deleteKeyword, getAllKeywords, toggleKeyword } from '@/features/keyword/api';
import { checkHotSpot, getAllHotSpots, getHotSpotStatus, searchHotSpots } from '@/features/hotspot/api';
import { getAllNotifications, markAllAsRead } from '@/features/notifications/api';
import { subscribeToKeywords } from '@/features/keyword/utils';
import { onNotification } from '@/features/notifications/utils';
import { attempt } from '@/utils/common';
import { defaultFilterState, type FilterState } from '@/components/FilterSortBar';

export function useAppLogic() {
    const [keywords, setKeywords] = useState<Keyword[]>([]);
    const [hotSpots, setHotSpots] = useState<Hotspot[]>([]);
    const [status, setStatus] = useState<Status | null>(null);
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [unreadCount, setUnreadCount] = useState(0);

    const [newKeyword, setNewKeyword] = useState('');
    const [searchQuery, setSearchQuery] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [isChecking, setIsChecking] = useState(false);
    const [showNotifications, setShowNotifications] = useState(false);
    const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);
    const [dashboardFilters, setDashboardFilters] = useState<FilterState>({ ...defaultFilterState });
    const [searchFilters, setSearchFilters] = useState<FilterState>({ ...defaultFilterState });
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [searchResults, setSearchResults] = useState<Hotspot[]>([]);

    // 展开/折叠状态
    const [expandedReasons, setExpandedReasons] = useState<Set<string>>(new Set());
    const [expandedContents, setExpandedContents] = useState<Set<string>>(new Set());
    const [allReasonsExpanded, setAllReasonsExpanded] = useState(false);

    const showToast = useCallback((message: string, type: 'success' | 'error') => {
        setToast({ message, type });
        setTimeout(() => setToast(null), 3000);
    }, []);

    // 加载数据
    const loadData = useCallback(async () => {
        setIsLoading(true);

        try {
            const filterParams: Record<string, string | number> = {
                limit: 20,
                page: currentPage
            };

            // Apply dashboard filters
            if (dashboardFilters.source) {
                filterParams.source = dashboardFilters.source;
            }

            if (dashboardFilters.importance) {
                filterParams.importance = dashboardFilters.importance;
            }

            if (dashboardFilters.keywordId) {
                filterParams.keywordId = dashboardFilters.keywordId;
            }

            if (dashboardFilters.timeRange) {
                filterParams.timeRange = dashboardFilters.timeRange;
            }

            if (dashboardFilters.isReal) {
                filterParams.isReal = dashboardFilters.isReal;
            }

            if (dashboardFilters.sortBy) {
                filterParams.sortBy = dashboardFilters.sortBy;
            }

            if (dashboardFilters.sortOrder) {
                filterParams.sortOrder = dashboardFilters.sortOrder;
            }

            const [keywordsRes, hotSpotsRes, statusRes, notificationRes] = await Promise.all([
                getAllKeywords.request(),
                getAllHotSpots.request(filterParams),
                getHotSpotStatus.request(),
                getAllNotifications.request({ limit: 20 })
            ]);

            setKeywords(keywordsRes.data);
            setHotSpots(hotSpotsRes.data);
            setTotalPages(hotSpotsRes.total);
            setStatus(statusRes.data);
            setNotifications(notificationRes.data);
            setUnreadCount(notificationRes.total);

            // 订阅关键词
            const activeKeywords = keywordsRes.data.filter(k => k.isActive).map(k => k.text);

            if (activeKeywords.length > 0) {
                subscribeToKeywords(activeKeywords);
            }
        } catch (error) {
            console.error('Failed to load data:', error);
        } finally {
            setIsLoading(false);
        }
    }, [dashboardFilters, currentPage]);

    // 当筛选条件变化时重置页码
    useEffect(() => {
        setCurrentPage(1);
    }, [dashboardFilters]);

    useEffect(() => {
        void loadData();
    }, [loadData]);

    // WebSocket 事件
    useEffect(() => {
        const unSubHotSpot = onNewHotSpot(async hotspot => {
            setHotSpots(prev => [hotspot as Hotspot, ...prev.slice(0, 19)]);
            showToast('发现新热点: ' + (hotspot as Hotspot).title.slice(0, 30), 'success');
            await loadData();
        });

        const unSubNotification = onNotification(() => {
            setUnreadCount(prev => prev + 1);
        });

        return () => {
            unSubHotSpot();
            unSubNotification();
        };
    }, [loadData, showToast]);

    // 添加关键词
    const handleAddKeyword = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!newKeyword.trim()) {
            return;
        }

        try {
            const { data: keyword } = await createKeyword.request({ text: newKeyword.trim() });

            setKeywords(prev => [keyword, ...prev]);
            setNewKeyword('');
            showToast('关键词添加成功', 'success');
            subscribeToKeywords([keyword.text]);
        } catch (error: any) {
            showToast(error.message || '添加失败', 'error');
        }
    };

    // 删除关键词
    const handleDeleteKeyword = async (id: string) => {
        const [err] = await attempt(() => deleteKeyword.request({ id }));

        if (err) {
            showToast('删除失败', 'error');

            return;
        }

        setKeywords(prev => prev.filter(k => k.id !== id));
        showToast('关键词已删除', 'success');
    };

    // 切换关键词状态
    const handleToggleKeyword = async (id: string) => {
        const [err, result] = await attempt(() => toggleKeyword.request({ id }));

        if (err) {
            showToast('操作失败', 'error');

            return;
        }

        setKeywords(prev => prev.map(k => (k.id === id ? result.data : k)));
    };

    // 手动搜索
    const handleSearch = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!searchQuery.trim()) {
            return;
        }

        const [err, result] = await attempt(() => {
            setIsLoading(true);

            return searchHotSpots.request({ query: searchQuery });
        });

        setIsLoading(false);

        if (err) {
            showToast('搜索失败', 'error');

            return;
        }

        setSearchResults(result.data);
        showToast(`找到 ${result.data.length} 条结果`, 'success');
    };

    // 手动触发检查
    const handleManualCheck = async () => {
        const [err] = await attempt(() => {
            setIsChecking(true);

            return checkHotSpot.request();
        });

        setIsChecking(false);

        if (err) {
            showToast('触发失败', 'error');

            return;
        }

        showToast('热点检查已触发', 'success');
        setTimeout(() => void loadData(), 5000);
    };

    // 标记通知为已读
    const handleMarkAllRead = async () => {
        const [err] = await attempt(() => markAllAsRead.request());

        if (err) {
            console.error('Failed to mark as read:', err);

            return;
        }

        setUnreadCount(0);
        setNotifications(prev => prev.map(n => ({ ...n, isRead: true })));
    };

    // 展开/折叠功能
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
        keywords,
        hotSpots,
        status,
        notifications,
        unreadCount,
        newKeyword,
        setNewKeyword,
        searchQuery,
        setSearchQuery,
        isLoading,
        isChecking,
        showNotifications,
        setShowNotifications,
        toast,
        dashboardFilters,
        setDashboardFilters,
        searchFilters,
        setSearchFilters,
        currentPage,
        setCurrentPage,
        totalPages,
        searchResults,
        expandedReasons,
        expandedContents,
        allReasonsExpanded,
        handleAddKeyword,
        handleDeleteKeyword,
        handleToggleKeyword,
        handleSearch,
        handleManualCheck,
        handleMarkAllRead,
        toggleReason,
        toggleContent,
        toggleAllReasons,
        loadData
    };
}

export type AppLogic = ReturnType<typeof useAppLogic>;
