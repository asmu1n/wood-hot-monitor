import { BaseRequest } from '@/lib/request';
import type { Hotspot, Status } from '@repo/types';

const BASE_URL = '/hotspots';

const getAllHotSpots = new BaseRequest<
    {
        page?: number;
        limit?: number;
        source?: string;
        importance?: string;
        keywordId?: string;
        isReal?: string;
        timeRange?: string;
        timeFrom?: string;
        timeTo?: string;
        sortBy?: string;
        sortOrder?: 'desc' | 'asc';
    },
    Hotspot[],
    true
>({ method: 'get', url: BASE_URL });

const getHotSpotById = new BaseRequest<{ id: string }, Hotspot>(({ id }) => ({ method: 'get', url: `${BASE_URL}/${id}` }));

const getHotSpotStatus = new BaseRequest<undefined, Status>({ method: 'get', url: `${BASE_URL}/status` });

const searchHotSpots = new BaseRequest<{ query: string; sources?: string[] }, Hotspot[]>({ method: 'post', url: `${BASE_URL}/search/` });

const deleteHotSpot = new BaseRequest<{ id: string }, void>(({ id }) => ({ method: 'delete', url: `${BASE_URL}/${id}` }));

const checkHotSpot = new BaseRequest<undefined, { message: string }>({ method: 'post', url: '/check' });

export { getAllHotSpots, getHotSpotById, getHotSpotStatus, searchHotSpots, deleteHotSpot, checkHotSpot };
