import { Link } from '@tanstack/react-router';
import type { Status } from '@wails/internal/module/hotspot';
import { Activity, Target, Search, Settings, Flame } from 'lucide-react';

const navItems = [
    { to: '/', label: '热点雷达', icon: Activity },
    { to: '/keywords', label: '监控词', icon: Target },
    { to: '/search', label: '搜索', icon: Search },
    { to: '/settings', label: '设置', icon: Settings }
] as const;

interface SidebarProps {
    status: Status | null;
}

export function Sidebar({ status }: SidebarProps) {
    return (
        <aside className="bg-sidebar text-sidebar-foreground border-sidebar-border flex h-screen w-56 shrink-0 flex-col border-r">
            <div className="border-sidebar-border flex items-center gap-3 border-b px-5 py-3">
                <div className="bg-sidebar-primary flex h-9 w-9 items-center justify-center rounded-lg">
                    <Flame className="text-sidebar-primary-foreground h-4.5 w-4.5" />
                </div>
                <div>
                    <h1 className="text-sm font-semibold tracking-tight">HotMonitor</h1>
                    <p className="text-sidebar-foreground/50 text-[11px]">AI 热点雷达</p>
                </div>
            </div>

            <nav className="flex-1 px-3 py-4">
                <ul className="space-y-1">
                    {navItems.map(({ to, label, icon: Icon }) => (
                        <li key={to}>
                            <Link
                                to={to}
                                activeOptions={{ exact: to === '/' }}
                                activeProps={{
                                    className: 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
                                }}
                                inactiveProps={{
                                    className: 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-accent-foreground'
                                }}
                                className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors">
                                <Icon className="h-4 w-4" />
                                {label}
                            </Link>
                        </li>
                    ))}
                </ul>
            </nav>

            {status && (
                <div className="border-sidebar-border border-t px-4 py-3">
                    <div className="grid grid-cols-3 gap-2 text-center">
                        <div>
                            <p className="text-sidebar-foreground text-sm font-semibold">{status.total}</p>
                            <p className="text-sidebar-foreground/50 text-[10px]">总热点</p>
                        </div>
                        <div>
                            <p className="text-chart-1 text-sm font-semibold">{status.today}</p>
                            <p className="text-sidebar-foreground/50 text-[10px]">今日</p>
                        </div>
                        <div>
                            <p className="text-destructive text-sm font-semibold">{status.urgent}</p>
                            <p className="text-sidebar-foreground/50 text-[10px]">紧急</p>
                        </div>
                    </div>
                </div>
            )}
        </aside>
    );
}
