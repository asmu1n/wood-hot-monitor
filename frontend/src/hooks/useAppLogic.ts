import { useState, useEffect, useCallback, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { onNewHotSpot } from '@/features/hotspot/utils';
import { keywordApi } from '@/features/keyword/api';
import { hotspotApi, type HotspotFilters } from '@/features/hotspot/api';
import { notificationApi } from '@/features/notifications/api';
import { subscribeToKeywords, unsubscribeFromKeywords } from '@/features/keyword/utils';
import { onNotification } from '@/features/notifications/utils';
import { attempt } from '@/utils/common';
import { defaultFilterState, type FilterState } from '@/components/FilterSortBar';
import type { Keyword, Hotspot } from '@/types';

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

    const [expandedReasons, setExpandedReasons] = useState<Set<string>>(new Set());
    const [expandedContents, setExpandedContents] = useState<Set<string>>(new Set());
    const [allReasonsExpanded, setAllReasonsExpanded] = useState(false);

    const showToast = useCallback((message: string, type: 'success' | 'error') => {
        setToast({ message, type });
        setTimeout(() => setToast(null), 3000);
    }, []);

    // --- Queries ---

    const { data: keywords = [] } = useQuery({
        queryKey: ['keywords'],
        queryFn: keywordApi.getAll
    });

    const hotspotParams: HotspotFilters = useMemo(
        () => ({
            limit: 20,
            page: currentPage,
            ...Object.fromEntries(Object.entries(dashboardFilters).filter(([, v]) => v != null && v !== ''))
        }),
        [dashboardFilters, currentPage]
    );

    const { data: hotspotsRes, isLoading: isHotspotsLoading } = useQuery({
        queryKey: ['hotspots', hotspotParams],
        queryFn: () => hotspotApi.getAll(hotspotParams)
    });

    const hotSpots = useMemo(() => hotspotsRes?.data ?? [], [hotspotsRes]);
    const totalPages = useMemo(() => hotspotsRes?.total ?? 1, [hotspotsRes]);

    const { data: status = null } = useQuery({
        queryKey: ['status'],
        queryFn: hotspotApi.getStatus
    });

    const notificationParams = useMemo(() => ({ limit: 20 }), []);

    const { data: notificationRes } = useQuery({
        queryKey: ['notifications', notificationParams],
        queryFn: () => notificationApi.getAll(notificationParams)
    });

    const notifications = useMemo(() => notificationRes?.data ?? [], [notificationRes]);
    const unreadCount = useMemo(() => notificationRes?.total ?? 0, [notificationRes]);

    const isLoading = isHotspotsLoading;

    // --- Mutations ---

    const addKeywordMutation = useMutation({
        mutationFn: (text: string) => keywordApi.create(text),
        onSuccess: async (keyword) => {
            await queryClient.invalidateQueries({ queryKey: ['keywords'] });
            setNewKeyword('');
            showToast('关键词添加成功', 'success');
            if (keyword) {
                subscribeToKeywords([keyword.text]);
            }
        },
        onError: (error: Error) => {
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

    const deleteKeywordMutation = useMutation({
        mutationFn: (keyword: Keyword) => keywordApi.delete(keyword.id),
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

    const toggleKeywordMutation = useMutation({
        mutationFn: (keyword: Keyword) => keywordApi.toggle(keyword.id),
        onSuccess: async (updatedKeyword) => {
            if (updatedKeyword) {
                if (updatedKeyword.isActive) {
                    subscribeToKeywords([updatedKeyword.text]);
                } else {
                    unsubscribeFromKeywords([updatedKeyword.text]);
                }
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

    const markAllReadMutation = useMutation({
        mutationFn: () => notificationApi.markAllAsRead(),
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: ['notifications'] });
        },
        onError: (error) => {
            console.error('Failed to mark as read:', error);
        }
    });

    const handleMarkAllRead = () => {
        markAllReadMutation.mutate();
    };

    const { mutate: manualCheck, isPending: isChecking } = useMutation({
        mutationFn: () => hotspotApi.check(),
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

    const handleSearch = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!searchQuery.trim()) {
            return;
        }

        const [err, result] = await attempt(() => {
            return hotspotApi.search(searchQuery);
        });

        if (err) {
            showToast('搜索失败', 'error');

            return;
        }

        setSearchResults(result);
        showToast(`找到 ${result.length} 条结果`, 'success');
    };

    const handleDashboardFilterChange = useCallback((newFilters: FilterState) => {
        setDashboardFilters(newFilters);
        setCurrentPage(1);
    }, []);

    // --- Effects ---

    useEffect(() => {
        if (keywords.length > 0) {
            const activeKeywords = keywords.filter((k: Keyword) => k.isActive).map((k: Keyword) => k.text);

            if (activeKeywords.length > 0) {
                subscribeToKeywords(activeKeywords);
            }
        }
    }, [keywords]);

    useEffect(() => {
        const unSubHotSpot = onNewHotSpot(async (hotspot) => {
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
        loadData: () => void queryClient.invalidateQueries({ queryKey: ['hotspots'] })
    };
}

export type AppLogic = ReturnType<typeof useAppLogic>;
