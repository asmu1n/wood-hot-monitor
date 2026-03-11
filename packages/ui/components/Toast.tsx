import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Check, X } from 'lucide-react';
import { cn } from 'utils';

interface ToastProps {
    toast: { message: string; type: 'success' | 'error' } | null;
}

const Toast: React.FC<ToastProps> = ({ toast }) => {
    return (
        <AnimatePresence>
            {toast && (
                <motion.div
                    initial={{ opacity: 0, y: -20, x: '-50%' }}
                    animate={{ opacity: 1, y: 0, x: '-50%' }}
                    exit={{ opacity: 0, y: -20 }}
                    className={cn(
                        'fixed top-6 left-1/2 z-50 flex items-center gap-3 rounded-xl px-5 py-3 shadow-2xl backdrop-blur-xl',
                        toast.type === 'success'
                            ? 'border-chart-2/30 bg-chart-2/10 text-chart-2 border'
                            : 'border-destructive/30 bg-destructive/10 text-destructive border'
                    )}>
                    {toast.type === 'success' ? <Check className="h-4 w-4" /> : <X className="h-4 w-4" />}
                    <span className="text-sm font-medium">{toast.message}</span>
                </motion.div>
            )}
        </AnimatePresence>
    );
};

export default Toast;
