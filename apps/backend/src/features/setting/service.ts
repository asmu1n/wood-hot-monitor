import db from '@/config/database';
import { settings as settingsTable } from '@/models/settings';
import { eq } from 'drizzle-orm';
import { NotFoundError } from '@/config/error';

export class SettingService {
    /**
     * 获取所有设置并转换为 Map
     */
    static async getSettingsMap() {
        const settings = await db.select().from(settingsTable);

        return settings.reduce(
            (acc: Record<string, string>, item) => {
                acc[item.key] = item.value;

                return acc;
            },
            {} as Record<string, string>
        );
    }

    /**
     * 获取单个设置
     */
    static async getByKey(key: string) {
        const result = await db.select().from(settingsTable).where(eq(settingsTable.key, key));

        if (result.length === 0) {
            throw new NotFoundError('设置未找到');
        }

        return result[0];
    }

    /**
     * 更新或创建单个设置 (Upsert)
     */
    static async upsert(key: string, value: any) {
        const result = await db
            .insert(settingsTable)
            .values({ key, value: String(value) })
            .onConflictDoUpdate({
                target: settingsTable.key,
                set: { value: String(value) }
            })
            .returning();

        return result[0];
    }

    /**
     * 批量更新设置
     */
    static async bulkUpdate(settings: Record<string, any>) {
        const updates = Object.entries(settings).map(([key, value]) => this.upsert(key, value));

        await Promise.all(updates);

        return { message: 'Settings updated' };
    }
}
