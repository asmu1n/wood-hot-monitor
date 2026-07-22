import type { AppConfig } from '@wails/config';
import { Get, Update } from '@wails/config/service';

export const settingsApi = {
    getConfig: () => Get(),
    updateConfig: (cfg: AppConfig) => Update(cfg)
};
