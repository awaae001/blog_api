<template>
  <PageShell title="友链管理">
    <template #actions>
      <el-button type="success" plain :icon="Refresh" :loading="inspectionRunning" @click="handleRecheckAll">
        {{ inspectionRunning ? `全量巡查中 ${progress?.processed ?? 0}/${progress?.total ?? 0}` : '全量巡查' }}
      </el-button>
      <el-button :icon="PriceTag" @click="openTagDrawer(null)">标签管理</el-button>
      <el-button type="primary" :icon="Plus" @click="openFormDialog()">
        新增友链
      </el-button>
    </template>

    <!-- Inspection progress -->
    <el-alert v-if="progress && progress.total > 0" class="inspection-alert" :type="inspectionAlertType"
      :closable="!inspectionRunning" show-icon>
      <template #title>
        {{ progress.label || '巡查' }}{{ inspectionRunning ? '进行中' : '已结束' }}：
        {{ progress.processed }}/{{ progress.total }}
        （存活 {{ progress.survival }}，失败 {{ progress.failed }}，新增 RSS {{ progress.rss_discovered }}）
        {{ progress.error ? `，异常：${progress.error}` : '' }}
      </template>
    </el-alert>

    <!-- Filter and Actions -->
    <div class="table-actions">
      <el-select v-model="filterIsDied" placeholder="按失效状态筛选" clearable @change="handleFilter"
        style="width: 150px; margin-right: 10px">
        <el-option label="已失效" :value="true"></el-option>
        <el-option label="未失效" :value="false"></el-option>
      </el-select>
      <el-select v-model="filterStatus" placeholder="按状态筛选" clearable @change="handleFilter"
        style="width: 150px; margin-right: 10px">
        <el-option label="正常" value="survival"></el-option>
        <el-option label="待定" value="pending"></el-option>
        <el-option label="超时" value="timeout"></el-option>
        <el-option label="错误" value="error"></el-option>
      </el-select>
      <el-input v-model="searchQuery" placeholder="搜索友链" clearable @input="handleFilter"
        style="width: 200px; margin-right: 10px" />
      <el-select v-model="filterTagId" placeholder="按标签筛选" clearable filterable :value-on-clear="null"
        style="width: 150px" @change="handleFilter">
        <el-option v-for="tag in tags" :key="tag.id" :label="tag.name" :value="tag.id" />
      </el-select>
    </div>

    <!-- Friend Link Table -->
    <el-scrollbar class="table-scroll">
      <el-table :data="friendLinks" v-loading="loading" style="width: 100%" :row-class-name="rowClassName">
        <!-- @vue-generic {FriendLink} -->
        <el-table-column prop="name" label="网站名称" width="200">
          <template #default="{ row }">
            <div class="link-cell">
              <el-avatar :size="20" :src="row.avatar" class="link-avatar">{{ row.name.slice(0, 1) }}</el-avatar>
              <a class="link-name" :class="nameClass(row)" :href="row.link" target="_blank" rel="noopener"
                :title="row.link">{{ row.name }}</a>
            </div>
          </template>
        </el-table-column>
        <!-- @vue-generic {FriendLink} -->
        <el-table-column label="标签" min-width="120">
          <template #default="{ row }">
            <div class="tag-cell">
              <el-tag v-for="tag in row.tags" :key="tag.id" size="small" :color="tag.color || undefined"
                effect="light" class="tag-item" title="点击编辑" @click="openTagDrawer(null, tag)">
                {{ tag.name }}
              </el-tag>
              <el-button class="tag-add" size="small" :icon="Plus" link @click="openTagDrawer(row)" />
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="times" label="失败次数" width="100" />
        <!-- @vue-generic {FriendLink} -->
        <el-table-column prop="reason" label="探针问题" width="220" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.reason || '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.updated_at) }}
          </template>
        </el-table-column>
        <!-- @vue-generic {FriendLink} -->
        <el-table-column label="跳过" width="56" align="center">
          <template #default="{ row }">
            <el-checkbox :model-value="row.skip_health_check" @change="handleHealthCheckToggle(row)" />
          </template>
        </el-table-column>
        <!-- @vue-generic {FriendLink} -->
        <el-table-column label="RSS" width="56" align="center">
          <template #default="{ row }">
            <el-checkbox :model-value="row.enable_rss" @change="handleRssToggle(row)" />
          </template>
        </el-table-column>
        <!-- @vue-generic {FriendLink} -->
        <el-table-column label="操作" width="190" fixed="right">
          <template #default="{ row }">
            <el-tooltip content="重新巡查" placement="top">
              <el-button type="success" link :icon="Refresh" :loading="recheckingId === row.id"
                :disabled="inspectionRunning || (recheckingId !== null && recheckingId !== row.id)"
                @click="handleRecheck(row.id)" />
            </el-tooltip>
            <el-button type="primary" link :icon="Edit" @click="openFormDialog(row)">
              编辑
            </el-button>
            <el-button type="danger" link :icon="Delete" @click="handleDelete(row.id)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-scrollbar>

    <TablePagination :total="total" :page-size="pageSize" :current-page="currentPage" @size-change="handleSizeChange"
      @current-change="handlePageChange" />

    <!-- Form Dialog for Add/Edit -->
    <el-dialog :title="isEditMode ? '编辑友链' : '新增友链'" v-model="dialogVisible" width="500px" @close="resetForm">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="网站名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="网站链接" prop="link">
          <el-input v-model="form.link" />
        </el-form-item>
        <el-form-item label="网站图标" prop="avatar">
          <el-input v-model="form.avatar" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input type="textarea" v-model="form.description" />
        </el-form-item>
        <el-form-item label="站长邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="订阅 RSS" prop="enable_rss">
          <el-switch v-model="form.enable_rss" />
        </el-form-item>
        <el-form-item label="是否失效" prop="is_died" v-if="isEditMode">
          <el-switch v-model="form.is_died" />
        </el-form-item>
        <el-form-item label="失败次数" prop="times" v-if="isEditMode">
          <el-input-number v-model="form.times" :min="0" />
        </el-form-item>
        <el-form-item label="状态" prop="status" v-if="isEditMode">
          <el-select v-model="form.status">
            <el-option label="正常" value="survival"></el-option>
            <el-option label="待定" value="pending"></el-option>
            <el-option label="超时" value="timeout"></el-option>
            <el-option label="错误" value="error"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="创建时间" v-if="isEditMode">
          <span>{{ formatDate(editingCreatedAt) }}</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <FriendTagDrawer v-model="tagDrawerVisible" :link="tagDrawerLink" :focus="tagDrawerFocus" :tags="tags"
      @changed="refreshTags" />
  </PageShell>
</template>

<script setup lang="ts">
import PageShell from '@/components/PageShell.vue'
import TablePagination from '@/components/TablePagination.vue'
import FriendTagDrawer from '@/components/FriendTagDrawer.vue'
import { ref, onMounted, onUnmounted, computed, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit, Refresh, Link, PriceTag } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getFriendLinks,
  createFriendLink,
  updateFriendLink,
  deleteFriendLink,
  recheckFriendLink,
  recheckAllFriendLinks,
  getRecheckProgress,
  getFriendTags,
  type FriendLink,
  type FriendTag,
  type FriendLinkInspectionProgress
} from '@/api/friend-links'
import { usePagination } from '@/utils/pagination'
import { formatDate } from '@/utils/date'

definePage({ path: '/friend', meta: { title: '友链', icon: Link, order: 3 } })

// Reactive State
const friendLinks = ref<FriendLink[]>([])
const loading = ref(false)
const recheckingId = ref<number | null>(null)
const progress = ref<FriendLinkInspectionProgress | null>(null)
const inspectionRunning = computed(() => progress.value?.running === true)
const inspectionAlertType = computed(() => {
  if (progress.value?.error || progress.value?.database_failures) return 'error'
  if (inspectionRunning.value) return 'info'
  return progress.value?.failed ? 'warning' : 'success'
})
const filterStatus = ref('')
const filterIsDied = ref<boolean | null>(null)
const searchQuery = ref('')
const filterTagId = ref<number | null>(null)
const tags = ref<FriendTag[]>([])
const tagDrawerVisible = ref(false)
const tagDrawerLink = ref<FriendLink | null>(null)
const tagDrawerFocus = ref<FriendTag | null>(null)
const dialogVisible = ref(false)
const isEditMode = ref(false)
const editingCreatedAt = ref(0)
const formRef = ref<FormInstance>()
const form = reactive<{
  id: number
  name: string
  link: string
  avatar: string
  description: string
  email: string
  times: number
  status: 'survival' | 'timeout' | 'error' | 'pending'
  enable_rss: boolean
  skip_health_check: boolean
  is_died: boolean
}>({
  id: 0,
  name: '',
  link: '',
  avatar: '',
  description: '',
  email: '',
  times: 0,
  status: 'survival',
  enable_rss: true,
  skip_health_check: false,
  is_died: false
})

const rules = reactive<FormRules>({
  name: [{ required: true, message: '请输入网站名称', trigger: 'blur' }],
  link: [{ required: true, message: '请输入网站链接', trigger: 'blur' }]
})

// Pagination
const { currentPage, pageSize, total, handlePageChange, handleSizeChange, reset } = usePagination(
  () => fetchFriendLinks(),
  10
)

// Fetch data
const fetchFriendLinks = async () => {
  loading.value = true
  try {
    const res = await getFriendLinks({
      page: currentPage.value,
      page_size: pageSize.value,
      status: filterStatus.value,
      search: searchQuery.value,
      is_died: filterIsDied.value === null ? undefined : filterIsDied.value,
      tag_id: filterTagId.value ?? undefined
    })
    friendLinks.value = res.data.items
    total.value = res.data.total
  } catch {
    // Request errors are reported by the interceptor.
  } finally {
    loading.value = false
  }
}

const fetchTags = async () => {
  try {
    tags.value = (await getFriendTags()).data
  } catch {
    // Request errors are reported by the interceptor.
  }
}

const refreshTags = async () => {
  await fetchTags()
  if (filterTagId.value && !tags.value.some((t) => t.id === filterTagId.value)) {
    filterTagId.value = null
    reset()
  }
  fetchFriendLinks()
}

onMounted(() => {
  fetchFriendLinks()
  fetchTags()
  refreshProgress()
})

onUnmounted(stopProgressPolling)

// Table and Actions
const handleFilter = () => {
  reset()
  fetchFriendLinks()
}

// Dialog and Form
const openFormDialog = (link?: FriendLink) => {
  editingCreatedAt.value = link?.created_at ?? 0
  if (link) {
    isEditMode.value = true
    Object.assign(form, {
      id: link.id,
      name: link.name,
      link: link.link,
      avatar: link.avatar,
      description: link.description,
      email: link.email ?? '',
      times: link.times ?? 0,
      status: link.status,
      enable_rss: link.enable_rss,
      skip_health_check: link.skip_health_check,
      is_died: link.is_died ?? false
    })
  } else {
    isEditMode.value = false
  }
  dialogVisible.value = true
}

const resetForm = () => {
  formRef.value?.resetFields()
  Object.assign(form, {
    id: 0,
    name: '',
    link: '',
    avatar: '',
    description: '',
    email: '',
    times: 0,
    status: 'survival',
    enable_rss: true,
    skip_health_check: false,
    is_died: false
  })
}

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (isEditMode.value) {
          const { id, ...data } = form
          await updateFriendLink(id, { data })
          ElMessage.success('更新成功')
        } else {
          const payload = {
            name: form.name,
            link: form.link,
            avatar: form.avatar,
            description: form.description,
            email: form.email,
            enable_rss: form.enable_rss
          }
          await createFriendLink(payload)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        fetchFriendLinks()
      } catch {
        // Request errors are reported by the interceptor.
      }
    }
  })
}

// Delete operations
const handleDelete = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除这个友链吗？', '警告', { type: 'warning' })
    await deleteFriendLink(id)
    ElMessage.success('删除成功')
    fetchFriendLinks()
  } catch {
    // Dismissals are normal; request errors are reported by the interceptor.
  }
}

const handleRecheck = async (id: number) => {
  if (recheckingId.value !== null || inspectionRunning.value) return

  recheckingId.value = id
  try {
    const res = await recheckFriendLink(id)
    if (res.data.success) {
      ElMessage.success('巡查成功')
    } else {
      ElMessage.error('巡查失败')
    }
    await fetchFriendLinks()
  } catch {
  } finally {
    recheckingId.value = null
  }
}

// Full inspection: the backend runs it in the background, so the UI polls.
let progressTimer: ReturnType<typeof setInterval> | null = null

function stopProgressPolling() {
  if (progressTimer !== null) {
    clearInterval(progressTimer)
    progressTimer = null
  }
}

const refreshProgress = async () => {
  try {
    const res = await getRecheckProgress()
    const wasRunning = inspectionRunning.value
    progress.value = res.data
    if (res.data.running) {
      startProgressPolling()
    } else {
      stopProgressPolling()
      if (wasRunning) {
        ElMessage.success('全量巡查完成')
        await fetchFriendLinks()
      }
    }
  } catch {
    // The response interceptor reports request failures.
    stopProgressPolling()
  }
}

function startProgressPolling() {
  if (progressTimer !== null) return
  progressTimer = setInterval(refreshProgress, 2000)
}

const handleRecheckAll = async () => {
  if (inspectionRunning.value) return

  try {
    await ElMessageBox.confirm(
      '全量巡查会重新抽查所有未关闭巡查的友链（含已失效的），可能持续几分钟，确定继续吗？',
      '提示',
      { confirmButtonText: '开始巡查', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }

  try {
    const res = await recheckAllFriendLinks()
    progress.value = res.data
    ElMessage.success('全量巡查已开始')
    startProgressPolling()
  } catch {
    // A 409 means another run holds the crawler; the interceptor reports it.
    await refreshProgress()
  }
}

// 状态合进名称列：文字颜色即状态，失效的整行置灰。
const nameClass = (row: FriendLink) => `link-name--${row.status}`

// 传 link 给这条友链选标签；传 tag 直接进入编辑。
const openTagDrawer = (link: FriendLink | null, tag: FriendTag | null = null) => {
  tagDrawerLink.value = link
  tagDrawerFocus.value = tag
  tagDrawerVisible.value = true
}

const rowClassName = ({ row }: { row: FriendLink }) => (row.is_died ? 'friend-link-died' : '')

const handleHealthCheckToggle = async (link: FriendLink) => {
  const originalValue = link.skip_health_check
  const newValue = !originalValue

  link.skip_health_check = newValue
  try {
    await updateFriendLink(link.id, { data: { skip_health_check: newValue } })
    ElMessage.success(`已${newValue ? '停止' : '恢复'}巡查`)
    fetchFriendLinks()
  } catch {
    link.skip_health_check = originalValue
  }
}

const handleRssToggle = async (link: FriendLink) => {
  const originalValue = link.enable_rss
  const newValue = !originalValue

  // If turning off, show confirmation dialog
  if (!newValue) {
    try {
      await ElMessageBox.confirm(
        '关闭 RSS 订阅将删除所有相关的订阅源和已抓取的文章。此操作不可逆，确定要继续吗？',
        '警告',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch {
      // User canceled, do nothing, the switch state is not yet changed in the UI data
      return
    }
  }

  // Optimistically update the UI
  link.enable_rss = newValue

  // Proceed with the API call
  try {
    await updateFriendLink(link.id, { data: { enable_rss: newValue } })
    ElMessage.success(`已${newValue ? '开启' : '关闭'} RSS 订阅`)
    // On success, fetch the data again to ensure consistency
    fetchFriendLinks()
  } catch {
    // Revert the switch on API failure
    link.enable_rss = originalValue
  }
}
</script>

<style scoped>
.table-scroll {
  flex: 1;
  min-height: 0;
}

.inspection-alert {
  margin-bottom: 16px;
}

.link-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.link-avatar {
  flex: none;
  font-size: 11px;
}

.link-name {
  display: block;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}

.link-name--survival {
  color: var(--el-color-success);
}

.link-name--pending {
  color: var(--el-color-info);
}

.link-name--timeout {
  color: var(--el-color-warning);
}

.link-name--error {
  color: var(--el-color-danger);
}

:deep(.friend-link-died .el-table__cell) {
  background-color: var(--el-fill-color-lighter);
}

:deep(.friend-link-died .el-table__cell:not(:first-child)) {
  color: var(--el-text-color-disabled);
}

:deep(.friend-link-died a),
:deep(.friend-link-died .el-button) {
  color: var(--el-text-color-disabled);
}

:deep(.friend-link-died .link-avatar) {
  filter: grayscale(1);
  opacity: 0.6;
}

/* 失效的友链只在名称上划一杠，颜色仍然表明真实巡查结果。 */
:deep(.friend-link-died .link-name) {
  text-decoration: line-through;
}

.table-actions {
  margin-bottom: 16px;
}

.tag-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tag-item {
  border: none;
  cursor: pointer;
}

.tag-add {
  height: 20px;
  padding: 0;
}
</style>
