import request from '@/utils/request';
import type { ApiResponse } from './response'

export interface SystemConfig {
  system_conf: {
    safe_conf: SafeConfig
    data_conf: DataConfig
    crawler_conf: CrawlerConfig
    moments_integrated_conf: MomentsIntegratedConfig
    oss_conf: OSSConfig
    verify_conf: VerifyConfig
    email_conf: EmailConfig
  }
}

type MergePatch<T> = {
  [K in keyof T]?: (T[K] extends unknown[] ? T[K] : T[K] extends object ? MergePatch<T[K]> : T[K]) | null
}

export interface SystemConfigPatch {
  system_conf: MergePatch<SystemConfig['system_conf']>
}

export interface SafeConfig {
  cors_allow_hostlist: string[]
  exclude_paths: string[]
  allow_extension: string[]
  enabled_public_apis: PublicAPIKey[]
}

export interface DataConfig {
  database: DatabaseConfig
  image: ImageConfig
  resource: ResourceConfig
}

export interface DatabaseConfig {
  path: string
}

export interface ImageConfig {
  path: string
  conv_to: string
}

export interface ResourceConfig {
  path: string
}

export type PublicAPIKey = 'moments' | 'image' | 'friend' | 'rss' | 'email'

export interface CrawlerConfig {
  concurrency: number
  rss_timeout_seconds: number
}

export interface MomentsIntegratedConfig {
  enable: boolean
  api_single_return_entries: number
  integrated: IntegratedTargets
}

export interface OSSConfig {
  provider: string
  enable: boolean
  accessKeyId: string
  accessKeySecret: string
  bucket: string
  endpoint: string
  region: string
  secure: boolean
  timeout: number
  prefix: string
  customDomain: string
}

export interface VerifyConfig {
  turnstile: TurnstileConfig
  fingerprint: FingerprintConfig
}

export interface TurnstileConfig {
  enable: boolean
  secret: string
  site_key: string
}

export interface FingerprintConfig {
  secret: string
}

export interface EmailConfig {
  enable: boolean
  host: string
  user_name: string
  password: string
  port: number
  sender: string
}

export interface IntegratedTargets {
  telegram: TelegramConfig
  discord: DiscordConfig
}

export interface TelegramConfig {
  enable: boolean
  sync_delete: boolean
  bot_token: string
  channel_id: string
  media_path: string
  filter_userid: string[]
}

export interface DiscordConfig {
  enable: boolean
  sync_delete: boolean
  bot_token: string
  guild_id: string
  channel_id: string
  filter_userid: string[]
}

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
