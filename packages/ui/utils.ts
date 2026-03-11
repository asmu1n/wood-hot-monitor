import { twJoin, twMerge, type ClassNameValue } from 'tailwind-merge';

function cn(...inputs: ClassNameValue[]) {
    return twMerge(twJoin(inputs));
}

export { cn };
