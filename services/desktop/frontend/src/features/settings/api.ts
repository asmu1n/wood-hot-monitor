import { BaseRequest } from '@/lib/request';

const BASE_URL = '/settings';

const getAllSettings = new BaseRequest<undefined, Record<string, string>>({ method: 'get', url: BASE_URL });

const updateSettings = new BaseRequest<Record<string, string>>({ method: 'put', url: BASE_URL });

export { getAllSettings, updateSettings };
