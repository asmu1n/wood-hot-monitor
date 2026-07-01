import { Router } from 'express';
import { HotspotService } from '@/features/hotspot/service';
import RequestHandler from '@/config/requestHandler';
import responseBody from '@/config/response';
import { BadRequestError } from '@/config/error';

const router = Router();

// 获取所有热点
router.get(
    '/',
    RequestHandler(async (req, res) => {
        const result = await HotspotService.getHotspots(req.query);

        res.json(responseBody(true, '获取热点成功', result));
    })
);

// 获取热点统计
router.get(
    '/status',
    RequestHandler(async (req, res) => {
        const status = await HotspotService.getStatus();

        res.json(responseBody(true, '获取当前信息成功', { data: status }));
    })
);

// 获取单个热点
router.get(
    '/:id',
    RequestHandler(async (req, res) => {
        const hotspot = await HotspotService.getById(req.params.id as string);

        res.json(responseBody(true, '获取热点成功', { data: hotspot }));
    })
);

// 搜索热点
router.post(
    '/search',
    RequestHandler(async (req, res) => {
        const { query, sources } = req.body;

        if (!query) {
            throw new BadRequestError('搜索词不能为空');
        }

        const results = await HotspotService.searchExternal(query, sources);

        res.json(responseBody(true, '搜索热点成功', { data: results }));
    })
);

// 删除热点
router.delete(
    '/:id',
    RequestHandler(async (req, res) => {
        await HotspotService.delete(req.params.id as string);

        res.status(204).send();
    })
);

export default router;
