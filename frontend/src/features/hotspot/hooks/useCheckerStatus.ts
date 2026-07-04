import { useMutation } from '@tanstack/react-query';
import { Run } from '@wails/worker/checker/service';
import { useToast } from '@/hooks/useToast';

export function useCheckerStatus() {
    const { showToast } = useToast();

    const { mutate: manualCheck, isPending: isChecking } = useMutation({
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

    return {
        isChecking,
        handleManualCheck
    };
}
