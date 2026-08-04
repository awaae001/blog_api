-- 修复 friend_rss (friend_link_id, rss_url) 重复问题：
-- 1) 将重复 feed 名下的文章统一改挂到每组保留行（MIN(id)）
-- 2) 删除重复 feed 行
-- 3) 建立唯一索引，杜绝 check-then-insert 竞态再次产生重复

-- 步骤 1: 文章改挂。非重复组 MIN(id) 即自身，属于无操作，整体幂等。
-- 安全性：friend_rss_post.link 有全局唯一索引，同一篇文章不可能同时挂在两条 feed 下，改挂不会冲突。
UPDATE friend_rss_post
SET rss_id = (
    SELECT MIN(f2.id)
    FROM friend_rss f2
    WHERE f2.friend_link_id = (SELECT friend_link_id FROM friend_rss WHERE id = friend_rss_post.rss_id)
      AND f2.rss_url       = (SELECT rss_url        FROM friend_rss WHERE id = friend_rss_post.rss_id)
);

-- 步骤 2: 每组 (friend_link_id, rss_url) 只保留 id 最小的一行
DELETE FROM friend_rss
WHERE id NOT IN (
    SELECT MIN(id) FROM friend_rss GROUP BY friend_link_id, rss_url
);

-- 步骤 3: 唯一索引兜底，并删除被其覆盖的旧非唯一复合索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_friend_rss_unique_link_url
    ON friend_rss(friend_link_id, rss_url);
DROP INDEX IF EXISTS idx_friend_rss_friend_link_id_rss_url;
