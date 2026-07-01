import { createLazyFileRoute } from '@tanstack/react-router';
import { motion, AnimatePresence } from 'framer-motion';
import { Plus, Target } from 'lucide-react';

import { useApp } from '@/context/AppContext';
import KeywordCard from '@/features/keyword/components/KeywordCard';

function Keywords() {
    const { keywords, newKeyword, setNewKeyword, handleAddKeyword, handleToggleKeyword, handleDeleteKeyword } = useApp();

    return (
        <div className="space-y-6">
            {/* Add Keyword Card */}
            <form onSubmit={handleAddKeyword} className="border-border bg-muted/30 rounded-2xl border p-5 shadow-sm">
                <div className="flex gap-3">
                    <div className="relative flex-1">
                        <input
                            type="text"
                            value={newKeyword}
                            onChange={e => setNewKeyword(e.target.value)}
                            placeholder="输入要监控的关键词，如：GPT-5、AI编程、Cursor..."
                            className="border-border bg-background text-foreground placeholder-muted-foreground/50 focus:border-primary/50 focus:ring-primary/20 w-full rounded-xl border px-4 py-3 transition-all focus:ring-2 focus:outline-none"
                        />
                    </div>
                    <motion.button
                        type="submit"
                        whileHover={{ scale: 1.02 }}
                        whileTap={{ scale: 0.98 }}
                        className="bg-primary text-primary-foreground shadow-primary/25 flex items-center gap-2 rounded-xl px-6 py-3 font-medium shadow-lg">
                        <Plus className="h-4 w-4" />
                        添加
                    </motion.button>
                </div>
            </form>

            {/* Keywords Grid */}
            <div className="grid gap-3 md:grid-cols-2">
                <AnimatePresence>
                    {keywords.map((keyword, i) => (
                        <KeywordCard key={keyword.id} keyword={keyword} index={i} onToggle={handleToggleKeyword} onDelete={handleDeleteKeyword} />
                    ))}
                </AnimatePresence>
            </div>

            {keywords.length === 0 && (
                <div className="border-border bg-muted/20 rounded-2xl border border-dashed py-16 text-center">
                    <div className="bg-muted mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full">
                        <Target className="text-muted-foreground h-8 w-8" />
                    </div>
                    <p className="text-foreground font-medium">还没有监控关键词</p>
                    <p className="text-muted-foreground mt-1 text-sm">添加你想追踪的技术热点词</p>
                </div>
            )}
        </div>
    );
}

export const Route = createLazyFileRoute('/keywords')({
    component: Keywords
});
