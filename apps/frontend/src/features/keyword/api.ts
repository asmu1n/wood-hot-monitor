import { BaseRequest } from '@/lib/request';

const BASE_URL = '/keywords';

const getAllKeywords = new BaseRequest<undefined, Keyword[], true>({
    method: 'get',
    url: BASE_URL
});

const getKeywordById = new BaseRequest<{ id: string }, Keyword>(({ id }) => ({
    method: 'get',
    url: `${BASE_URL}/${id}`
}));

const createKeyword = new BaseRequest<
    {
        text: string;
        category?: string;
    },
    Keyword
>({
    method: 'POST',
    url: BASE_URL
});

const updateKeyword = new BaseRequest<{ id: string }, Keyword>(({ id }) => ({
    method: 'PUT',
    url: `${BASE_URL}/${id}`
}));

const deleteKeyword = new BaseRequest<{ id: string }, void>(({ id }) => ({
    method: 'DELETE',
    url: `${BASE_URL}/${id}`
}));

const toggleKeyword = new BaseRequest<{ id: string }, Keyword>(({ id }) => ({
    method: 'PATCH',
    url: `${BASE_URL}/${id}/toggle`
}));

export { getAllKeywords, getKeywordById, createKeyword, updateKeyword, deleteKeyword, toggleKeyword };
