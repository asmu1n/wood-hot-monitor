import { useState, useMemo, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import { hotspotApi, type HotspotFilters } from '@/features/hotspot/api';
import { defaultFilterState, type FilterState } from '@/components/FilterSortBar';

const LIMIT_COUNT = 20;

export function useHotspots() {
    const [dashboardFilters, setDashboardFilters] = useState<FilterState>({ ...defaultFilterState });
    const [currentPage, setCurrentPage] = useState(1);

    const hotspotParams: HotspotFilters = useMemo(
        () => ({
            limit: LIMIT_COUNT,
            page: currentPage,
            ...Object.fromEntries(Object.entries(dashboardFilters).filter(([, v]) => v != null && v !== ''))
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
