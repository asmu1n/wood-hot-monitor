import { useState, useMemo, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import { hotspotApi } from '@/features/hotspot/api';
import { defaultFilterState, type FilterState } from '@/components/FilterSortBar';
import type { GetAllParams } from '@wails/biz/hotspot';

const LIMIT_COUNT = 20;

// 热点信息条件分页查询
export function useHotspots() {
    const [dashboardFilters, setDashboardFilters] = useState<FilterState>({ ...defaultFilterState });
    const [currentPage, setCurrentPage] = useState(1);

    const hotspotParams: GetAllParams = useMemo(
        () => ({
            limit: LIMIT_COUNT,
            page: currentPage,
            timeFrom: null,
            timeTo: null,
            ...dashboardFilters
        }),
        [dashboardFilters, currentPage]
    );

    const { data: hotspotsRes, isLoading } = useQuery({
        queryKey: ['hotspots', hotspotParams],
        queryFn: () => hotspotApi.getAll(hotspotParams)
    });

    const { data: status = null } = useQuery({
        queryKey: ['status'],
        queryFn: hotspotApi.getStatus
    });

    const hotSpots = useMemo(() => hotspotsRes?.data ?? [], [hotspotsRes]);
    const totalPages = useMemo(() => Math.ceil((hotspotsRes?.total ?? 1) / (hotspotsRes?.limit ?? LIMIT_COUNT)), [hotspotsRes]);

    const handleDashboardFilterChange = useCallback((newFilters: FilterState) => {
        setDashboardFilters(newFilters);
        setCurrentPage(1);
    }, []);

    return {
        hotSpots,
        isLoading,
        status,
        dashboardFilters,
        currentPage,
        totalPages,
        setDashboardFilters: handleDashboardFilterChange,
        setCurrentPage
    };
}
