'use client';

import { cn } from '@repo/ui';
import { useEffect, useState } from 'react';

export const Meteors = ({ number = 12, className }: { number?: number; className?: string }) => {
    const [meteorStyles, setMeteorStyles] = useState<{ left: string; animationDelay: string; animationDuration: string }[]>([]);

    useEffect(() => {
        const styles = [...new Array(number)].map(() => ({
            left: Math.floor(Math.random() * 800 - 400) + 'px',
            animationDelay: (Math.random() * (0.8 - 0.2) + 0.2).toFixed(2) + 's',
            animationDuration: Math.floor(Math.random() * (10 - 2) + 2) + 's'
        }));

        const timer = setTimeout(() => {
            setMeteorStyles(styles);
        }, 0);

        return () => clearTimeout(timer);
    }, [number]);

    return (
        <>
            {meteorStyles.map((style, idx) => (
                <span
                    key={'meteor' + idx}
                    className={cn(
                        'animate-meteor-effect bg-primary/50 absolute h-0.5 w-0.5 rotate-215 rounded-full shadow-[0_0_0_1px_var(--primary-foreground)]/10',
                        "before:from-primary/30 before:absolute before:top-1/2 before:h-px before:w-[50px] before:-translate-y-[50%] before:transform before:bg-linear-to-r before:to-transparent before:content-['']",
                        className
                    )}
                    style={{
                        top: 0,
                        ...style
                    }}
                />
            ))}
        </>
    );
};
