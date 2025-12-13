<template>
  <div class="friend-link-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>友链管理</span>
          <el-button type="primary" :icon="Plus" @click="openFormDialog()">
            新增友链
          </el-button>
        </div>
      </template>

      <!-- Filter and Actions -->
      <div class="table-actions">
        <el-select
          v-model="filterStatus"
          placeholder="按状态筛选"
          clearable
          @change="handleFilter"
          style="width: 150px; margin-right: 10px"
        >
          <el-option label="正常" value="survival"></el-option>
          <el-option label="待定" value="pending"></el-option>
          <el-option label="超时" value="timeout"></el-option>
          <el-option label="错误" value="error"></el-option>
          <el-option label="失效" value="died"></el-option>
          <el-option label="忽略" value="ignored"></el-option>
        </el-select>
        <el-input
          v-model="searchQuery"
          placeholder="搜索友链"
          clearable
          @input="handleSearch"
          style="width: 200px; margin-right: 10px"
        />
        <el-button
          type="danger"
          :icon="Delete"
          @click="handleBulkDelete"
          :disabled="selectedLinks.length === 0"
        >
          批量删除
        </el-button>
      </div>

      <!-- Friend Link Table -->
      <el-table
        :data="friendLinks"
        v-loading="loading"
        @selection-change="handleSelectionChange"
        style="width: 100%"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="website_name" label="网站名称" width="180" />
        <el-table-column prop="website_url" label="链接">
          <template #default="{ row }">
            <a :href="row.website_url" target="_blank">{{ row.website_url }}</a>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="times" label="失败次数" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180" />
        <el-table-column label="订阅 RSS" width="100">
          <template #default="{ row }">
            <el-switch :model-value="row.enable_rss" @change="handleRssToggle(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Edit" @click="openFormDialog(row)">
              编辑
            </el-button>
            <el-button type="danger" link :icon="Delete" @click="handleDelete(row.id)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <el-pagination
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="totalLinks"
        :page-sizes="[10, 20, 50, 100]"
        :page-size="pageSize"
        :current-page="currentPage"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        class="pagination-container"
      />
    </el-card>

    <!-- Form Dialog for Add/Edit -->
    <el-dialog
      :title="isEditMode ? '编辑友链' : '新增友链'"
      v-model="dialogVisible"
      width="500px"
      @close="resetForm"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="网站名称" prop="website_name">
          <el-input v-model="form.website_name" />
        </el-form-item>
        <el-form-item label="网站链接" prop="website_url">
          <el-input v-model="form.website_url" />
        </el-form-item>
        <el-form-item label="网站图标" prop="website_icon_url">
          <el-input v-model="form.website_icon_url" />
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
        <el-form-item label="失败次数" prop="times" v-if="isEditMode">
          <el-input-number v-model="form.times" :min="0" />
        </el-form-item>
        <el-form-item label="状态" prop="status" v-if="isEditMode">
          <el-select v-model="form.status">
            <el-option label="正常" value="survival"></el-option>
            <el-option label="待定" value="pending"></el-option>
            <el-option label="超时" value="timeout"></el-option>
            <el-option label="错误" value="error"></el-option>
            <el-option label="失效" value="died"></el-option>
            <el-option label="忽略" value="ignored"></el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getFriendLinks,
  createFriendLink,
  updateFriendLink,
  deleteFriendLink,
  type FriendLink
} from '@/api/friendLink'

// Reactive State
const friendLinks = ref<FriendLink[]>([])
const selectedLinks = ref<FriendLink[]>([])
const loading = ref(false)
const totalLinks = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const filterStatus = ref('')
const searchQuery = ref('')
const dialogVisible = ref(false)
const isEditMode = ref(false)
const formRef = ref<FormInstance>()
const form = reactive<{
  id: number
  website_name: string
  website_url: string
  website_icon_url: string
  description: string
  email: string
  times: number
  status: 'survival' | 'timeout' | 'error' | 'died' | 'pending' | 'ignored'
  enable_rss: boolean
}>({
  id: 0,
  website_name: '',
  website_url: '',
  website_icon_url: '',
  description: '',
  email: '',
  times: 0,
  status: 'pending',
  enable_rss: true
})

const rules = reactive<FormRules>({
  website_name: [{ required: true, message: '请输入网站名称', trigger: 'blur' }],
  website_url: [{ required: true, message: '请输入网站链接', trigger: 'blur' }]
})

// Fetch data
const fetchFriendLinks = async () => {
  loading.value = true
  try {
    const res = await getFriendLinks({
      page: currentPage.value,
      page_size: pageSize.value,
      status: filterStatus.value,
      search: searchQuery.value
    })
    if (res.code === 200) {
      friendLinks.value = res.data.items
      totalLinks.value = res.data.total
    } else {
      ElMessage.error(res.message || '获取友链列表失败')
    }
  } catch (error) {
    ElMessage.error('请求友链列表时出错')
  } finally {
    loading.value = false
  }
}

onMounted(fetchFriendLinks)

// Table and Pagination
const handleSelectionChange = (selection: FriendLink[]) => {
  selectedLinks.value = selection
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchFriendLinks()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1 // Reset to the first page
  fetchFriendLinks()
}

const handleFilter = () => {
  currentPage.value = 1
  fetchFriendLinks()
}

const handleSearch = () => {
  currentPage.value = 1
  fetchFriendLinks()
}

// Dialog and Form
const openFormDialog = (link?: FriendLink) => {
  if (link) {
    isEditMode.value = true
    Object.assign(form, link)
  } else {
    isEditMode.value = false
  }
  dialogVisible.value = true
}

const resetForm = () => {
  formRef.value?.resetFields()
  Object.assign(form, {
    id: 0,
    website_name: '',
    website_url: '',
    website_icon_url: '',
    description: '',
    email: '',
    times: 0,
    status: 'pending',
    enable_rss: true
  })
}

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (isEditMode.value) {
          const { id, ...data } = form
          await updateFriendLink({ id, data })
          ElMessage.success('更新成功')
        } else {
          const { id, status, ...payload } = form
          await createFriendLink(payload)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        fetchFriendLinks()
      } catch (error) {
        ElMessage.error(isEditMode.value ? '更新失败' : '创建失败')
      }
    }
  })
}

// Delete operations
const handleDelete = (id: number) => {
  ElMessageBox.confirm('确定要删除这个友链吗？', '警告', {
    type: 'warning'
  }).then(async () => {
    try {
      await deleteFriendLink({ ids: [id] })
      ElMessage.success('删除成功')
      fetchFriendLinks()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

const handleBulkDelete = () => {
  ElMessageBox.confirm('确定要删除选中的友链吗？', '警告', {
    type: 'warning'
  }).then(async () => {
    try {
      const ids = selectedLinks.value.map((link) => link.id)
      await deleteFriendLink({ ids })
      ElMessage.success('批量删除成功')
      fetchFriendLinks()
    } catch (error) {
      ElMessage.error('批量删除失败')
    }
  })
}

// UI Helpers
const statusTagType = (status: string) => {
  switch (status) {
    case 'survival':
      return 'success'
    case 'pending':
    case 'ignored':
      return 'info'
    case 'timeout':
      return 'warning'
    case 'error':
    case 'died':
      return 'danger'
    default:
      return 'info'
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
    await updateFriendLink({
      id: link.id,
      data: { enable_rss: newValue }
    })
    ElMessage.success(`已${newValue ? '开启' : '关闭'} RSS 订阅`)
    // On success, fetch the data again to ensure consistency
    fetchFriendLinks()
  } catch (error) {
    ElMessage.error('更新 RSS 订阅状态失败')
    // Revert the switch on API failure
    link.enable_rss = originalValue
  }
}
</script>

<style scoped>
.friend-link-container {
  padding: 20px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.table-actions {
  margin-bottom: 16px;
}
.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>