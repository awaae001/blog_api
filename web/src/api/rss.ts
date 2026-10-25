import request from '@/utils/request'
import type { ApiResponse, PaginatedResponse } from './response'

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

/**
 * GET /api/action/rss
 */
export const getRssFeeds = (
  params: RssFeedListParams = {}
): Promise<ApiResponse<PaginatedResponse<RssFeed>>> => {
  return request({
    url: '/action/rss',
    method: 'get',
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 20,
      status: params.status || undefined,
      friend_link_id: params.friend_link_id || undefined
    }
  })
}

/**
 * POST /api/action/rss/discover
 * Lists the feeds a web page advertises. Nothing is subscribed.
 */
export const discoverRssFeeds = (url: string): Promise<ApiResponse<FeedCandidate[]>> => {
  return request({
    url: '/action/rss/discover',
    method: 'post',
    data: { url },
    timeout: 0
  })
}

/**
 * POST /api/action/rss
 * The backend fetches the feed once to validate it and stores its articles.
 */
export const createRssFeed = (
  data: CreateRssFeedPayload
): Promise<ApiResponse<CreateRssFeedResult>> => {
  return request({
    url: '/action/rss',
    method: 'post',
    data,
    timeout: 0
  })
}

/**
 * PUT /api/action/rss/:id
 */
export const updateRssFeed = (id: number, data: UpdateRssFeedPayload): Promise<ApiResponse> => {
  return request({
    url: `/action/rss/${id}`,
    method: 'put',
    data
  })
}

/**
 * DELETE /api/action/rss/:id
 */
export const deleteRssFeed = (id: number): Promise<ApiResponse> => {
  return request({
    url: `/action/rss/${id}`,
    method: 'delete'
  })
}

/**
 * POST /api/action/rss/:id/fetch
 * Refreshes one feed now, bypassing its schedule and cache.
 */
export const fetchRssFeed = (id: number): Promise<ApiResponse<RssFetchResult>> => {
  return request({
    url: `/action/rss/${id}/fetch`,
    method: 'post',
    timeout: 0
  })
}

/**
 * GET /api/public/rss?rss_id=
 */
export const getPostsByFeed = (
  rssId: number,
  page = 1,
  pageSize = 20
): Promise<ApiResponse<PaginatedResponse<RssPost>>> => {
  return request({
    url: '/public/rss',
    method: 'get',
    params: {
      rss_id: rssId,
      page,
      page_size: pageSize
    }
  })
}

/**
 * GET /api/public/rss
 */
export const getAllPosts = (
  page = 1,
  pageSize = 20
): Promise<ApiResponse<PaginatedResponse<RssPost>>> => {
  return request({
    url: '/public/rss',
    method: 'get',
    params: {
      page,
      page_size: pageSize
    }
  })
}

/**
 * DELETE /api/action/rss/posts/:id
 */
export const deleteRssPost = (id: number): Promise<ApiResponse<{ rows_affected: number }>> => {
  return request({
    url: `/action/rss/posts/${id}`,
    method: 'delete'
  })
}
