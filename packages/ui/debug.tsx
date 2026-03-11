import { scan } from 'react-scan';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import Button from 'components/Button';
import './index.css';

scan({ enabled: true, log: false });

const rootElement = document.getElementById('root')!;

if (!rootElement.innerHTML) {
    const root = createRoot(rootElement);

    root.render(
        <StrictMode>
            <Button>Button</Button>
        </StrictMode>
    );
}
