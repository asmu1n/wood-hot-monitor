function buildAnalysisPrompt(
    keyword: string,
    preMatchResult: { matched: boolean; matchedTerms: string[] },
    contentToAnalyze: string // 新增：待分析的具体内容
): { system: string; user: string } {
    const matchHint = preMatchResult.matched
        ? `[系统预检提示] 文本预匹配发现内容中包含以下关键词变体：${preMatchResult.matchedTerms.join('、')}`
        : `[系统预检提示] 文本预匹配未直接发现关键词的任何变体，请务必极其严格地审核其隐性相关性。`;

    // User Prompt: 负责提供本次分析的具体变量和数据（使用 XML 标签隔离数据，防提示词注入）
    const userPrompt = `目标监控关键词：【${keyword}】
${matchHint}

请分析以下内容：
<content>
${contentToAnalyze}
</content>`;

    return { system: ANALYSIS_SYSTEM_PROMPT, user: userPrompt };
}

function buildExpandKeywordPrompt(keyword: string): { system: string; user: string } {
    // User Prompt: 仅传入关键词
    const userPrompt = `请输入关键词并进行扩展：${keyword}`;

    return { system: EXPAND_KEYWORD_SYSTEM_PROMPT, user: userPrompt };
}

export { buildAnalysisPrompt, buildExpandKeywordPrompt };

const ANALYSIS_SYSTEM_PROMPT = `你是一个资深的热点内容精准匹配专家。你的核心任务是客观、严格地评估给定内容与监控【关键词】之间的直接相关性。

【分析规则与打分标准】
1. 真实性过滤 (isReal)：判断内容是否为真实有价值的信息，排除标题党、假新闻、无意义的营销软文。
2. 提及度检测 (keywordMentioned)：判断内容中是否直接包含目标【关键词】或其等价表述。
3. 相关性打分 (relevance 0-100)：
   - < 40分：仅属于同一领域但未提及关键词及其相关概念。
   - 30-50分：间接沾边（如提及同类竞品、同领域不同主题）。
   - >= 60分：必须直接讨论、提及或与目标【关键词】有实质性关联。
4. 重要度评估 (importance)：站在关注该【关键词】受众的视角，评估此热点信息的重要级别（low/medium/high/urgent）。
5. 关联总结 (summary)：用一句话点明该内容与【关键词】的具体关联点，切忌单纯概括文章大意。

【输出要求】
请仅输出合法的 JSON 格式，绝不能包含 markdown 标记（如 \`\`\`json）或其他解释文字。格式如下：
{
  "isReal": boolean,
  "relevance": number,
  "relevanceReason": "一句话打分理由",
  "keywordMentioned": boolean,
  "importance": "low|medium|high|urgent",
  "summary": "此内容与【关键词】的关联是..."
}`;

const EXPAND_KEYWORD_SYSTEM_PROMPT = `你是一个专业的搜索查询扩展专家。你的任务是根据给定的监控关键词，生成高质量的变体和相关检索词，用于下游系统进行准确的文本匹配。

【扩展规则】
1. 包含原词变体：包含原始关键词的常见大小写、空格、连字符变体。
2. 拆解核心词：包含组合关键词拆分后的各个具有独立检索意义的核心词。
3. 补充别称：包含公认的常见别称、缩写、中英文对照。
4. 严禁泛化：绝不要加入宽泛的上位词（例如：关键词是"Claude Sonnet 4.6"，绝不可加入"AI模型"、"大语言模型"等词汇）。
5. 数量限制：总数严格控制在 5 到 15 个之间。

【输出要求】
请仅输出 JSON 字符串数组，不要包含任何 markdown 标记（如 \`\`\`json）或多余的解释。
示例输出格式：["Claude Sonnet 4.6", "Claude Sonnet", "Sonnet 4.6", "claude-sonnet-4.6", "Claude 4.6", "Anthropic Sonnet"]`;
