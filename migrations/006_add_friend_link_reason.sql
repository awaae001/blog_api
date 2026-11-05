-- 最近一次友链探针的问题；正常或未探测时为空。
ALTER TABLE friend_link ADD COLUMN reason TEXT NOT NULL DEFAULT '';
