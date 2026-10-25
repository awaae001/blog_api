import request from '@/utils/request';
import type { SystemConfig, SystemConfigPatch } from '@/model/config';
import type { ApiResponse } from '@/model/response';

const CONFIG_FILE_PATH = '/config/system_config.json';

export interface UpdateConfigResult {
  message: string;
  reloaded: boolean;
  restart_required_keys: string[];
}

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

/** 只提交变化的字段，按 JSON Merge Patch 合并。 */
export const updateSystemConfig = (patch: SystemConfigPatch) => {
  return request.patch<any, ApiResponse<UpdateConfigResult>>('/action/config', patch, {
    headers: { 'Content-Type': 'application/merge-patch+json' },
  });
};

/**
 * 请求重启后端进程。
 * 实际拉起由外部守护进程负责。
 */
export const restartSystem = () => {
  return request.post<any, ApiResponse<{ detail: string }>>('/action/system/restart');
};
