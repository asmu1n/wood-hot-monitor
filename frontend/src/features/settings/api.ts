import type { AppConfig } from '@wails/internal/config';
import { Get, Update } from '@wails/internal/config/service';

export const settingsApi = {
    getConfig: () => Get(),
    updateConfig: (cfg: AppConfig) => Update(cfg)
};
