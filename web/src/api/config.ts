import request from '@/utils/request';
import type { SystemConfig } from '@/model/config';
import type { ApiResponse } from '@/model/response';

const CONFIG_FILE_PATH = '/config/system_config.json';

/**
 * 获取系统配置
 * 注意：资源 API 直接返回文件内容，不包装在 ApiResponse 中
 * @returns SystemConfig
 */
export const getSystemConfig = () => {
  return request.get<any, SystemConfig>(`/action/resource/${CONFIG_FILE_PATH}`, {
    params: {
      // 添加时间戳以防止缓存
      _t: new Date().getTime(),
    },
  });
};

/**
 * 更新系统配置项
 * @param key 配置键 (e.g., "system_conf.safe_conf.cors_allow_hostlist")
 * @param value 配置值
 * @returns
 */
export const updateSystemConfig = (key: string, value: any) => {
  return request.put<any, ApiResponse<{ message: string }>>('/action/config', {
    key,
    value,
  });
};