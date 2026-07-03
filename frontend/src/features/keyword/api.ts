import { GetAll, GetByID, Create, Delete, Toggle } from '@wails/biz/keyword/service.js';
import type { Keyword } from '@/types';

export const keywordApi = {
    getAll: () => GetAll(false) as Promise<Keyword[]>,
    getById: (id: string) => GetByID(id) as Promise<Keyword | null>,
    create: (text: string, category?: string) => Create(text, category ?? null) as Promise<Keyword | null>,
    delete: (id: string) => Delete(id),
    toggle: (id: string) => Toggle(id) as Promise<Keyword | null>
};
