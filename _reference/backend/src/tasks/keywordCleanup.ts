import db from '@/config/database';
import { keywordExpansions } from '@/models/keywordExpansions';

export function clearKeywordExpansions() {
    return db.delete(keywordExpansions);
}
