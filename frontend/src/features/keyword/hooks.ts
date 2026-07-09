import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { keywordApi } from '@/features/keyword/api';
import { useToast } from '@/hooks/useToast';
import type { Keyword } from '@wails/core/models';

export function useKeywords() {
    const queryClient = useQueryClient();
    const { showToast } = useToast();
    const [newKeyword, setNewKeyword] = useState('');

    const { data: keywords = [] } = useQuery({
        queryKey: ['keywords'],
        queryFn: keywordApi.getAll
    });

    const addKeywordMutation = useMutation({
        mutationFn: (text: string) => keywordApi.create(text),
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: ['keywords'] });
            setNewKeyword('');
            showToast('关键词添加成功', 'success');
        },
        onError: (error: Error) => {
            showToast(error.message || '添加失败', 'error');
        }
    });

    const handleAddKeyword = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!newKeyword.trim()) {
            return;
        }

        addKeywordMutation.mutate(newKeyword.trim());
    };

    const deleteKeywordMutation = useMutation({
        mutationFn: (keyword: Keyword) => keywordApi.delete(keyword.id),
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: ['keywords'] });
            showToast('关键词已删除', 'success');
        },
        onError: () => {
            showToast('删除失败', 'error');
        }
    });

    const handleDeleteKeyword = (keyword: Keyword) => {
        deleteKeywordMutation.mutate(keyword);
    };

    const toggleKeywordMutation = useMutation({
        mutationFn: (keyword: Keyword) => keywordApi.toggle(keyword.id),
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: ['keywords'] });
        },
        onError: () => {
            showToast('操作失败', 'error');
        }
    });

    const handleToggleKeyword = (keyword: Keyword) => {
        toggleKeywordMutation.mutate(keyword);
    };

    return {
        keywords,
        newKeyword,
        setNewKeyword,
        handleAddKeyword,
        handleDeleteKeyword,
        handleToggleKeyword
    };
}
