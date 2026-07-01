import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { resolve } from 'path';
import { tanstackRouter } from '@tanstack/router-plugin/vite';
import wails from '@wailsio/runtime/plugins/vite';

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, process.cwd(), '');

    return {
        plugins: [
            react({
                babel: {
                    plugins: ['babel-plugin-react-compiler']
                }
            }),
            tailwindcss(),
            tanstackRouter(),
            wails('./bindings')
        ],
        resolve: {
            alias: {
                '@': resolve(__dirname, './src'),
                '@wails': resolve(__dirname, './bindings/wood-hot-monitor/internal')
            }
        },
        server: {
            host: '127.0.0.1',
            port: parseInt(env.VITE_PORT || '9245', 10),
            strictPort: true
        },
        build: {
            outDir: 'dist'
        }
    };
});
