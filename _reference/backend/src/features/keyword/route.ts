import { Router } from 'express';
import { KeywordService } from '@/features/keyword/service';
import RequestHandler from '@/config/requestHandler';
import responseBody from '@/config/response';

const router = Router();

// 获取所有关键词
router.get(
    '/',
    RequestHandler(async (req, res) => {
        const keywords = await KeywordService.getAll();

        res.json(responseBody(true, '获取关键词成功', { data: keywords }));
    })
);

// 获取单个关键词
router.get(
    '/:id',
    RequestHandler(async (req, res) => {
        const keyword = await KeywordService.getById(req.params.id as string);

        res.json(responseBody(true, '获取关键词成功', { data: keyword }));
    })
);

// 创建关键词
router.post(
    '/',
    RequestHandler(async (req, res) => {
        const { text, category } = req.body;

        const keyword = await KeywordService.create({ text, category });

        res.status(201).json(responseBody(true, '创建关键词成功', { data: keyword }));
    })
);

// 更新关键词
router.put(
    '/:id',
    RequestHandler(async (req, res) => {
        const keyword = await KeywordService.update(req.params.id as string, req.body);

        res.json(responseBody(true, '更新关键词成功', { data: keyword }));
    })
);

// 删除关键词
router.delete(
    '/:id',
    RequestHandler(async (req, res) => {
        await KeywordService.delete(req.params.id as string);

        res.status(204).send();
    })
);

// 切换关键词状态
router.patch(
    '/:id/toggle',
    RequestHandler(async (req, res) => {
        const updated = await KeywordService.toggle(req.params.id as string);

        res.json(responseBody(true, '切换关键词成功', { data: updated }));
    })
);

export default router;
