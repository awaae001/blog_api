BEGIN IMMEDIATE;

-- Keep the oldest row for links duplicated by a legacy database.
DELETE FROM friend_rss_post
WHERE id NOT IN (
  SELECT MIN(id)
  FROM friend_rss_post
  GROUP BY link
);

-- Legacy databases created this as a non-unique index. Rebuild it so
-- INSERT ... ON CONFLICT(link) matches a unique constraint.
DROP INDEX IF EXISTS idx_friend_rss_post_link;
CREATE UNIQUE INDEX idx_friend_rss_post_link ON friend_rss_post(link);

COMMIT;
