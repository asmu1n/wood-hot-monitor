import { createLazyFileRoute } from '@tanstack/react-router';
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Save, Brain, Clock, Mail, Key, Loader2, Bell } from 'lucide-react';
import { settingsApi } from '@/features/settings/api';
import { useToast } from '@/hooks/useToast';
import { Importance } from 'bindings/wood-hot-monitor/pkg/types';

const IMPORTANCE_OPTIONS = [
    { value: Importance.ImportanceLow, label: '低（全部新热点）' },
    { value: Importance.ImportanceMedium, label: '中及以上' },
    { value: Importance.ImportanceHigh, label: '高及以上' },
    { value: Importance.ImportanceUrgent, label: '仅紧急' }
] as const;

interface FormState {
    llmModel: string;
    llmApiKey: string;
    llmBaseUrl: string;
    checkInterval: number;
    emailAddress: string;
    twitterApiKey: string;
    resendApiKey: string;
    osNotifyEnabled: boolean;
    osNotifyMinImportance: Importance;
}

const CLS_INPUT =
    'border-border bg-background text-foreground placeholder-muted-foreground/50 focus:border-primary/50 focus:ring-primary/20 w-full rounded-lg border px-3 py-2.5 text-sm transition-all focus:ring-2 focus:outline-none';

// 1. 父组件：负责数据获取和加载状态
function SettingsPage() {
    const { data: config, isLoading } = useQuery({
        queryKey: ['config'],
        queryFn: settingsApi.getConfig
    });

    if (isLoading) {
        return (
            <div className="flex items-center justify-center py-20">
                <Loader2 className="text-muted-foreground h-6 w-6 animate-spin" />
            </div>
        );
    }

    // 当数据加载完毕后，再去挂载表单组件
    return <SettingsForm config={config} />;
}

// 2. 子组件：专注于表单逻辑处理
function SettingsForm({ config }: { config: any }) {
    const queryClient = useQueryClient();
    const { showToast } = useToast();

    const [form, setForm] = useState<FormState>(() => ({
        llmModel: config?.llmModel || '',
        llmApiKey: config?.llmApiKey || '',
        llmBaseUrl: config?.llmBaseUrl || '',
        checkInterval: config?.checkInterval || 30,
        emailAddress: config?.emailAddress || '',
        twitterApiKey: config?.twitterApiKey || '',
        resendApiKey: config?.resendApiKey || '',
        osNotifyEnabled: config?.osNotifyEnabled ?? true,
        osNotifyMinImportance: config?.osNotifyMinImportance || Importance.ImportanceHigh
    }));

    const saveMutation = useMutation({
        mutationFn: async (data: FormState) => {
            await settingsApi.updateConfig({
                llmModel: data.llmModel,
                llmApiKey: data.llmApiKey,
                llmBaseUrl: data.llmBaseUrl,
                checkInterval: data.checkInterval,
                emailAddress: data.emailAddress,
                osNotifyEnabled: data.osNotifyEnabled,
                osNotifyMinImportance: data.osNotifyMinImportance,
                twitterApiKey: data.twitterApiKey,
                resendApiKey: data.resendApiKey
            });
        },
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: ['config'] });
            showToast('设置已保存', 'success');
        },
        onError: () => {
            showToast('保存失败', 'error');
        }
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        saveMutation.mutate(form);
    };

    const set =
        <K extends keyof FormState>(key: K) =>
        (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
            const value = key === 'checkInterval' ? Number(e.target.value) || 0 : e.target.value;

            setForm(prev => ({ ...prev, [key]: value }));
        };

    return (
        <form onSubmit={handleSubmit} className="mx-auto h-full max-w-2xl scrollbar-none space-y-8 overflow-y-scroll">
            <Section icon={Brain} title="LLM 配置">
                <Field label="模型">
                    <input value={form.llmModel} onChange={set('llmModel')} placeholder="Pro/deepseek-ai/DeepSeek-V3.2" className={CLS_INPUT} />
                </Field>
                <Field label="API Key">
                    <input type="password" value={form.llmApiKey} onChange={set('llmApiKey')} placeholder="sk-..." className={CLS_INPUT} />
                </Field>
                <Field label="Base URL" hint="自定义接入地址">
                    <input value={form.llmBaseUrl} onChange={set('llmBaseUrl')} placeholder="https://api.openai.com/v1" className={CLS_INPUT} />
                </Field>
            </Section>

            <Section icon={Clock} title="监控配置">
                <Field label="检查间隔（分钟）">
                    <input type="number" min={1} value={form.checkInterval} onChange={set('checkInterval')} className={CLS_INPUT} />
                </Field>
            </Section>

            <Section icon={Bell} title="系统通知">
                <Field label="启用操作系统通知" hint="发现新热点时弹出系统通知，由后台按策略发送">
                    <label className="flex cursor-pointer items-center gap-3">
                        <input
                            type="checkbox"
                            checked={form.osNotifyEnabled}
                            onChange={e => setForm(prev => ({ ...prev, osNotifyEnabled: e.target.checked }))}
                            className="border-border text-primary focus:ring-primary/20 h-4 w-4 rounded"
                        />
                        <span className="text-muted-foreground text-sm">{form.osNotifyEnabled ? '已开启' : '已关闭'}</span>
                    </label>
                </Field>
                <Field label="最低重要度" hint="仅当热点重要度达到此级别时发送系统通知">
                    <select
                        value={form.osNotifyMinImportance}
                        onChange={set('osNotifyMinImportance')}
                        disabled={!form.osNotifyEnabled}
                        className={`${CLS_INPUT} disabled:opacity-50`}>
                        {IMPORTANCE_OPTIONS.map(opt => (
                            <option key={opt.value} value={opt.value}>
                                {opt.label}
                            </option>
                        ))}
                    </select>
                </Field>
            </Section>

            <Section icon={Mail} title="邮件通知">
                <Field label="通知邮箱" hint="高/紧急热点会发送邮件（需配置 Resend API Key）">
                    <input
                        type="email"
                        value={form.emailAddress}
                        onChange={set('emailAddress')}
                        placeholder="you@example.com"
                        className={CLS_INPUT}
                    />
                </Field>
            </Section>

            <Section icon={Key} title="集成密钥">
                <Field label="Twitter API Key">
                    <input
                        type="password"
                        value={form.twitterApiKey}
                        onChange={set('twitterApiKey')}
                        placeholder="输入 Twitter API Key"
                        className={CLS_INPUT}
                    />
                </Field>
                <Field label="Resend API Key">
                    <input type="password" value={form.resendApiKey} onChange={set('resendApiKey')} placeholder="re_..." className={CLS_INPUT} />
                </Field>
            </Section>

            <div className="border-border border-t pt-6">
                <motion.button
                    type="submit"
                    disabled={saveMutation.isPending}
                    whileHover={{ scale: 1.01 }}
                    whileTap={{ scale: 0.99 }}
                    className="bg-primary text-primary-foreground hover:bg-primary/90 flex items-center gap-2 rounded-lg px-5 py-2.5 text-sm font-medium transition-colors disabled:opacity-50">
                    {saveMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                    保存设置
                </motion.button>
            </div>
        </form>
    );
}

function Section({ icon: Icon, title, children }: { icon: React.ElementType; title: string; children: React.ReactNode }) {
    return (
        <section className="border-border bg-muted/30 space-y-4 rounded-xl border p-5">
            <h2 className="text-foreground flex items-center gap-2 text-sm font-semibold">
                <Icon className="text-primary h-4 w-4" />
                {title}
            </h2>
            <div className="space-y-4">{children}</div>
        </section>
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

export const Route = createLazyFileRoute('/settings')({
    component: SettingsPage
});
