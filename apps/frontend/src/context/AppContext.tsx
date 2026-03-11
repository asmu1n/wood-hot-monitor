import React, { createContext, useContext } from 'react';
import { useAppLogic, type AppLogic } from '@/hooks/useAppLogic';

const AppContext = createContext<AppLogic | null>(null);

export function AppProvider({ children }: { children: React.ReactNode }) {
    const logic = useAppLogic();

    return <AppContext.Provider value={logic}>{children}</AppContext.Provider>;
}

export function useApp() {
    const context = useContext(AppContext);

    if (!context) {
        throw new Error('useApp must be used within an AppProvider');
    }

    return context;
}
