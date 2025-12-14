/**
 * Represents an RSS feed, corresponding to the backend's model.FriendRss.
 */
export interface RssFeed {
  id: number
  friend_link_id: number
  name: string
  rss_url: string
  status: string
  updated_at: number
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
  time: number
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