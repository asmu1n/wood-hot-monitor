import { GetAll, BulkUpdate } from '@wails/setting/service.js';

export const settingsApi = {
    getAll: () => GetAll() as Promise<Record<string, string>>,
    update: (settings: Record<string, string>) => BulkUpdate(settings)
};
