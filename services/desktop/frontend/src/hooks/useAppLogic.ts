import { useState, useEffect, useCallback, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { onNewHotSpot } from '@/features/hotspot/utils';
import { createKeyword, deleteKeyword, getAllKeywords, toggleKeyword } from '@/features/keyword/api';
import { checkHotSpot, getAllHotSpots, getHotSpotStatus, searchHotSpots } from '@/features/hotspot/api';
import { getAllNotifications, markAllAsRead } from '@/features/notifications/api';
import { subscribeToKeywords, unsubscribeFromKeywords } from '@/features/keyword/utils';
import { onNotification } from '@/features/notifications/utils';
import { attempt } from '@/utils/common';
import { defaultFilterState, type FilterState } from '@/components/FilterSortBar';
import type { Keyword, Hotspot } from '@repo/types';

export function useAppLogic() {
    const queryClient = useQueryClient();

    // UI Local State
    const [newKeyword, setNewKeyword] = useState('');
    const [searchQuery, setSearchQuery] = useState('');
    const [showNotifications, setShowNotifications] = useState(false);
    const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);
    const [dashboardFilters, setDashboardFilters] = useState<FilterState>({ ...defaultFilterState });
    const [searchFilters, setSearchFilters] = useState<FilterState>({ ...defaultFilterState });
    const [currentPage, setCurrentPage] = useState(1);
    const [searchResults, setSearchResults] = useState<Hotspot[]>([]);

    // 展开/折叠状态
    const [expandedReasons, setExpandedReasons] = useState<Set<string>>(new Set());
    const [expandedContents, setExpandedContents] = useState<Set<string>>(new Set());
    const [allReasonsExpanded, setAllReasonsExpanded] = useState(false);

    const showToast = useCallback((message: string, type: 'success' | 'error') => {
        setToast({ message, type });
        setTimeout(() => setToast(null), 3000);
    }, []);

    // --- Queries ---

    // Keywords Query
    const { data: keywordsRes } = useQuery({
        queryKey: ['keywords'],
        queryFn: getAllKeywords.getQueryFn
    });

    const keywords: Keyword[] = useMemo(() => keywordsRes?.data || [], [keywordsRes]);

    // Hotspots Query
    const hotspotParams = useMemo(
        () => ({
            limit: 20,
            page: currentPage,
            ...Object.fromEntries(Object.entries(dashboardFilters).filter(([, v]) => v != null && v !== ''))
        }),
        [dashboardFilters, currentPage]
    );

    const { data: hotspotsRes, isLoading: isHotspotsLoading } = useQuery({
        queryKey: ['hotspots', hotspotParams],
        queryFn: getAllHotSpots.getQueryFn
    });

    const hotSpots = useMemo(() => hotspotsRes?.data || [], [hotspotsRes]);
    const totalPages = useMemo(() => hotspotsRes?.total || 1, [hotspotsRes]);

    // Status Query
    const { data: statusRes } = useQuery({
        queryKey: ['status'],
        queryFn: getHotSpotStatus.getQueryFn
    });

    const status = useMemo(() => statusRes?.data || null, [statusRes]);

    // Notifications Query
    const notificationParams = useMemo(() => ({ limit: 20 }), []);

    const { data: notificationRes } = useQuery({
        queryKey: ['notifications', notificationParams],
        queryFn: getAllNotifications.getQueryFn
    });

    const notifications = useMemo(() => notificationRes?.data || [], [notificationRes]);
    const unreadCount = useMemo(() => notificationRes?.total || 0, [notificationRes]);

    // Combined Loading state
    const isLoading = isHotspotsLoading;

    // --- Mutations ---

    // 1. Add Keyword
    const addKeywordMutation = useMutation({
        mutationFn: (text: string) => createKeyword.request({ text }),
        onSuccess: async ({ data: keyword }) => {
            await queryClient.invalidateQueries({ queryKey: ['keywords'] });
            setNewKeyword('');
            showToast('关键词添加成功', 'success');
            subscribeToKeywords([keyword.text]);
        },
        onError: (error: any) => {
            showToast(error.message || '添加失败', 'error');
        }
    });

    const handleAddKeyword = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!newKeyword.trim()) {
            return;
        }

        addKeywordMutation.mutate(newKeyword.trim());
    };

    // 2. Delete Keyword
    const deleteKeywordMutation = useMutation({
        mutationFn: (keyword: Keyword) => deleteKeyword.request({ id: keyword.id }),
        onSuccess: async (_, keyword) => {
            unsubscribeFromKeywords([keyword.text]);
            await queryClient.invalidateQueries({ queryKey: ['keywords'] });
            showToast('关键词已删除', 'success');
        },
        onError: () => {
            showToast('删除失败', 'error');
        }
    });

    const handleDeleteKeyword = (keyword: Keyword) => {
        deleteKeywordMutation.mutate(keyword);
    };

    // 3. Toggle Keyword
    const toggleKeywordMutation = useMutation({
        mutationFn: (keyword: Keyword) => toggleKeyword.request({ id: keyword.id }),
        onSuccess: async ({ data: updatedKeyword }) => {
            if (updatedKeyword.isActive) {
                subscribeToKeywords([updatedKeyword.text]);
            } else {
                unsubscribeFromKeywords([updatedKeyword.text]);
            }

            await queryClient.invalidateQueries({ queryKey: ['keywords'] });
        },
        onError: () => {
            showToast('操作失败', 'error');
        }
    });

    const handleToggleKeyword = (keyword: Keyword) => {
        toggleKeywordMutation.mutate(keyword);
    };

    // 4. Mark All Read
    const markAllReadMutation = useMutation({
        mutationFn: () => markAllAsRead.request(),
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: ['notifications'] });
        },
        onError: error => {
            console.error('Failed to mark as read:', error);
        }
    });

    const handleMarkAllRead = () => {
        markAllReadMutation.mutate();
    };

    // 5. Manual Check
    const { mutate: manualCheck, isPending: isChecking } = useMutation({
        mutationFn: () => checkHotSpot.request(),
        onSuccess: () => {
            showToast('热点检查已触发', 'success');
            setTimeout(() => {
                void queryClient.invalidateQueries({ queryKey: ['hotspots'] });
            }, 5000);
        },
        onError: () => {
            showToast('触发失败', 'error');
        }
    });

    const handleManualCheck = () => {
        manualCheck();
    };

    // Manual search (keep it manual as it is typically a "search on click" action)
    const handleSearch = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!searchQuery.trim()) {
            return;
        }

        const [err, result] = await attempt(() => {
            return searchHotSpots.request({ query: searchQuery });
        });

        if (err) {
            showToast('搜索失败', 'error');

            return;
        }

        setSearchResults(result.data);
        showToast(`找到 ${result.data.length} 条结果`, 'success');
    };

    const handleDashboardFilterChange = useCallback((newFilters: FilterState) => {
        setDashboardFilters(newFilters);
        setCurrentPage(1);
    }, []);

    // --- Effects ---

    // 关键词订阅状态同步
    useEffect(() => {
        if (keywords.length > 0) {
            const activeKeywords = keywords.filter(k => k.isActive).map(k => k.text);

            if (activeKeywords.length > 0) {
                subscribeToKeywords(activeKeywords);
            }
        }
    }, [keywords]);

    // WebSocket 事件
    useEffect(() => {
        const unSubHotSpot = onNewHotSpot(async hotspot => {
            await queryClient.invalidateQueries({ queryKey: ['hotspots'] });
            showToast('发现新热点: ' + hotspot.title.slice(0, 30), 'success');
        });

        const unSubNotification = onNotification(() => {
            void queryClient.invalidateQueries({ queryKey: ['notifications'] });
        });

        return () => {
            unSubHotSpot();
            unSubNotification();
        };
    }, [queryClient, showToast]);

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

    // Exposed interface (maintaining compatibility)
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
        setDashboardFilters: handleDashboardFilterChange,
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
        loadData: () => void queryClient.invalidateQueries({ queryKey: ['hotspots'] }) // Compatibility with manual reloads
    };
}

export type AppLogic = ReturnType<typeof useAppLogic>;
