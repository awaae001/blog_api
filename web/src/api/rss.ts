import request from '@/utils/request'
import type { ApiResponse } from '@/model/response'

import type {
  CreateRssFeedPayload,
  CreateRssFeedResult,
  FeedCandidate,
  PaginatedResponse,
  RssFeed,
  RssFeedListParams,
  RssFetchResult,
  RssPost,
  UpdateRssFeedPayload
} from '@/model/rss'

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
