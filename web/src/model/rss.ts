export type RssFeedStatus = 'active' | 'paused' | 'dead'

/**
 * One subscription, corresponding to the backend's model.RssFeed.
 */
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
  /** Poll interval in seconds chosen by the scheduler. */
  fetch_interval: number
  next_fetch_at: number
  last_fetched_at: number
  last_success_at: number
  created_at: number
  updated_at: number
}

/**
 * A feed advertised by a page through <link rel="alternate">.
 */
export interface FeedCandidate {
  url: string
  title: string
  type: string
}

/**
 * Represents a post from an RSS feed, corresponding to the backend's model.RssPost.
 */
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
  /** Empty falls back to the feed title. */
  name?: string
  /** 0 means no owner. */
  friend_link_id: number
}

export interface UpdateRssFeedPayload {
  name?: string
  feed_url?: string
  /** 0 clears the owner. */
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
 * Generic interface for paginated API responses.
 */
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}
