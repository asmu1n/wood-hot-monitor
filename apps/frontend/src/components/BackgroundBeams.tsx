import { cn } from '@repo/ui';

export const BackgroundBeams = ({ className }: { className?: string }) => {
    return (
        <div
            className={cn(
                'pointer-events-none absolute inset-0 overflow-hidden mask-[radial-gradient(ellipse_at_center,transparent_20%,black)]',
                className
            )}>
            {/* Primary Glow Layer */}
            <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-20%,var(--primary),transparent)] opacity-15" />

            {/* Accent Ambient Glow */}
            <div className="absolute inset-0 bg-[radial-gradient(ellipse_60%_40%_at_20%_10%,var(--accent),transparent)] opacity-10" />

            <svg className="absolute inset-0 h-full w-full" xmlns="http://www.w3.org/2000/svg">
                <defs>
                    <pattern id="grid-pattern" width="40" height="40" patternUnits="userSpaceOnUse">
                        {/* Using muted-foreground for the grid dots with low opacity */}
                        <circle cx="1" cy="1" r="0.7" fill="currentColor" className="text-muted-foreground/20" />
                    </pattern>
                </defs>
                <rect width="100%" height="100%" fill="url(#grid-pattern)" />
            </svg>
        </div>
    );
};
