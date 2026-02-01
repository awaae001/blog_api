<template>
  <div class="resource-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span class="title">本地资源管理</span>
            <el-breadcrumb separator="/">
              <el-breadcrumb-item
                v-for="crumb in breadcrumbs"
                :key="crumb.path"
                class="breadcrumb-item"
              >
                <el-link
                  :underline="false"
                  @click="handleBreadcrumbClick(crumb.path)"
                >
                  {{ crumb.label }}
                </el-link>
              </el-breadcrumb-item>
            </el-breadcrumb>
          </div>
          <div class="header-actions">
            <el-button :disabled="!currentPath" @click="goToParent">
              返回上级
            </el-button>
            <el-button type="primary" :icon="Upload" @click="openUploadDialog">
              上传文件
            </el-button>
            <el-button :icon="Refresh" @click="refreshList">
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="sortedEntries"
        v-loading="loading"
        style="width: 100%"
      >
        <el-table-column label="名称" min-width="240">
          <template #default="{ row }">
            <div class="name-cell">
              <el-icon class="name-icon">
                <Folder v-if="row.is_dir" />
                <Document v-else />
              </el-icon>
              <el-link
                v-if="row.is_dir"
                :underline="false"
                @click="handleOpen(row)"
              >
                {{ row.name }}
              </el-link>
              <span v-else>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="260" show-overflow-tooltip />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_dir ? 'info' : 'success'">
              {{ row.is_dir ? '文件夹' : '文件' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="大小" width="120">
          <template #default="{ row }">
            {{ row.is_dir ? '-' : formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column label="修改时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.mod_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.is_dir"
              type="primary"
              link
              :icon="FolderOpened"
              @click="handleOpen(row)"
            >
              打开
            </el-button>
            <el-button
              v-else
              type="primary"
              link
              :icon="Download"
              @click="handleDownload(row)"
            >
              下载
            </el-button>
            <el-button
              type="danger"
              link
              :icon="Delete"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="uploadDialogVisible" title="上传文件" width="520px" @close="resetUploadForm">
      <el-form label-width="90px">
        <el-form-item label="目标目录">
          <el-input v-model="uploadPath" placeholder="相对 data 目录路径（可为空）" />
        </el-form-item>
        <el-form-item label="覆盖同名">
          <el-switch v-model="overwrite" />
        </el-form-item>
        <el-form-item label="文件">
          <el-upload
            v-model:file-list="uploadFiles"
            drag
            multiple
            :auto-upload="false"
          >
            <el-icon class="upload-icon"><UploadFilled /></el-icon>
            <div class="el-upload__text">拖拽文件到此处，或点击选择</div>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">
          上传
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  UploadFilled,
  Upload,
  Refresh,
  Folder,
  FolderOpened,
  Document,
  Delete,
  Download
} from '@element-plus/icons-vue'
import type { UploadFile } from 'element-plus'
import { listResources, uploadFile, deleteFile } from '@/api/resource'
import type { ResourceEntry } from '@/model/resource'
import { formatDate } from '@/utils/date'

const entries = ref<ResourceEntry[]>([])
const loading = ref(false)
const currentPath = ref('')
const uploadDialogVisible = ref(false)
const uploadFiles = ref<UploadFile[]>([])
const uploadPath = ref('')
const overwrite = ref(false)
const uploading = ref(false)

const normalizePath = (path: string) => path.replace(/^\/+/, '').replace(/\\/g, '/')
const encodePath = (path: string) =>
  encodeURI(normalizePath(path)).replace(/[?#]/g, (match) => encodeURIComponent(match))

const breadcrumbs = computed(() => {
  const parts = currentPath.value ? currentPath.value.split('/').filter(Boolean) : []
  const crumbs = [{ label: '根目录', path: '' as string }]
  let acc = ''
  for (const part of parts) {
    acc = acc ? `${acc}/${part}` : part
    crumbs.push({ label: part, path: acc })
  }
  return crumbs
})

const sortedEntries = computed(() => {
  return [...entries.value].sort((a, b) => {
    if (a.is_dir !== b.is_dir) return a.is_dir ? -1 : 1
    return a.name.localeCompare(b.name)
  })
})

const fetchResources = async (path = currentPath.value) => {
  loading.value = true
  try {
    const res = await listResources(path)
    if (Array.isArray(res.data)) {
      entries.value = res.data.map((item) => ({
        ...item,
        path: normalizePath(item.path)
      }))
    } else {
      entries.value = []
    }
  } catch (error) {
    ElMessage.error('获取资源列表失败')
  } finally {
    loading.value = false
  }
}

const refreshList = () => {
  fetchResources()
}

const handleOpen = (entry: ResourceEntry) => {
  if (!entry.is_dir) return
  currentPath.value = normalizePath(entry.path)
  fetchResources(currentPath.value)
}

const goToParent = () => {
  if (!currentPath.value) return
  const parts = currentPath.value.split('/').filter(Boolean)
  parts.pop()
  currentPath.value = parts.join('/')
  fetchResources(currentPath.value)
}

const handleBreadcrumbClick = (path: string) => {
  currentPath.value = path
  fetchResources(currentPath.value)
}

const handleDownload = (entry: ResourceEntry) => {
  const url = `/api/action/resource/${encodePath(entry.path)}`
  window.open(url, '_blank')
}

const openUploadDialog = () => {
  uploadPath.value = currentPath.value
  uploadDialogVisible.value = true
}

const resetUploadForm = () => {
  uploadFiles.value = []
  uploadPath.value = currentPath.value
  overwrite.value = false
}

const handleUpload = async () => {
  if (!uploadFiles.value.length) {
    ElMessage.error('请选择要上传的文件')
    return
  }
  uploading.value = true
  const targetPath = normalizePath(uploadPath.value)
  try {
    for (const file of uploadFiles.value) {
      if (!file.raw) continue
      const formData = new FormData()
      formData.append('file', file.raw)
      formData.append('path', targetPath)
      formData.append('overwrite', overwrite.value ? 'true' : 'false')
      await uploadFile(formData, 'local')
    }
    ElMessage.success('上传成功')
    uploadDialogVisible.value = false
    fetchResources(currentPath.value)
  } catch (error) {
    ElMessage.error('上传失败')
  } finally {
    uploading.value = false
  }
}

const handleDelete = (entry: ResourceEntry) => {
  ElMessageBox.confirm(`确定要删除 ${entry.name} 吗？`, '提示', {
    type: 'warning'
  }).then(async () => {
    try {
      await deleteFile(normalizePath(entry.path), 'local')
      ElMessage.success('删除成功')
      fetchResources(currentPath.value)
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

const formatSize = (size: number) => {
  if (size <= 0) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let index = 0
  let value = size
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024
    index++
  }
  return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`
}

onMounted(() => {
  fetchResources()
})
</script>

<style scoped>
.resource-management {
  padding: 6px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}
.header-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.title {
  font-size: 16px;
  font-weight: 600;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.name-icon {
  color: #409eff;
}
.upload-icon {
  font-size: 28px;
  color: #8c939d;
}
</style>
