import { useMutation } from '@tanstack/react-query';
import { useToast } from '@/hooks/useToast';
import { useEffect, useState } from 'react';
import { onCheckComplete, onCheckStart } from '../utils';
import { Run } from '@wails/module/checker/service';

export function useCheckerStatus() {
    const { showToast } = useToast();
    const [isChecking, setIsChecking] = useState(false);

    const { mutate: manualCheck } = useMutation({
        mutationFn: () => Run(),
        onSuccess: () => {
            showToast('热点检查已完成', 'success');
        },
        onError: () => {
            showToast('热点检查失败', 'error');
        }
    });

    const handleManualCheck = () => {
        manualCheck();
    };

    useEffect(() => {
        const startCancel = onCheckStart(() => {
            setIsChecking(true);
        });

        const completeCancel = onCheckComplete(() => {
            setIsChecking(false);
        });

        return () => {
            startCancel();
            completeCancel();
        };
    }, []);

    return {
        isChecking,
        handleManualCheck
    };
}
