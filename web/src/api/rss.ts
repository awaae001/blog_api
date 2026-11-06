import { api } from '@/utils/request'
import type { PaginatedResponse } from './response'

export type RssFeedStatus = 'active' | 'paused' | 'dead'

/** 对应后端 model.RssFeed。 */
export interface RssFeed {
  id: number
  friend_link_id: number | null
  name: string
  feed_url: string
  site_url: string
  feed_type: string
  status: RssFeedStatus
  error_count: number
  last_error: string
  /** 调度器选择的轮询间隔，单位为秒。 */
  fetch_interval: number
  next_fetch_at: number
  last_fetched_at: number
  last_success_at: number
  created_at: number
  updated_at: number
}

/** 网页通过 <link rel="alternate"> 声明的订阅源。 */
export interface FeedCandidate {
  url: string
  title: string
  type: string
}

export interface RssPost {
  id: number
  rss_id: number
  title: string
  link: string
  description: string
  author: string
  time: number
}

export interface RssFeedListParams {
  page?: number
  page_size?: number
  status?: RssFeedStatus | ''
  friend_link_id?: number
}

export interface CreateRssFeedPayload {
  feed_url: string
  /** 留空时使用订阅源标题。 */
  name?: string
  /** 0 表示无归属。 */
  friend_link_id: number
}

export interface UpdateRssFeedPayload {
  name?: string
  feed_url?: string
  /** 0 表示清除归属。 */
  friend_link_id?: number
  status?: 'active' | 'paused'
}

export interface RssFetchResult {
  not_modified: boolean
  checked_items: number
  inserted_items: number
}

export interface CreateRssFeedResult {
  feed: RssFeed
  result: RssFetchResult
}

export const getRssFeeds = (params: RssFeedListParams = {}) =>
  api.get<PaginatedResponse<RssFeed>>('/action/rss', {
    page: params.page ?? 1,
    page_size: params.page_size ?? 20,
    status: params.status || undefined,
    friend_link_id: params.friend_link_id || undefined
  })

// 下面三个会让后端现场去抓 feed，不设超时。
/** 列出网页声明的 feed，不会订阅。 */
export const discoverRssFeeds = (url: string) =>
  api.post<FeedCandidate[]>('/action/rss/discover', { url }, { timeout: 0 })
/** 后端先抓一次校验，顺手把文章存下来。 */
export const createRssFeed = (data: CreateRssFeedPayload) =>
  api.post<CreateRssFeedResult>('/action/rss', data, { timeout: 0 })
/** 跳过调度和缓存立即刷新。 */
export const fetchRssFeed = (id: number) => api.post<RssFetchResult>(`/action/rss/${id}/fetch`, undefined, { timeout: 0 })

export const updateRssFeed = (id: number, data: UpdateRssFeedPayload) => api.put(`/action/rss/${id}`, data)
export const deleteRssFeed = (id: number) => api.del(`/action/rss/${id}`)

export const getPostsByFeed = (rssId: number, page = 1, pageSize = 20) =>
  api.get<PaginatedResponse<RssPost>>('/public/rss', { rss_id: rssId, page, page_size: pageSize })
export const getAllPosts = (page = 1, pageSize = 20) =>
  api.get<PaginatedResponse<RssPost>>('/public/rss', { page, page_size: pageSize })
export const deleteRssPost = (id: number) => api.del<{ rows_affected: number }>(`/action/rss/posts/${id}`)
