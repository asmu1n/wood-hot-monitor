import { Get, Update } from '@wails/core/config/service.js';
import type { AppConfig } from '@wails/core/models';

export const settingsApi = {
    getConfig: () => Get(),
    updateConfig: (cfg: AppConfig) => Update(cfg)
};
