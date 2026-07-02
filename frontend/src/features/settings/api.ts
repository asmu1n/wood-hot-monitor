import { Get, UpdateSettings } from '@wails/config/service.js';

export const settingsApi = {
    getAll: async () => {
        const cfg = await Get();
        return cfg.settings as Record<string, any>;
    },
    update: (settings: Record<string, any>) => UpdateSettings(settings)
};
