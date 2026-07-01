import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { resolve } from 'path';
import { tanstackRouter } from '@tanstack/router-plugin/vite';
import wails from '@wailsio/runtime/plugins/vite';

export default defineConfig(({ command, mode }) => {
    const env = loadEnv(mode, process.cwd(), '');

    const common = {
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
        resolve: { alias: { '@': resolve(__dirname, './src') } }
    };

    if (command === 'serve') {
        return {
            ...common,
            server: {
                host: '127.0.0.1',
                port: parseInt(env.VITE_PORT || '9245', 10),
                strictPort: true,
                proxy: {
                    '/api': {
                        target: 'http://localhost:3001',
                        changeOrigin: true
                    },
                    '/socket.io': {
                        target: 'http://localhost:3001',
                        ws: true
                    }
                }
            }
        };
    } else {
        return {
            ...common,
            server: {
                host: '127.0.0.1',
                port: parseInt(env.VITE_PORT || '9245', 10),
                strictPort: true
            },
            build: {
                outDir: 'dist'
            }
        };
    }
});
