import { createLazyFileRoute } from '@tanstack/react-router';
import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Save, Brain, Clock, Mail, Key, Loader2 } from 'lucide-react';
import { settingsApi } from '@/features/settings/api';
import { useApp } from '@/context/AppContext';

interface FormState {
    llmProvider: string;
    llmModel: string;
    llmApiKey: string;
    llmBaseUrl: string;
    checkInterval: number;
    emailAddress: string;
    twitterApiKey: string;
    resendApiKey: string;
}

const emptyForm: FormState = {
    llmProvider: '',
    llmModel: '',
    llmApiKey: '',
    llmBaseUrl: '',
    checkInterval: 30,
    emailAddress: '',
    twitterApiKey: '',
    resendApiKey: '',
};

function SettingsPage() {
    const queryClient = useQueryClient();
    const { showToast } = useApp();
    const [form, setForm] = useState<FormState>(emptyForm);

    const { data: config, isLoading } = useQuery({
        queryKey: ['config'],
        queryFn: settingsApi.getConfig,
    });

    useEffect(() => {
        if (config) {
            setForm({
                llmProvider: config.llmProvider || '',
                llmModel: config.llmModel || '',
                llmApiKey: config.llmApiKey || '',
                llmBaseUrl: config.llmBaseUrl || '',
                checkInterval: config.checkInterval || 30,
                emailAddress: config.emailAddress || '',
                twitterApiKey: (config.settings?.twitterApiKey as string) || '',
                resendApiKey: (config.settings?.resendApiKey as string) || '',
            });
        }
    }, [config]);

    const saveMutation = useMutation({
        mutationFn: async (data: FormState) => {
            await settingsApi.updateConfig({
                llmProvider: data.llmProvider,
                llmModel: data.llmModel,
                llmApiKey: data.llmApiKey,
                llmBaseUrl: data.llmBaseUrl,
                checkInterval: data.checkInterval,
                emailAddress: data.emailAddress,
                settings: {
                    ...(config?.settings ?? {}),
                    twitterApiKey: data.twitterApiKey,
                    resendApiKey: data.resendApiKey,
                },
            });
        },
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: ['config'] });
            showToast('设置已保存', 'success');
        },
        onError: () => {
            showToast('保存失败', 'error');
        },
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        saveMutation.mutate(form);
    };

    const set = <K extends keyof FormState>(key: K) => (e: React.ChangeEvent<HTMLInputElement>) => {
        setForm(prev => ({ ...prev, [key]: key === 'checkInterval' ? Number(e.target.value) || 0 : e.target.value }));
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center py-20">
                <Loader2 className="text-muted-foreground h-6 w-6 animate-spin" />
            </div>
        );
    }

    const inputCls =
        'border-border bg-background text-foreground placeholder-muted-foreground/50 focus:border-primary/50 focus:ring-primary/20 w-full rounded-lg border px-3 py-2.5 text-sm transition-all focus:ring-2 focus:outline-none';

    return (
        <form onSubmit={handleSubmit} className="mx-auto max-w-2xl space-y-8">
            <Section icon={Brain} title="LLM 配置">
                <Field label="服务提供商" hint="如 openai、deepseek 等">
                    <input value={form.llmProvider} onChange={set('llmProvider')} placeholder="openai" className={inputCls} />
                </Field>
                <Field label="模型">
                    <input value={form.llmModel} onChange={set('llmModel')} placeholder="Pro/deepseek-ai/DeepSeek-V3.2" className={inputCls} />
                </Field>
                <Field label="API Key">
                    <input type="password" value={form.llmApiKey} onChange={set('llmApiKey')} placeholder="sk-..." className={inputCls} />
                </Field>
                <Field label="Base URL" hint="自定义接入地址">
                    <input value={form.llmBaseUrl} onChange={set('llmBaseUrl')} placeholder="https://api.openai.com/v1" className={inputCls} />
                </Field>
            </Section>

            <Section icon={Clock} title="监控配置">
                <Field label="检查间隔（分钟）">
                    <input type="number" min={1} value={form.checkInterval} onChange={set('checkInterval')} className={inputCls} />
                </Field>
            </Section>

            <Section icon={Mail} title="通知配置">
                <Field label="通知邮箱">
                    <input type="email" value={form.emailAddress} onChange={set('emailAddress')} placeholder="you@example.com" className={inputCls} />
                </Field>
            </Section>

            <Section icon={Key} title="集成密钥">
                <Field label="Twitter API Key">
                    <input type="password" value={form.twitterApiKey} onChange={set('twitterApiKey')} placeholder="输入 Twitter API Key" className={inputCls} />
                </Field>
                <Field label="Resend API Key">
                    <input type="password" value={form.resendApiKey} onChange={set('resendApiKey')} placeholder="re_..." className={inputCls} />
                </Field>
            </Section>

            <div className="border-border border-t pt-6">
                <motion.button
                    type="submit"
                    disabled={saveMutation.isPending}
                    whileHover={{ scale: 1.01 }}
                    whileTap={{ scale: 0.99 }}
                    className="bg-primary text-primary-foreground hover:bg-primary/90 flex items-center gap-2 rounded-lg px-5 py-2.5 text-sm font-medium transition-colors disabled:opacity-50">
                    {saveMutation.isPending ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                        <Save className="h-4 w-4" />
                    )}
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
    component: SettingsPage,
});
