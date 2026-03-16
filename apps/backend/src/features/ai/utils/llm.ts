import { SILICONFLOW_API_KEY } from 'env.config';

const SILICONFLOW_API_URL = 'https://api.siliconflow.cn/v1/chat/completions';
const MODEL_NAME = 'Pro/deepseek-ai/DeepSeek-V3.2';

export async function fetchSiliconFlow(prompt: string, systemPrompt?: string): Promise<string> {
    const messages = [];

    if (systemPrompt) {
        messages.push({ role: 'system', content: systemPrompt });
    }

    messages.push({ role: 'user', content: prompt });

    const response = await fetch(SILICONFLOW_API_URL, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${SILICONFLOW_API_KEY}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            model: MODEL_NAME,
            messages,
            temperature: 0.2,
            top_p: 0.95,
            max_tokens: 1024,
            response_format: { type: 'json_object' }
        })
    });

    if (!response.ok) {
        const errText = await response.text();

        throw new Error(`SiliconFlow API error: ${response.status} - ${errText}`);
    }

    const data = await response.json();

    return data.choices?.[0]?.message?.content || '';
}
