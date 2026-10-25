<template>
  <PageShell :title="viewTitle">
    <template #actions>
      <template v-if="isAllPostsView">
        <el-button :icon="Back" @click="showFeedList">返回订阅源</el-button>
      </template>
      <template v-else>
        <el-select
          v-model="statusFilter"
          placeholder="全部状态"
          clearable
          style="width: 120px"
          @change="handleStatusFilterChange"
        >
          <el-option
            v-for="(meta, key) in statusMeta"
            :key="key"
            :label="meta.label"
            :value="key"
          />
        </el-select>
        <el-button type="primary" link @click="showAllPosts">查看所有文章</el-button>
      </template>
      <el-button type="primary" :icon="Plus" @click="handleCreate">订阅</el-button>
    </template>

    <!-- All posts view -->
    <template v-if="isAllPostsView">
      <el-scrollbar class="content-scroll">
        <RssPostTable
          :posts="posts"
          :loading="postsLoading"
          show-author
          @delete="handleDeletePost"
        />
      </el-scrollbar>
      <TablePagination
        :total="totalPosts"
        :page-size="postPageSize"
        :current-page="currentPostPage"
        @size-change="handlePostSizeChange"
        @current-change="handlePostPageChange"
        class="pagination-container"
      />
    </template>

    <!-- Accordion feed list -->
    <template v-else>
      <el-scrollbar class="content-scroll" v-loading="feedsLoading">
        <el-empty v-if="!feedsLoading && feeds.length === 0" description="暂无订阅源" />
        <el-collapse v-else v-model="activeFeedId" accordion @change="handleFeedChange">
          <el-collapse-item v-for="feed in feeds" :key="feed.id" :name="feed.id">
            <template #title>
              <div class="feed-title">
                <el-tooltip :content="feed.feed_url" placement="top">
                  <a :href="feed.feed_url" target="_blank" class="feed-link" @click.stop>
                    <el-icon><Link /></el-icon>
                  </a>
                </el-tooltip>
                <a
                  v-if="feed.site_url"
                  :href="feed.site_url"
                  target="_blank"
                  class="feed-name feed-name-link"
                  @click.stop
                >
                  {{ feed.name }}
                </a>
                <span v-else class="feed-name">{{ feed.name }}</span>
                <el-tag v-if="feed.feed_type" size="small" effect="plain">{{ feed.feed_type }}</el-tag>
                <el-tag :type="statusMeta[feed.status]?.tag ?? 'info'" size="small">
                  {{ statusMeta[feed.status]?.label ?? feed.status }}
                </el-tag>
                <el-tooltip v-if="feed.error_count > 0" :content="feed.last_error || '未知错误'" placement="top">
                  <el-tag type="warning" size="small" effect="plain">失败 {{ feed.error_count }} 次</el-tag>
                </el-tooltip>
                <span class="feed-meta">{{ friendLinkName(feed.friend_link_id) }}</span>
                <span class="feed-meta">上次成功 {{ formatDate(feed.last_success_at, 'MM-DD HH:mm') }}</span>
                <span v-if="feed.status !== 'paused'" class="feed-meta">
                  下次 {{ formatNextFetch(feed.next_fetch_at) }}
                </span>
                <span v-if="feed.fetch_interval" class="feed-meta">
                  每 {{ formatInterval(feed.fetch_interval) }}
                </span>
                <div class="feed-actions" @click.stop>
                  <el-button
                    type="success"
                    link
                    :icon="Refresh"
                    :loading="fetchingId === feed.id"
                    :disabled="fetchingId !== null && fetchingId !== feed.id"
                    @click="handleImmediateFetch(feed)"
                  >
                    立刻获取
                  </el-button>
                  <el-button
                    v-if="feed.status === 'paused'"
                    type="primary"
                    link
                    :icon="VideoPlay"
                    @click="handleTogglePause(feed)"
                  >
                    恢复
                  </el-button>
                  <el-button v-else type="warning" link :icon="VideoPause" @click="handleTogglePause(feed)">
                    暂停
                  </el-button>
                  <el-button type="primary" link :icon="Edit" @click="handleEdit(feed)">编辑</el-button>
                  <el-button type="danger" link :icon="Delete" @click="handleDelete(feed)">删除</el-button>
                </div>
              </div>
            </template>

            <RssPostTable
              :posts="posts"
              :loading="postsLoading"
              :max-height="420"
              @delete="handleDeletePost"
            />
            <TablePagination
              :total="totalPosts"
              :page-size="postPageSize"
              :current-page="currentPostPage"
              @size-change="handlePostSizeChange"
              @current-change="handlePostPageChange"
              class="pagination-container"
            />
          </el-collapse-item>
        </el-collapse>
      </el-scrollbar>
      <TablePagination
        layout="total, sizes, prev, pager, next"
        :total="totalFeeds"
        :page-sizes="[10, 20, 50]"
        :page-size="feedPageSize"
        :current-page="currentFeedPage"
        @size-change="handleFeedSizeChange"
        @current-change="handleFeedPageChange"
        class="pagination-container feed-pagination"
      />
    </template>

    <!-- Edit Dialog -->
    <el-dialog v-model="editDialogVisible" title="编辑订阅源" width="520px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="订阅地址">
          <el-input v-model="editForm.feed_url" />
        </el-form-item>
        <el-form-item label="归属友链">
          <el-select
            v-model="editForm.friend_link_id"
            filterable
            :loading="friendLinksLoading"
            placeholder="请选择归属友链"
            style="width: 100%"
          >
            <el-option label="无归属" :value="0" />
            <el-option v-for="link in friendLinks" :key="link.id" :label="link.name" :value="link.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="editForm.status">
            <el-radio-button value="active">正常</el-radio-button>
            <el-radio-button value="paused">暂停</el-radio-button>
            <el-radio-button value="dead" disabled>已失效</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-alert
          v-if="editingFeed?.status === 'dead' && editForm.status === 'dead'"
          type="info"
          :closable="false"
          show-icon
          title="已失效的源每 48 小时自动探测一次，切到「正常」会马上重新抓取。"
        />
        <el-alert
          v-if="editingFeed?.last_error"
          type="warning"
          :closable="false"
          show-icon
          :title="`最近错误：${editingFeed.last_error}`"
          class="dialog-alert"
        />
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- Subscribe Dialog -->
    <el-dialog v-model="createDialogVisible" title="订阅 RSS" width="600px">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="归属友链">
          <el-select
            v-model="createForm.friend_link_id"
            filterable
            :loading="friendLinksLoading"
            placeholder="请选择归属友链"
            style="width: 100%"
            @change="handleCreateFriendLinkChange"
          >
            <el-option label="无归属" :value="0" />
            <el-option v-for="link in friendLinks" :key="link.id" :label="link.name" :value="link.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="网站地址">
          <div class="discover-row">
            <el-input
              v-model="createForm.site_url"
              placeholder="https://example.com，从页面里找 feed"
              @keyup.enter="handleDiscover"
            />
            <el-button :icon="Search" :loading="discovering" @click="handleDiscover">发现</el-button>
          </div>
        </el-form-item>
        <el-form-item v-if="discovered" label="候选">
          <el-empty v-if="candidates.length === 0" :image-size="60" description="页面没有声明 feed，可以直接填订阅地址" />
          <el-radio-group v-else v-model="createForm.feed_url" class="candidate-list">
            <el-radio v-for="candidate in candidates" :key="candidate.url" :value="candidate.url" class="candidate">
              <div class="candidate-body">
                <div class="candidate-title">
                  <el-tag size="small" effect="plain">{{ candidate.type }}</el-tag>
                  <span>{{ candidate.title || '未命名' }}</span>
                </div>
                <div class="candidate-url">{{ candidate.url }}</div>
              </div>
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="订阅地址" required>
          <el-input v-model="createForm.feed_url" placeholder="feed 地址，也可以直接粘贴" />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="createForm.name" placeholder="留空则使用 feed 标题" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" :disabled="!createForm.feed_url" @click="handleCreateSave">
          订阅
        </el-button>
      </template>
    </el-dialog>
  </PageShell>
</template>

<script setup lang="ts">
import PageShell from '@/components/PageShell.vue'
import TablePagination from '@/components/TablePagination.vue'
import RssPostTable from '@/components/RssPostTable.vue'
import { ref, onMounted, reactive, computed } from 'vue'
import { usePagination } from '@/utils/pagination'
import { formatDate } from '@/utils/date'
import { ElMessage, ElMessageBox, type CollapseModelValue } from 'element-plus'
import { Edit, Delete, Link, Refresh, Plus, Back, Search, VideoPause, VideoPlay, Document } from '@element-plus/icons-vue'
import {
  getRssFeeds,
  getPostsByFeed,
  getAllPosts,
  updateRssFeed,
  deleteRssFeed,
  deleteRssPost,
  createRssFeed,
  discoverRssFeeds,
  fetchRssFeed,
  type FeedCandidate,
  type RssFeed,
  type RssFeedStatus,
  type RssPost,
  type UpdateRssFeedPayload
} from '@/api/rss'
import { getFriendLinks, type FriendLink } from '@/api/friend-links'

definePage({ meta: { title: 'RSS', icon: Document, order: 4 } })

type TagType = 'success' | 'info' | 'warning' | 'danger'

const statusMeta: Record<RssFeedStatus, { label: string; tag: TagType }> = {
  active: { label: '正常', tag: 'success' },
  paused: { label: '暂停', tag: 'info' },
  dead: { label: '已失效', tag: 'danger' }
}

const feeds = ref<RssFeed[]>([])
const posts = ref<RssPost[]>([])
const friendLinks = ref<FriendLink[]>([])
const selectedFeed = ref<RssFeed | null>(null)
const isAllPostsView = ref(false)
const statusFilter = ref<RssFeedStatus | ''>('')
// Accordion keeps at most one panel open; '' means all collapsed.
const activeFeedId = ref<number | ''>('')

const feedsLoading = ref(false)
const postsLoading = ref(false)
const friendLinksLoading = ref(false)
const fetchingId = ref<number | null>(null)
const saving = ref(false)

const viewTitle = computed(() => (isAllPostsView.value ? '所有文章' : 'RSS 订阅源'))

// ---- Edit ----

const editDialogVisible = ref(false)
const editingFeed = ref<RssFeed | null>(null)
const editForm = reactive({
  name: '',
  feed_url: '',
  status: 'active' as RssFeedStatus,
  friend_link_id: 0
})

const handleEdit = (feed: RssFeed) => {
  editingFeed.value = feed
  Object.assign(editForm, {
    name: feed.name,
    feed_url: feed.feed_url,
    status: feed.status,
    friend_link_id: feed.friend_link_id ?? 0
  })
  editDialogVisible.value = true
}

const handleSave = async () => {
  const feed = editingFeed.value
  if (!feed) return

  // Only send what changed: touching feed_url or status resets the fetch cache.
  const payload: UpdateRssFeedPayload = {}
  if (editForm.name !== feed.name) payload.name = editForm.name
  if (editForm.feed_url !== feed.feed_url) payload.feed_url = editForm.feed_url
  if (editForm.friend_link_id !== (feed.friend_link_id ?? 0)) payload.friend_link_id = editForm.friend_link_id
  if (editForm.status !== feed.status && editForm.status !== 'dead') payload.status = editForm.status
  if (Object.keys(payload).length === 0) {
    editDialogVisible.value = false
    return
  }

  saving.value = true
  try {
    await updateRssFeed(feed.id, payload)
    ElMessage.success('已保存')
    editDialogVisible.value = false
    await fetchFeeds()
  } catch {
    // The request interceptor handles error messages
  } finally {
    saving.value = false
  }
}

const handleTogglePause = async (feed: RssFeed) => {
  const status = feed.status === 'paused' ? 'active' : 'paused'
  try {
    await updateRssFeed(feed.id, { status })
    ElMessage.success(status === 'paused' ? '已暂停' : '已恢复，稍后会重新抓取')
    await fetchFeeds()
  } catch {
    // The request interceptor handles error messages
  }
}

// ---- Subscribe ----

const createDialogVisible = ref(false)
const discovering = ref(false)
const discovered = ref(false)
const candidates = ref<FeedCandidate[]>([])
const createForm = reactive({
  site_url: '',
  feed_url: '',
  name: '',
  friend_link_id: 0
})

const handleCreate = () => {
  Object.assign(createForm, { site_url: '', feed_url: '', name: '', friend_link_id: 0 })
  candidates.value = []
  discovered.value = false
  createDialogVisible.value = true
}

// Picking an owner pre-fills its homepage so one click on "发现" is enough.
const handleCreateFriendLinkChange = (id: number) => {
  const link = friendLinks.value.find((item) => item.id === id)
  if (link) createForm.site_url = link.link
}

const handleDiscover = async () => {
  const url = createForm.site_url.trim()
  if (!url) {
    ElMessage.warning('先填网站地址')
    return
  }
  discovering.value = true
  try {
    const res = await discoverRssFeeds(url)
    candidates.value = res.data
    discovered.value = true
    const first = res.data[0]
    if (first) {
      createForm.feed_url = first.url
    }
  } catch {
    // The request interceptor handles error messages
  } finally {
    discovering.value = false
  }
}

const handleCreateSave = async () => {
  saving.value = true
  try {
    const res = await createRssFeed({
      feed_url: createForm.feed_url.trim(),
      name: createForm.name.trim(),
      friend_link_id: createForm.friend_link_id
    })
    ElMessage.success(`订阅成功：${res.data.feed.name}，导入 ${res.data.result.inserted_items} 篇文章`)
    createDialogVisible.value = false
    await fetchFeeds()
  } catch {
    // The request interceptor handles error messages
  } finally {
    saving.value = false
  }
}

// ---- Lists ----

const fetchFeeds = async () => {
  feedsLoading.value = true
  try {
    const res = await getRssFeeds({
      page: currentFeedPage.value,
      page_size: feedPageSize.value,
      status: statusFilter.value
    })
    feeds.value = res.data.items
    totalFeeds.value = res.data.total
    if (selectedFeed.value) {
      selectedFeed.value = feeds.value.find((feed) => feed.id === selectedFeed.value?.id) ?? selectedFeed.value
    }
  } catch {
    // The request interceptor handles error messages
  } finally {
    feedsLoading.value = false
  }
}

const fetchFriendLinkOptions = async () => {
  friendLinksLoading.value = true
  try {
    const res = await getFriendLinks({ page: 1, page_size: 1000 })
    friendLinks.value = res.data.items
  } catch {
    // The request interceptor handles error messages
  } finally {
    friendLinksLoading.value = false
  }
}

const fetchPosts = async () => {
  if (!selectedFeed.value) return
  postsLoading.value = true
  try {
    const res = await getPostsByFeed(selectedFeed.value.id, currentPostPage.value, postPageSize.value)
    posts.value = res.data.items
    totalPosts.value = res.data.total
  } catch {
    // The request interceptor handles error messages
  } finally {
    postsLoading.value = false
  }
}

const fetchAllPosts = async () => {
  postsLoading.value = true
  try {
    const res = await getAllPosts(currentPostPage.value, postPageSize.value)
    posts.value = res.data.items
    totalPosts.value = res.data.total
  } catch {
    // The request interceptor handles error messages
  } finally {
    postsLoading.value = false
  }
}

const fetchCurrentViewPosts = () => (isAllPostsView.value ? fetchAllPosts() : fetchPosts())

const {
  currentPage: currentFeedPage,
  pageSize: feedPageSize,
  total: totalFeeds,
  handlePageChange: handleFeedPageChange,
  handleSizeChange: handleFeedSizeChange
} = usePagination(fetchFeeds, 20)

const {
  currentPage: currentPostPage,
  pageSize: postPageSize,
  total: totalPosts,
  handlePageChange: handlePostPageChange,
  handleSizeChange: handlePostSizeChange,
  reset: resetPostPagination
} = usePagination(fetchCurrentViewPosts, 20)

const handleStatusFilterChange = () => {
  currentFeedPage.value = 1
  activeFeedId.value = ''
  selectedFeed.value = null
  fetchFeeds()
}

const handleFeedChange = (name: CollapseModelValue) => {
  // Accordion mode emits a single name, but the typing allows an array.
  const activeName = Array.isArray(name) ? name[0] : name
  if (activeName === '' || activeName === undefined) {
    selectedFeed.value = null
    posts.value = []
    return
  }
  const feed = feeds.value.find((item) => item.id === Number(activeName))
  if (!feed) return
  selectedFeed.value = feed
  resetPostPagination()
  fetchPosts()
}

const showAllPosts = () => {
  isAllPostsView.value = true
  selectedFeed.value = null
  activeFeedId.value = ''
  resetPostPagination()
  fetchAllPosts()
}

const showFeedList = () => {
  isAllPostsView.value = false
  selectedFeed.value = null
  activeFeedId.value = ''
  posts.value = []
  resetPostPagination()
}

const handleImmediateFetch = async (feed: RssFeed) => {
  if (fetchingId.value !== null) return

  fetchingId.value = feed.id
  try {
    const res = await fetchRssFeed(feed.id)
    ElMessage.success(`获取完成：检查 ${res.data.checked_items} 篇，新增 ${res.data.inserted_items} 篇`)
    if (isAllPostsView.value || selectedFeed.value?.id === feed.id) {
      await fetchCurrentViewPosts()
    }
  } catch {
    // The request interceptor handles error messages
  } finally {
    fetchingId.value = null
    // Failures are recorded on the feed too, so refresh either way.
    await fetchFeeds()
  }
}

const handleDelete = (feed: RssFeed) => {
  ElMessageBox.confirm(`确定要删除订阅源 "${feed.name}" 吗？它的文章会一起删除。`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteRssFeed(feed.id)
      ElMessage.success('删除成功')
      if (selectedFeed.value?.id === feed.id) {
        selectedFeed.value = null
        activeFeedId.value = ''
        posts.value = []
      }
      await fetchFeeds()
    } catch {
      // The request interceptor handles error messages
    }
  })
}

const handleDeletePost = (post: RssPost) => {
  ElMessageBox.confirm(`确定要删除文章“${post.title}”吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteRssPost(post.id)
      ElMessage.success('文章已删除')
      if (posts.value.length === 1 && currentPostPage.value > 1) {
        currentPostPage.value--
      }
      await fetchCurrentViewPosts()
    } catch {
      // The request interceptor handles error messages
    }
  })
}

// ---- Formatting ----

const friendLinkName = (friendLinkId: number | null) => {
  if (!friendLinkId) return '无归属'
  return friendLinks.value.find((link) => link.id === friendLinkId)?.name ?? `#${friendLinkId}`
}

const formatInterval = (seconds: number) => {
  if (seconds >= 3600) return `${Math.round((seconds / 3600) * 10) / 10} 小时`
  return `${Math.max(1, Math.round(seconds / 60))} 分钟`
}

const formatNextFetch = (timestamp: number) => {
  if (timestamp <= Date.now() / 1000) return '待抓取'
  return formatDate(timestamp, 'MM-DD HH:mm')
}

onMounted(() => {
  fetchFeeds()
  fetchFriendLinkOptions()
})
</script>

<style scoped>
.content-scroll {
  flex: 1;
  min-height: 0;
}

.feed-title {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding-right: 12px;
  font-weight: normal;
}

.feed-link {
  display: inline-flex;
  align-items: center;
  color: inherit;
}

.feed-name {
  font-weight: 600;
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feed-name-link {
  color: inherit;
  text-decoration: none;
}

.feed-name-link:hover {
  color: #409eff;
}

.feed-meta {
  font-size: 12px;
  color: #909399;
}

.feed-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 4px;
}

.discover-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

.candidate-list {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 6px;
  width: 100%;
}

.candidate {
  height: auto;
  margin-right: 0;
  padding: 6px 0;
  align-items: flex-start;
}

.candidate-body {
  line-height: 1.5;
}

.candidate-title {
  display: flex;
  align-items: center;
  gap: 6px;
}

.candidate-url {
  font-size: 12px;
  color: #909399;
  word-break: break-all;
  white-space: normal;
}

.dialog-alert {
  margin-top: 8px;
}

.pagination-container {
  margin-top: 16px;
}

.feed-pagination {
  padding-top: 4px;
  border-top: 1px solid var(--el-border-color);
}
</style>
