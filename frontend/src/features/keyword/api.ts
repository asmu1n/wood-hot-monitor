import { GetAll, GetByID, Create, Delete, Toggle } from '@wails/internal/module/keyword/service';

export const keywordApi = {
    getAll: () => GetAll(false),
    getById: (id: string) => GetByID(id),
    create: (text: string, category?: string) => Create(text, category ?? null),
    delete: (id: string) => Delete(id),
    toggle: (id: string) => Toggle(id)
};
