import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { keywordApi } from '@/features/keyword/api';
import { subscribeToKeywords, unsubscribeFromKeywords } from '@/features/keyword/utils';
import { useToast } from '@/hooks/useToast';
import type { Keyword } from '@/types';

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
        onSuccess: async keyword => {
            await queryClient.invalidateQueries({ queryKey: ['keywords'] });
            setNewKeyword('');
            showToast('关键词添加成功', 'success');

            if (keyword) {
                subscribeToKeywords([keyword.text]);
            }
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
        onSuccess: async (_, keyword) => {
            unsubscribeFromKeywords([keyword.text]);
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
        onSuccess: async updatedKeyword => {
            if (updatedKeyword) {
                if (updatedKeyword.isActive) {
                    subscribeToKeywords([updatedKeyword.text]);
                } else {
                    unsubscribeFromKeywords([updatedKeyword.text]);
                }
            }

            await queryClient.invalidateQueries({ queryKey: ['keywords'] });
        },
        onError: () => {
            showToast('操作失败', 'error');
        }
    });

    const handleToggleKeyword = (keyword: Keyword) => {
        toggleKeywordMutation.mutate(keyword);
    };

    useEffect(() => {
        if (keywords.length > 0) {
            const activeKeywords = keywords.filter((k: Keyword) => k.isActive).map((k: Keyword) => k.text);

            if (activeKeywords.length > 0) {
                subscribeToKeywords(activeKeywords);
            }
        }
    }, [keywords]);

    return {
        keywords,
        newKeyword,
        setNewKeyword,
        handleAddKeyword,
        handleDeleteKeyword,
        handleToggleKeyword
    };
}
