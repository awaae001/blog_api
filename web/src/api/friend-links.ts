import { api } from "@/utils/request";
import type { PaginatedResponse } from "./response";

export interface FriendTag {
  id: number;
  name: string;
  color: string;
  sort_order: number;
  /** 只有标签列表接口会带。 */
  link_count?: number;
}

export type FriendTagPayload = Pick<FriendTag, "name" | "color" | "sort_order">;

export interface FriendLink {
  id: number;
  name: string;
  link: string;
  avatar: string;
  description: string;
  readonly reason: string;
  status: "survival" | "timeout" | "error" | "pending";
  tags: FriendTag[];
  enable_rss: boolean;
  skip_health_check: boolean;
  is_died?: boolean;
  email?: string;
  times?: number;
  readonly created_at: number;
  updated_at: number;
}

export interface FriendLinkListParams {
  page?: number;
  page_size?: number;
  is_died?: boolean;
  status?: string;
  search?: string;
  tag_id?: number;
}

export type PaginatedFriendLinks = PaginatedResponse<FriendLink>;

export interface CreateFriendLinkPayload {
  name: string;
  link: string;
  avatar?: string;
  description?: string;
  email?: string;
  enable_rss?: boolean;
}

export type UpdateFriendLinkData = Partial<
  Omit<FriendLink, "id" | "created_at" | "updated_at" | "reason" | "tags">
> & {
  website_name?: string;
  website_url?: string;
  website_icon_url?: string;
};

export interface UpdateFriendLinkPayload {
  data: UpdateFriendLinkData;
  opt?: {
    overwrite_if_blank?: boolean;
  };
}

/** 一次巡查任务的进度，进程重启后重置。 */
export interface FriendLinkInspectionProgress {
  running: boolean;
  label?: string;
  started_at?: number;
  finished_at?: number;
  total: number;
  processed: number;
  survival: number;
  failed: number;
  database_failures: number;
  rss_discovered: number;
  error?: string;
}

export interface FriendLinkRecheckResult {
  success: boolean;
}

// 前端字段名和后端 website_* 不一致，提交前映射一下。
const fieldMap: Record<string, string> = {
  name: "website_name",
  link: "website_url",
  avatar: "website_icon_url",
};

export const getFriendLinks = (params: FriendLinkListParams) => api.get<PaginatedFriendLinks>("/action/friend", params);
export const createFriendLink = (data: CreateFriendLinkPayload) => api.post("/action/friend", data);
export const updateFriendLink = (id: number, payload: UpdateFriendLinkPayload) =>
  api.put(`/action/friend/${id}`, {
    ...payload,
    data: Object.fromEntries(Object.entries(payload.data).map(([key, value]) => [fieldMap[key] ?? key, value])),
  });
export const deleteFriendLink = (id: number) => api.del(`/action/friend/${id}`);

/** 手动重新巡查单条友链；开启不巡查时只返回结果，不更新数据。 */
export const recheckFriendLink = (id: number) => api.post<FriendLinkRecheckResult>(`/action/friend/recheck/${id}`);
/** 全量巡查在后台异步跑，用 getRecheckProgress 轮询进度。 */
export const recheckAllFriendLinks = () => api.post<FriendLinkInspectionProgress>("/action/friend/recheck");
export const getRecheckProgress = () => api.get<FriendLinkInspectionProgress>("/action/friend/recheck");

/** 全部标签，包括还没挂到友链上的。 */
export const getFriendTags = () => api.get<FriendTag[]>("/action/friend/tags");
export const createFriendTag = (data: FriendTagPayload) => api.post<FriendTag>("/action/friend/tags", data);
/** 局部更新，没传的字段保持不变。 */
export const updateFriendTag = (id: number, data: Partial<FriendTagPayload>) => api.put<FriendTag>(`/action/friend/tags/${id}`, data);
export const deleteFriendTag = (id: number) => api.del(`/action/friend/tags/${id}`);
/** 整体替换友链的标签，传 [] 清空。 */
export const replaceFriendLinkTags = (id: number, tagIds: number[]) => api.put<FriendTag[]>(`/action/friend/${id}/tags`, { tag_ids: tagIds });
