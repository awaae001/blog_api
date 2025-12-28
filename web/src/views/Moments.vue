<template>
  <div class="moments-page" v-loading="pageLoading" element-loading-text="加载中...">
    <el-card class="composer-card">
      <template #header>
        <div class="card-header">
          <span>我的动态</span>
          <div class="header-actions">
            <el-select v-model="composer.uploadTarget" size="small" style="width: 120px">
              <el-option label="本地存储" value="local" />
              <el-option label="OSS 存储" value="oss" />
            </el-select>
            <el-button type="primary" :loading="actionLoading" @click="handleCreateMoment">
              发布
            </el-button>
          </div>
        </div>
      </template>

      <el-input
        v-model="composer.content"
        type="textarea"
        :rows="3"
        placeholder="写下今天的动态..."
        maxlength="500"
        show-word-limit
        class="mb-2"
      />
      
      <el-input
        v-model="composer.message_link"
        placeholder="来源链接 (可选)"
        :prefix-icon="LinkIcon"
        clearable
        class="mb-2"
      />

      <div class="composer-toolbar">
        <el-upload
          :auto-upload="false"
          :show-file-list="false"
          :on-change="handleComposerFileChange"
          accept="image/*,video/*"
          multiple
        >
          <el-button size="small" :loading="uploading">上传图片/视频</el-button>
        </el-upload>
        <span class="upload-hint">先本地预览，发布时再上传</span>
      </div>

      <div v-if="composer.mediaItems.length" class="media-grid">
        <div v-for="(item, index) in composer.mediaItems" :key="item.id" class="media-item">
          <el-image
            v-if="item.media_type === 'image'"
            :src="item.previewUrl"
            fit="cover"
            class="media-thumb"
            :preview-src-list="[item.previewUrl]"
            preview-teleported
          />
          <div v-else class="media-video">
            <el-icon><VideoCamera /></el-icon>
            <span>视频</span>
          </div>
          <el-button link type="danger" @click="removeComposerMedia(index)">移除</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="list-card">
      <div class="list-filters">
        <el-select v-model="filters.status" placeholder="状态筛选" clearable style="width: 140px" @change="handleFilter">
          <el-option label="全部" value="" />
          <el-option label="可见" value="visible" />
          <el-option label="隐藏" value="hidden" />
          <el-option label="已删除" value="deleted" />
        </el-select>
      </div>

      <el-scrollbar height="68vh">
        <div v-if="!moments.length" class="empty-state">
          还没有动态，先发一条吧。
        </div>
        
        <div class="waterfall-container">
          <div v-for="(col, colIndex) in waterfallColumns" :key="colIndex" class="waterfall-column">
            <div
              v-for="moment in col"
              :key="moment.id"
              class="moment-card"
              :class="{ 'clickable': !!moment.message_link }"
              @click="handleCardClick(moment, $event)"
            >
              <!-- Media Display (Top) -->
              <div v-if="moment.media.length" class="moment-media-grid" :class="`grid-${Math.min(moment.media.length, 4)}`">
                <div v-for="(media, idx) in moment.media.slice(0, 4)" :key="media.id" class="media-cell">
                  <el-image
                    v-if="media.media_type === 'image'"
                    :src="media.media_url"
                    fit="cover"
                    class="media-content"
                    :preview-src-list="moment.media.map(m => m.media_url)"
                    :initial-index="idx"
                    preview-teleported
                    @click.stop
                  />
                  <div v-else class="media-video-placeholder">
                    <el-icon><VideoCamera /></el-icon>
                  </div>
                  <div v-if="idx === 3 && moment.media.length > 4" class="more-media-mask">
                    +{{ moment.media.length - 4 }}
                  </div>
                </div>
              </div>

              <div class="moment-body">
                <div class="moment-content">{{ moment.content }}</div>
                
                <div class="moment-footer">
                  <div class="moment-info">
                    <span class="moment-time">{{ formatTime(moment.created_at) }}</span>
                    <el-tag size="small" :type="statusTagType(moment.status)" effect="plain" class="status-tag">
                      {{ moment.status }}
                    </el-tag>
                  </div>
                  
                  <div class="moment-source" v-if="getSourceInfo(moment)">
                    <el-tooltip :content="getSourceInfo(moment)?.label" placement="top">
                      <div class="source-icon" :class="getSourceInfo(moment)?.icon">
                        <template v-if="getSourceInfo(moment)?.icon === 'tg'">
                          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 11.944 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"/></svg>
                        </template>
                        <template v-else-if="getSourceInfo(moment)?.icon === 'dc'">
                          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M20.317 4.3698a19.7913 19.7913 0 00-4.8851-1.5152.0741.0741 0 00-.0785.0371c-.211.3753-.4447.8648-.6083 1.2495-1.8447-.2762-3.68-.2762-5.4868 0-.1636-.3933-.4058-.8742-.6177-1.2495a.077.077 0 00-.0785-.037 19.7363 19.7363 0 00-4.8852 1.515.0699.0699 0 00-.0321.0277C.5334 9.0458-.319 13.5799.0992 18.0578a.0824.0824 0 00.0312.0561c2.0528 1.5076 4.0413 2.4228 5.9929 3.0294a.0777.0777 0 00.0842-.0276c.4616-.6304.8731-1.2952 1.226-1.9942a.076.076 0 00-.0416-.1057c-.6528-.2476-1.2743-.5495-1.8722-.8923a.077.077 0 01-.0076-.1277c.1258-.0943.2517-.1923.3718-.2914a.0743.0743 0 01.0776-.0105c3.9278 1.7933 8.18 1.7933 12.0614 0a.0739.0739 0 01.0785.0095c.1202.099.246.1981.3728.2924a.077.077 0 01-.0066.1276 12.2986 12.2986 0 01-1.873.8914.0766.0766 0 00-.0407.1067c.3604.698.7719 1.3628 1.225 1.9932a.076.076 0 00.0842.0286c1.961-.6067 3.9495-1.5219 6.0023-3.0294a.077.077 0 00.0313-.0552c.5004-5.177-.8382-9.6739-3.5485-13.6604a.061.061 0 00-.0312-.0286zM8.02 15.3312c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9555-2.4189 2.157-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.946 2.419-2.1568 2.419zm7.9748 0c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9554-2.4189 2.1569-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.946 2.419-2.1568 2.419z"/></svg>
                        </template>
                        <template v-else>
                          <el-icon><LinkIcon /></el-icon>
                        </template>
                      </div>
                    </el-tooltip>
                  </div>
                </div>

                <div class="moment-actions-bar">
                  <el-button link size="small" :icon="Edit" @click.stop="openEditDialog(moment)">编辑</el-button>
                  <el-popconfirm title="确定要删除这条动态吗？" @confirm="handleDeleteMoment(moment)">
                    <template #reference>
                      <el-button link size="small" type="danger" :icon="Delete" @click.stop>删除</el-button>
                    </template>
                  </el-popconfirm>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-scrollbar>

      <el-pagination
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-sizes="[10, 20, 50]"
        :page-size="pageSize"
        :current-page="currentPage"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        class="pagination"
      />
    </el-card>

    <el-dialog v-model="editDialogVisible" title="编辑动态" width="640px" @close="resetEditForm">
      <el-form label-width="70px">
        <el-form-item label="内容">
          <el-input v-model="editForm.content" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="editForm.status" style="width: 160px">
            <el-option label="可见" value="visible" />
            <el-option label="隐藏" value="hidden" />
            <el-option label="已删除" value="deleted" />
          </el-select>
        </el-form-item>
        <el-form-item label="来源链接">
          <div class="source-edit">
            <el-input v-model="editForm.message_link" placeholder="https://..." />
            <el-button
              v-if="editForm.message_link"
              class="source-link"
              link
              type="primary"
              @click="openSourceLink(editForm.message_link)"
            >
              打开
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="Guild ID">
          <el-input v-model="editForm.guild_id" disabled />
        </el-form-item>
        <el-form-item label="Channel ID">
          <el-input v-model="editForm.channel_id" disabled />
        </el-form-item>
        <el-form-item label="Message ID">
          <el-input v-model="editForm.message_id" disabled />
        </el-form-item>
        <el-form-item label="媒体">
          <div class="edit-media-toolbar">
            <el-select v-model="editUploadTarget" size="small" style="width: 120px">
              <el-option label="本地存储" value="local" />
              <el-option label="OSS 存储" value="oss" />
            </el-select>
            <el-upload
              :auto-upload="false"
              :show-file-list="false"
              :on-change="handleEditFileChange"
              accept="image/*,video/*"
            >
              <el-button size="small" :loading="uploading">添加媒体</el-button>
            </el-upload>
            <span class="upload-hint">先本地预览，保存时再上传</span>
          </div>
          <div v-if="editForm.media.length" class="media-grid">
            <div v-for="media in editForm.media" :key="media.id" class="media-item">
              <el-image
                v-if="media.media_type === 'image'"
                :src="media.media_url"
                fit="cover"
                class="media-thumb"
                :preview-src-list="[media.media_url]"
                preview-teleported
              />
              <div v-else class="media-video">
                <el-icon><VideoCamera /></el-icon>
                <span>视频</span>
              </div>
              <el-button link type="danger" @click="handleDeleteMedia(media)">移除</el-button>
            </div>
          </div>
          <div v-if="editPendingMedia.length" class="media-grid">
            <div v-for="item in editPendingMedia" :key="item.id" class="media-item">
              <el-image
                v-if="item.media_type === 'image'"
                :src="item.previewUrl"
                fit="cover"
                class="media-thumb"
                :preview-src-list="[item.previewUrl]"
                preview-teleported
              />
              <div v-else class="media-video">
                <el-icon><VideoCamera /></el-icon>
                <span>视频</span>
              </div>
              <el-tag size="small" type="info">待上传</el-tag>
              <el-button link type="danger" @click="removeEditPendingMedia(item.id)">移除</el-button>
            </div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="actionLoading" @click="handleUpdateMoment">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Edit, VideoCamera, Link as LinkIcon } from '@element-plus/icons-vue'
import { usePagination } from '@/utils/pagination'
import { uploadFile } from '@/api/resource'
import {
  getMoments,
  createMoment,
  updateMoment,
  deleteMoment,
  createMomentMedia,
  deleteMomentMedia
} from '@/api/moment'
import type { MomentWithMedia, MomentMedia, CreateMomentPayload } from '@/model/moment'
import type { UploadFile } from 'element-plus'

type UploadTarget = 'local' | 'oss'

const moments = ref<MomentWithMedia[]>([])
const pageLoading = ref(false)
const actionLoading = ref(false)
const uploading = ref(false)
const filters = reactive({
  status: ''
})

type ComposerMediaItem = {
  id: string
  file: UploadFile
  previewUrl: string
  media_type: 'image' | 'video'
}

const composer = reactive({
  content: '',
  message_link: '',
  mediaItems: [] as ComposerMediaItem[],
  uploadTarget: 'local' as UploadTarget
})

const columnCount = ref(3)
const updateColumnCount = () => {
  const width = window.innerWidth
  if (width < 768) {
    columnCount.value = 1
  } else if (width < 1200) {
    columnCount.value = 2
  } else {
    columnCount.value = 3
  }
}

const waterfallColumns = computed(() => {
  if (columnCount.value === 1) return [moments.value]
  const cols = Array.from({ length: columnCount.value }, () => [] as MomentWithMedia[])
  moments.value.forEach((moment, index) => {
    cols[index % columnCount.value].push(moment)
  })
  return cols
})

const editDialogVisible = ref(false)
const editUploadTarget = ref<UploadTarget>('local')
const editPendingMedia = ref<ComposerMediaItem[]>([])
const editForm = reactive({
  id: 0,
  content: '',
  status: 'visible' as 'visible' | 'hidden' | 'deleted',
  message_link: '',
  guild_id: '',
  channel_id: '',
  message_id: '',
  media: [] as MomentMedia[]
})

const { currentPage, pageSize, total, handlePageChange, handleSizeChange, reset } = usePagination(
  () => fetchMoments(),
  10
)

const fetchMoments = async () => {
  pageLoading.value = true
  try {
    const res = await getMoments({
      page: currentPage.value,
      page_size: pageSize.value,
      status: filters.status
    })
    moments.value = res.data.moments
    total.value = res.data.total
  } catch (error) {
    console.error(error)
  } finally {
    pageLoading.value = false
  }
}

const handleFilter = () => {
  reset()
  fetchMoments()
}

const formatTime = (timestamp: number) => {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleString()
}

const statusTagType = (status: string) => {
  switch (status) {
    case 'visible':
      return 'success'
    case 'hidden':
      return 'warning'
    case 'deleted':
      return 'info'
    default:
      return ''
  }
}

const getSourceInfo = (moment: MomentWithMedia) => {
  if (moment.message_link) {
    if (moment.message_link.includes('t.me')) {
      return { label: 'Telegram', icon: 'tg', link: moment.message_link }
    }
    if (moment.message_link.includes('discord.com')) {
      return { label: 'Discord', icon: 'dc', link: moment.message_link }
    }
    return { label: '来源', icon: 'link', link: moment.message_link }
  }
  if (moment.guild_id && moment.guild_id > 0) {
    return { label: 'Discord', icon: 'dc', link: '' }
  }
  if (moment.channel_id && moment.channel_id > 0) {
    return { label: 'Telegram', icon: 'tg', link: '' }
  }
  return null
}

const handleCardClick = (moment: MomentWithMedia, e: MouseEvent) => {
  // 如果点击的是按钮或图片，不跳转
  const target = e.target as HTMLElement
  if (target.closest('.el-button') || target.closest('.el-image') || target.closest('.el-popconfirm')) {
    return
  }
  if (moment.message_link) {
    window.open(moment.message_link, '_blank', 'noopener')
  }
}

const openSourceLink = (link: string) => {
  if (!link) return
  window.open(link, '_blank', 'noopener')
}

const getMediaType = (file: UploadFile): 'image' | 'video' => {
  const mime = file.raw?.type || ''
  if (mime.startsWith('video/')) {
    return 'video'
  }
  return 'image'
}

const getDatePath = () => {
  const now = new Date()
  const yy = String(now.getFullYear()).slice(2)
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  return `${yy}${mm}${dd}`
}

const uploadMediaFile = async (
  file: UploadFile,
  target: UploadTarget,
  basePath: string
): Promise<{ url: string; mediaType: 'image' | 'video' }> => {
  if (!file.raw) {
    throw new Error('无效文件')
  }
  const formData = new FormData()
  formData.append('file', file.raw)
  formData.append('path', basePath)
  const res = await uploadFile(formData, target)
  return {
    url: res.data.url,
    mediaType: getMediaType(file)
  }
}

const handleComposerFileChange = (file: UploadFile) => {
  if (!file.raw) return
  const previewUrl = URL.createObjectURL(file.raw)
  composer.mediaItems.push({
    id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
    file,
    previewUrl,
    media_type: getMediaType(file)
  })
}

const removeComposerMedia = (index: number) => {
  const item = composer.mediaItems[index]
  if (item?.previewUrl) {
    URL.revokeObjectURL(item.previewUrl)
  }
  composer.mediaItems.splice(index, 1)
}

const uploadComposerMedia = async (
  basePath: string
): Promise<Array<{ media_url: string; media_type: 'image' | 'video' }>> => {
  const uploaded: Array<{ media_url: string; media_type: 'image' | 'video' }> = []
  for (const item of composer.mediaItems) {
    if (!item.file.raw) continue
    const result = await uploadMediaFile(item.file, composer.uploadTarget, basePath)
    uploaded.push({ media_url: result.url, media_type: result.mediaType })
  }
  return uploaded
}

const handleCreateMoment = async () => {
  if (!composer.content.trim()) {
    ElMessage.error('请输入动态内容')
    return
  }
  actionLoading.value = true
  try {
    uploading.value = true
    const basePath = `moments/${getDatePath()}`
    const uploadedMedia = await uploadComposerMedia(basePath)
    const payload: CreateMomentPayload = {
      content: composer.content.trim(),
      message_link: composer.message_link.trim() || undefined,
      media: uploadedMedia
    }
    await createMoment(payload)
    ElMessage.success('发布成功')
    composer.mediaItems.forEach((item) => {
      if (item.previewUrl) {
        URL.revokeObjectURL(item.previewUrl)
      }
    })
    composer.content = ''
    composer.message_link = ''
    composer.mediaItems = []
    reset()
    fetchMoments()
  } catch (error) {
    console.error(error)
    ElMessage.error('上传失败，请重试')
  } finally {
    uploading.value = false
    actionLoading.value = false
  }
}

const openEditDialog = (moment: MomentWithMedia) => {
  editPendingMedia.value.forEach((item) => {
    if (item.previewUrl) {
      URL.revokeObjectURL(item.previewUrl)
    }
  })
  editPendingMedia.value = []
  editForm.id = moment.id
  editForm.content = moment.content
  editForm.status = moment.status
  editForm.message_link = moment.message_link || ''
  editForm.guild_id = moment.guild_id ? String(moment.guild_id) : ''
  editForm.channel_id = moment.channel_id ? String(moment.channel_id) : ''
  editForm.message_id = moment.message_id ? String(moment.message_id) : ''
  editForm.media = moment.media.map((item) => ({ ...item }))
  editDialogVisible.value = true
}

const resetEditForm = () => {
  editForm.id = 0
  editForm.content = ''
  editForm.status = 'visible'
  editForm.message_link = ''
  editForm.guild_id = ''
  editForm.channel_id = ''
  editForm.message_id = ''
  editForm.media = []
  editUploadTarget.value = 'local'
  editPendingMedia.value.forEach((item) => {
    if (item.previewUrl) {
      URL.revokeObjectURL(item.previewUrl)
    }
  })
  editPendingMedia.value = []
}

const handleUpdateMoment = async () => {
  if (!editForm.id) return
  actionLoading.value = true
  try {
    await updateMoment(editForm.id, {
      content: editForm.content,
      status: editForm.status,
      message_link: editForm.message_link || undefined
    })
    if (editPendingMedia.value.length) {
      uploading.value = true
      const pendingItems = [...editPendingMedia.value]
      const basePath = `moments/${getDatePath()}`
      for (const item of pendingItems) {
        const result = await uploadMediaFile(item.file, editUploadTarget.value, basePath)
        const res = await createMomentMedia({
          moment_id: editForm.id,
          media_url: result.url,
          media_type: result.mediaType,
          is_local: editUploadTarget.value === 'local' ? 1 : 0
        })
        editForm.media.push(res.data)
        const target = moments.value.find((entry) => entry.id === editForm.id)
        if (target) {
          target.media.push(res.data)
        }
        removeEditPendingMedia(item.id)
      }
    }
    const target = moments.value.find((item) => item.id === editForm.id)
    if (target) {
      target.content = editForm.content
      target.status = editForm.status
      target.message_link = editForm.message_link || ''
      target.media = editForm.media.map((item) => ({ ...item }))
    }
    ElMessage.success('更新成功')
    editDialogVisible.value = false
  } catch (error) {
    console.error(error)
    ElMessage.error('更新失败，请重试')
  } finally {
    uploading.value = false
    actionLoading.value = false
  }
}

const handleDeleteMoment = async (moment: MomentWithMedia) => {
  actionLoading.value = true
  try {
    await deleteMoment(moment.id)
    ElMessage.success('已删除')
    fetchMoments()
  } catch (error) {
    console.error(error)
  } finally {
    actionLoading.value = false
  }
}

const handleEditFileChange = (file: UploadFile) => {
  if (!editForm.id) return
  if (!file.raw) return
  const previewUrl = URL.createObjectURL(file.raw)
  editPendingMedia.value.push({
    id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
    file,
    previewUrl,
    media_type: getMediaType(file)
  })
}

const removeEditPendingMedia = (id: string) => {
  const index = editPendingMedia.value.findIndex((item) => item.id === id)
  if (index === -1) return
  const item = editPendingMedia.value[index]
  if (item.previewUrl) {
    URL.revokeObjectURL(item.previewUrl)
  }
  editPendingMedia.value.splice(index, 1)
}

const handleDeleteMedia = async (media: MomentMedia) => {
  actionLoading.value = true
  try {
    await deleteMomentMedia(media.id)
    editForm.media = editForm.media.filter((item) => item.id !== media.id)
    const target = moments.value.find((item) => item.id === editForm.id)
    if (target) {
      target.media = target.media.filter((item) => item.id !== media.id)
    }
    ElMessage.success('已移除媒体')
  } catch (error) {
    console.error(error)
  } finally {
    actionLoading.value = false
  }
}

onMounted(() => {
  fetchMoments()
  updateColumnCount()
  window.addEventListener('resize', updateColumnCount)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateColumnCount)
})
</script>

<style scoped>
.moments-page {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 6px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.composer-card {
  border-radius: 10px;
}

.composer-toolbar {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.upload-hint {
  color: #909399;
  font-size: 12px;
}

.media-grid {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.media-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
}

.media-thumb {
  width: 120px;
  height: 120px;
  border-radius: 8px;
}

.media-video {
  width: 120px;
  height: 120px;
  border-radius: 8px;
  border: 1px dashed var(--el-border-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
  gap: 6px;
}

.list-card {
  border-radius: 10px;
}

.list-filters {
  margin-bottom: 12px;
  display: flex;
  justify-content: flex-end;
}

.mb-2 {
  margin-bottom: 8px;
}

.waterfall-container {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.waterfall-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0; /* Prevent flex item from overflowing */
}

.moment-card {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #ebeef5;
  overflow: hidden;
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
}

.moment-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transform: translateY(-2px);
}

.moment-card.clickable {
  cursor: pointer;
}

.moment-media-grid {
  display: grid;
  gap: 2px;
  width: 100%;
  aspect-ratio: 16/9;
}

.moment-media-grid.grid-1 { grid-template-columns: 1fr; }
.moment-media-grid.grid-2 { grid-template-columns: 1fr 1fr; }
.moment-media-grid.grid-3 { grid-template-columns: 1fr 1fr; grid-template-rows: 1fr 1fr; }
.moment-media-grid.grid-3 .media-cell:first-child { grid-row: span 2; }
.moment-media-grid.grid-4 { grid-template-columns: 1fr 1fr; grid-template-rows: 1fr 1fr; }

.media-cell {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.media-content {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.media-video-placeholder {
  width: 100%;
  height: 100%;
  background: #f5f7fa;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.more-media-mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: bold;
}

.moment-body {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.moment-content {
  white-space: pre-wrap;
  color: #303133;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-word;
}

.moment-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 4px;
}

.moment-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.moment-time {
  color: #909399;
  font-size: 12px;
}

.status-tag {
  height: 20px;
  padding: 0 6px;
  font-size: 11px;
}

.moment-source {
  display: flex;
  align-items: center;
}

.source-icon {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #f0f2f5;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  transition: all 0.2s;
}

.source-icon.tg { color: #2481cc; background: rgba(36, 129, 204, 0.1); }
.source-icon.dc { color: #5865f2; background: rgba(88, 101, 242, 0.1); }
.source-icon.link { color: #606266; }

.moment-actions-bar {
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid #f0f2f5;
  padding-top: 8px;
  margin-top: 4px;
}

.source-edit {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.source-edit .el-input {
  flex: 1;
}

.empty-state {
  text-align: center;
  color: #909399;
  padding: 40px 0;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.edit-media-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}
</style>
