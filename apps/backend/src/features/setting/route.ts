import { Router } from 'express';
import { SettingService } from '@/features/setting/service';
import RequestHandler from '@/config/requestHandler';
import responseBody from '@/config/response';
import { BadRequestError } from '@/config/error';

const router = Router();

// 获取所有设置
router.get(
    '/',
    RequestHandler(async (req, res) => {
        const settings = await SettingService.getSettingsMap();

        res.json(responseBody(true, '获取设置成功', { data: settings }));
    })
);

// 获取单个设置
router.get(
    '/:key',
    RequestHandler(async (req, res) => {
        const setting = await SettingService.getByKey(req.params.key as string);

        res.json(responseBody(true, '获取设置成功', { data: setting }));
    })
);

// 批量更新设置
router.put(
    '/',
    RequestHandler(async (req, res) => {
        const settings = req.body;

        if (typeof settings !== 'object' || settings === null) {
            throw new BadRequestError('Invalid settings format');
        }

        const result = await SettingService.bulkUpdate(settings);

        res.json(responseBody(true, result.message));
    })
);

// 更新或创建单个设置
router.put(
    '/:key',
    RequestHandler(async (req, res) => {
        const { value } = req.body;

        if (value === undefined) {
            throw new BadRequestError('Value is required');
        }

        const setting = await SettingService.upsert(req.params.key as string, value);

        res.json(setting);
    })
);

export default router;
