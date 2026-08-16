import { createLazyFileRoute } from '@tanstack/react-router';
import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { motion, AnimatePresence } from 'framer-motion';
import { AlertTriangle, Eraser, Loader2, RotateCcw, Search, Trash2 } from 'lucide-react';

import { hotspotApi } from '@/features/hotspot/api';
import { useKeywords } from '@/features/keyword/hooks';
import { useToast } from '@/hooks/useToast';
import type { DeleteParams } from '@wails/internal/module/hotspot';
import { Importance } from '@wails/pkg/types';

const CLS_INPUT =
    'border-border bg-background text-foreground placeholder-muted-foreground/50 focus:border-primary/50 focus:ring-primary/20 w-full rounded-lg border px-3 py-2.5 text-sm transition-all focus:ring-2 focus:outline-none disabled:opacity-50';

const IMPORTANCE_OPTIONS = [
    { value: '', label: '不限' },
    { value: Importance.ImportanceLow, label: '低及以下（仅低）' },
    { value: Importance.ImportanceMedium, label: '中及以下' },
    { value: Importance.ImportanceHigh, label: '高及以下' },
    { value: Importance.ImportanceUrgent, label: '全部重要程度' }
] as const;

const READ_OPTIONS = [
    { value: '', label: '不限' },
    { value: 'true', label: '仅已读' },
    { value: 'false', label: '仅未读' }
] as const;

interface FormState {
    keywordId: string;
    publishedAt: string;
    createdAt: string;
    isRead: string;
    maxRelevance: string;
    maxImportance: string;
}

const EMPTY_FORM: FormState = {
    keywordId: '',
    publishedAt: '',
    createdAt: '',
    isRead: '',
    maxRelevance: '',
    maxImportance: ''
};

/** datetime-local → ISO；空串 → null */
function toTimeOrNull(value: string): string | null {
    if (!value) {
        return null;
    }

    const d = new Date(value);

    if (Number.isNaN(d.getTime())) {
        return null;
    }

    return d.toISOString();
}

function buildDeleteParams(form: FormState): DeleteParams | null {
    const params: DeleteParams = {
        keywordId: form.keywordId || null,
        publishedAt: toTimeOrNull(form.publishedAt),
        createdAt: toTimeOrNull(form.createdAt),
        isRead: form.isRead === '' ? null : form.isRead === 'true',
        maxRelevance: form.maxRelevance === '' ? null : Number(form.maxRelevance),
        maxImportance: (form.maxImportance || null) as DeleteParams['maxImportance']
    };

    const hasFilter =
        params.keywordId != null ||
        params.publishedAt != null ||
        params.createdAt != null ||
        params.isRead != null ||
        params.maxRelevance != null ||
        params.maxImportance != null;

    if (!hasFilter) {
        return null;
    }

    if (params.maxRelevance != null && (Number.isNaN(params.maxRelevance) || params.maxRelevance < 0 || params.maxRelevance > 100)) {
        return null;
    }

    return params;
}

function ManagePage() {
    const queryClient = useQueryClient();
    const { showToast } = useToast();
    const { keywords } = useKeywords();

    const [form, setForm] = useState<FormState>({ ...EMPTY_FORM });
    const [confirmOpen, setConfirmOpen] = useState(false);

    const deleteParams = useMemo(() => buildDeleteParams(form), [form]);

    const { data: matchCount, isFetching: isCounting, isError: countError } = useQuery({
        queryKey: ['hotspot-delete-count', deleteParams],
        queryFn: () => hotspotApi.countByDeleteParams(deleteParams!),
        enabled: deleteParams != null,
        placeholderData: previous => previous
    });

    const deleteMutation = useMutation({
        mutationFn: (params: DeleteParams) => hotspotApi.delete(params),
        onSuccess: async count => {
            setConfirmOpen(false);
            showToast(`已删除 ${count} 条热点`, 'success');
            await Promise.all([
                queryClient.invalidateQueries({ queryKey: ['hotspots'] }),
                queryClient.invalidateQueries({ queryKey: ['status'] }),
                queryClient.invalidateQueries({ queryKey: ['notifications'] }),
                queryClient.invalidateQueries({ queryKey: ['unreadCount'] }),
                queryClient.invalidateQueries({ queryKey: ['hotspot-delete-count'] })
            ]);
        },
        onError: (error: Error) => {
            showToast(error.message || '删除失败', 'error');
        }
    });

    const setField =
        <K extends keyof FormState>(key: K) =>
        (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
            setForm(prev => ({ ...prev, [key]: e.target.value }));
        };

    const handleReset = () => {
        setForm({ ...EMPTY_FORM });
        setConfirmOpen(false);
    };

    const canDelete = deleteParams != null && (matchCount ?? 0) > 0 && !deleteMutation.isPending;

    const activeFilters = useMemo(() => {
        const tags: string[] = [];

        if (form.keywordId) {
            const kw = keywords.find(k => k.id === form.keywordId);
            tags.push(`关键词: ${kw?.text ?? form.keywordId.slice(0, 8)}`);
        }

        if (form.publishedAt) {
            tags.push(`发布不晚于 ${form.publishedAt.replace('T', ' ')}`);
        }

        if (form.createdAt) {
            tags.push(`创建不晚于 ${form.createdAt.replace('T', ' ')}`);
        }

        if (form.isRead === 'true') {
            tags.push('仅已读');
        } else if (form.isRead === 'false') {
            tags.push('仅未读');
        }

        if (form.maxRelevance !== '') {
            tags.push(`关联度 ≤ ${form.maxRelevance}`);
        }

        if (form.maxImportance) {
            const label = IMPORTANCE_OPTIONS.find(o => o.value === form.maxImportance)?.label ?? form.maxImportance;
            tags.push(`重要程度: ${label}`);
        }

        return tags;
    }, [form, keywords]);

    return (
        <div className="mx-auto h-full max-w-3xl space-y-6 overflow-y-auto pb-8 scrollbar-none">
            <div className="border-border bg-muted/20 rounded-xl border border-dashed p-4">
                <div className="flex items-start gap-3">
                    <div className="bg-destructive/10 text-destructive mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg">
                        <Eraser className="h-4 w-4" />
                    </div>
                    <div>
                        <h2 className="text-foreground text-sm font-semibold">条件批量清理</h2>
                        <p className="text-muted-foreground mt-1 text-xs leading-relaxed">
                            按关键词、时间、已读状态、关联度与重要程度筛选后删除。时间条件为「不晚于」所选时刻；关联度与重要程度为「不高于」所选上限。至少选择一个条件，删除前可预览匹配条数。
                        </p>
                    </div>
                </div>
            </div>

            <section className="border-border bg-muted/30 space-y-4 rounded-xl border p-5">
                <h3 className="text-foreground flex items-center gap-2 text-sm font-semibold">
                    <Search className="text-primary h-4 w-4" />
                    筛选条件
                </h3>

                <div className="grid gap-4 sm:grid-cols-2">
                    <Field label="关键词">
                        <select value={form.keywordId} onChange={setField('keywordId')} className={CLS_INPUT}>
                            <option value="">不限</option>
                            {keywords.map(k => (
                                <option key={k.id} value={k.id}>
                                    {k.text}
                                    {k.isActive ? '' : '（已停用）'}
                                </option>
                            ))}
                        </select>
                    </Field>

                    <Field label="已读状态">
                        <select value={form.isRead} onChange={setField('isRead')} className={CLS_INPUT}>
                            {READ_OPTIONS.map(o => (
                                <option key={o.value || 'any'} value={o.value}>
                                    {o.label}
                                </option>
                            ))}
                        </select>
                    </Field>

                    <Field label="发布时间不晚于" hint="删除该时刻及更早发布的热点">
                        <input type="datetime-local" value={form.publishedAt} onChange={setField('publishedAt')} className={CLS_INPUT} />
                    </Field>

                    <Field label="创建时间不晚于" hint="删除该时刻及更早入库的热点">
                        <input type="datetime-local" value={form.createdAt} onChange={setField('createdAt')} className={CLS_INPUT} />
                    </Field>

                    <Field label="最大关联度" hint="删除关联度 ≤ 该值的热点（0-100）">
                        <input
                            type="number"
                            min={0}
                            max={100}
                            placeholder="例如 40"
                            value={form.maxRelevance}
                            onChange={setField('maxRelevance')}
                            className={CLS_INPUT}
                        />
                    </Field>

                    <Field label="最大重要程度" hint="删除不高于该等级的热点">
                        <select value={form.maxImportance} onChange={setField('maxImportance')} className={CLS_INPUT}>
                            {IMPORTANCE_OPTIONS.map(o => (
                                <option key={o.value || 'any'} value={o.value}>
                                    {o.label}
                                </option>
                            ))}
                        </select>
                    </Field>
                </div>

                {activeFilters.length > 0 && (
                    <div className="flex flex-wrap gap-2 pt-1">
                        {activeFilters.map(tag => (
                            <span
                                key={tag}
                                className="border-primary/20 bg-primary/10 text-primary inline-flex items-center rounded-md border px-2 py-1 text-[11px] font-medium">
                                {tag}
                            </span>
                        ))}
                    </div>
                )}
            </section>

            <section className="border-border bg-muted/30 flex flex-col gap-4 rounded-xl border p-5 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <p className="text-muted-foreground text-xs">匹配条数</p>
                    <div className="mt-1 flex items-baseline gap-2">
                        {deleteParams == null ? (
                            <p className="text-muted-foreground text-2xl font-semibold tabular-nums">—</p>
                        ) : isCounting && matchCount == null ? (
                            <Loader2 className="text-muted-foreground h-6 w-6 animate-spin" />
                        ) : countError ? (
                            <p className="text-destructive text-sm">预览失败，请检查条件</p>
                        ) : (
                            <>
                                <p className="text-foreground text-2xl font-semibold tabular-nums">{matchCount ?? 0}</p>
                                <span className="text-muted-foreground text-xs">条热点将被删除</span>
                                {isCounting && <Loader2 className="text-muted-foreground h-3.5 w-3.5 animate-spin" />}
                            </>
                        )}
                    </div>
                    {deleteParams == null && <p className="text-muted-foreground mt-1 text-xs">请至少选择一个筛选条件</p>}
                </div>

                <div className="flex flex-wrap items-center gap-2">
                    <button
                        type="button"
                        onClick={handleReset}
                        className="border-border text-muted-foreground hover:bg-muted hover:text-foreground flex items-center gap-1.5 rounded-lg border px-3.5 py-2 text-sm transition-colors">
                        <RotateCcw className="h-3.5 w-3.5" />
                        重置
                    </button>
                    <button
                        type="button"
                        disabled={!canDelete}
                        onClick={() => setConfirmOpen(true)}
                        className="bg-destructive text-white hover:bg-destructive/90 flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm font-medium transition-colors disabled:pointer-events-none disabled:opacity-40">
                        <Trash2 className="h-3.5 w-3.5" />
                        删除匹配项
                    </button>
                </div>
            </section>

            <AnimatePresence>
                {confirmOpen && (
                    <>
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            className="fixed inset-0 z-50 bg-black/40 backdrop-blur-[2px]"
                            onClick={() => !deleteMutation.isPending && setConfirmOpen(false)}
                        />
                        <motion.div
                            initial={{ opacity: 0, y: 12, scale: 0.97 }}
                            animate={{ opacity: 1, y: 0, scale: 1 }}
                            exit={{ opacity: 0, y: 12, scale: 0.97 }}
                            className="border-border bg-popover fixed top-1/2 left-1/2 z-50 w-[min(92vw,28rem)] -translate-x-1/2 -translate-y-1/2 rounded-2xl border p-5 shadow-2xl">
                            <div className="flex items-start gap-3">
                                <div className="bg-destructive/10 text-destructive flex h-10 w-10 shrink-0 items-center justify-center rounded-full">
                                    <AlertTriangle className="h-5 w-5" />
                                </div>
                                <div className="min-w-0 flex-1">
                                    <h4 className="text-popover-foreground text-base font-semibold">确认删除？</h4>
                                    <p className="text-muted-foreground mt-1.5 text-sm leading-relaxed">
                                        将永久删除 <span className="text-destructive font-semibold">{matchCount ?? 0}</span> 条匹配热点，此操作不可撤销。
                                    </p>
                                    {activeFilters.length > 0 && (
                                        <ul className="text-muted-foreground mt-3 list-inside list-disc space-y-0.5 text-xs">
                                            {activeFilters.map(t => (
                                                <li key={t}>{t}</li>
                                            ))}
                                        </ul>
                                    )}
                                </div>
                            </div>
                            <div className="mt-5 flex justify-end gap-2">
                                <button
                                    type="button"
                                    disabled={deleteMutation.isPending}
                                    onClick={() => setConfirmOpen(false)}
                                    className="border-border text-muted-foreground hover:bg-muted rounded-lg border px-3.5 py-2 text-sm transition-colors disabled:opacity-50">
                                    取消
                                </button>
                                <button
                                    type="button"
                                    disabled={deleteMutation.isPending || deleteParams == null}
                                    onClick={() => deleteParams && deleteMutation.mutate(deleteParams)}
                                    className="bg-destructive text-white hover:bg-destructive/90 flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm font-medium transition-colors disabled:opacity-50">
                                    {deleteMutation.isPending ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Trash2 className="h-3.5 w-3.5" />}
                                    确认删除
                                </button>
                            </div>
                        </motion.div>
                    </>
                )}
            </AnimatePresence>
        </div>
    );
}

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
    return (
        <div>
            <label className="text-foreground mb-1.5 block text-sm font-medium">
                {label}
                {hint && <span className="text-muted-foreground ml-2 text-xs font-normal">({hint})</span>}
            </label>
            {children}
        </div>
    );
}

export const Route = createLazyFileRoute('/manage')({
    component: ManagePage
});
