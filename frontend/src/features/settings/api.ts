import type { AppConfig } from '@/types';
import { Get, Update } from '@wails/core/config/service.js';

export const settingsApi = {
    getConfig: () => Get(),
    updateConfig: (cfg: AppConfig) => Update(cfg)
};
