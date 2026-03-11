import { drizzle } from 'drizzle-orm/better-sqlite3';
import Database from 'better-sqlite3';
import * as keywords from '@/models/keywords';
import * as hotspots from '@/models/hotspots';
import * as notifications from '@/models/notifications';
import * as settings from '@/models/settings';
import { DATABASE_CONFIG } from '@env';

const sqlite = new Database(DATABASE_CONFIG.connectUrl);
const db = drizzle(sqlite, {
    schema: {
        ...keywords,
        ...hotspots,
        ...notifications,
        ...settings
    },
    logger: true
});

export default db;
