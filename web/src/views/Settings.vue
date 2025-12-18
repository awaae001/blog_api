<template>
  <div class="settings-container">
    <el-card class="settings-card">
      <template #header>
        <div class="card-header">
          <span>系统设置</span>
          <el-button type="primary" @click="saveConfig" :loading="saving">保存配置</el-button>
        </div>
      </template>

      <el-tabs v-model="activeTab">
        <!-- 安全配置 -->
        <el-tab-pane label="安全配置" name="safe">
          <el-form :model="config" label-width="150px">
            <el-form-item label="CORS 白名单">
              <el-tag
                v-for="(host, index) in config.system_conf.safe_conf.cors_allow_hostlist"
                :key="index"
                closable
                @close="removeArrayItem('cors_allow_hostlist', index)"
                style="margin-right: 8px; margin-bottom: 8px"
              >
                {{ host }}
              </el-tag>
              <el-input
                v-model="newCorsHost"
                placeholder="输入域名后按回车添加"
                @keyup.enter="addCorsHost"
                style="width: 300px"
              />
            </el-form-item>

            <el-form-item label="排除路径">
              <el-tag
                v-for="(path, index) in config.system_conf.safe_conf.exclude_paths"
                :key="index"
                closable
                @close="removeArrayItem('exclude_paths', index)"
                style="margin-right: 8px; margin-bottom: 8px"
              >
                {{ path }}
              </el-tag>
              <el-input
                v-model="newExcludePath"
                placeholder="输入路径后按回车添加"
                @keyup.enter="addExcludePath"
                style="width: 300px"
              />
            </el-form-item>

            <el-form-item label="允许的扩展名">
              <el-tag
                v-for="(ext, index) in config.system_conf.safe_conf.allow_extension"
                :key="index"
                closable
                @close="removeArrayItem('allow_extension', index)"
                style="margin-right: 8px; margin-bottom: 8px"
              >
                {{ ext }}
              </el-tag>
              <el-input
                v-model="newAllowExtension"
                placeholder="输入扩展名后按回车添加"
                @keyup.enter="addAllowExtension"
                style="width: 300px"
              />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 数据配置 -->
        <el-tab-pane label="数据配置" name="data">
          <el-form :model="config" label-width="150px">
            <el-divider content-position="left">数据库配置</el-divider>
            <el-form-item label="数据库路径">
              <el-input
                v-model="config.system_conf.data_conf.database.path"
                placeholder="例如: data/blog.db"
              />
            </el-form-item>

            <el-divider content-position="left">图片配置</el-divider>
            <el-form-item label="图片存储路径">
              <el-input
                v-model="config.system_conf.data_conf.image.path"
                placeholder="例如: data/images"
              />
            </el-form-item>
            <el-form-item label="图片转换格式">
              <el-select v-model="config.system_conf.data_conf.image.conv_to">
                <el-option label="webp" value="webp" />
                <el-option label="jpeg" value="jpeg" />
                <el-option label="png" value="png" />
                <el-option label="不转换" value="" />
              </el-select>
            </el-form-item>

            <el-divider content-position="left">资源配置</el-divider>
            <el-form-item label="资源存储路径">
              <el-input
                v-model="config.system_conf.data_conf.resource.path"
                placeholder="例如: data/resources"
              />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 爬虫配置 -->
        <el-tab-pane label="爬虫配置" name="crawler">
          <el-form :model="config" label-width="150px">
            <el-form-item label="并发数量">
              <el-input-number
                v-model="config.system_conf.crawler_conf.concurrency"
                :min="1"
                :max="20"
              />
              <div style="color: #909399; font-size: 12px; margin-top: 8px">
                设置 RSS 爬虫的并发数量，建议值为 5-10
              </div>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSystemConfig, updateSystemConfig } from '@/api/config'
import type { SystemConfig } from '@/model/config'

const activeTab = ref('safe')
const saving = ref(false)
const newCorsHost = ref('')
const newExcludePath = ref('')
const newAllowExtension = ref('')

const config = ref<SystemConfig>({
  system_conf: {
    safe_conf: {
      cors_allow_hostlist: [],
      exclude_paths: [],
      allow_extension: []
    },
    data_conf: {
      database: {
        path: ''
      },
      image: {
        path: '',
        conv_to: ''
      },
      resource: {
        path: ''
      }
    },
    crawler_conf: {
      concurrency: 5
    }
  }
})

onMounted(async () => {
  try {
    const res = await getSystemConfig()
    config.value = res
  } catch (error) {
    ElMessage.error('请求配置时出错')
    console.error(error)
  }
})

const addCorsHost = () => {
  if (newCorsHost.value.trim()) {
    config.value.system_conf.safe_conf.cors_allow_hostlist.push(newCorsHost.value.trim())
    newCorsHost.value = ''
  }
}

const addExcludePath = () => {
  if (newExcludePath.value.trim()) {
    config.value.system_conf.safe_conf.exclude_paths.push(newExcludePath.value.trim())
    newExcludePath.value = ''
  }
}

const addAllowExtension = () => {
  if (newAllowExtension.value.trim()) {
    config.value.system_conf.safe_conf.allow_extension.push(newAllowExtension.value.trim())
    newAllowExtension.value = ''
  }
}

const removeArrayItem = (field: string, index: number) => {
  const safeConf = config.value.system_conf.safe_conf as any
  safeConf[field].splice(index, 1)
}

const saveConfig = async () => {
  saving.value = true
  try {
    // 保存所有配置项
    const updates = [
      { key: 'system_conf.safe_conf.cors_allow_hostlist', value: config.value.system_conf.safe_conf.cors_allow_hostlist },
      { key: 'system_conf.safe_conf.exclude_paths', value: config.value.system_conf.safe_conf.exclude_paths },
      { key: 'system_conf.safe_conf.allow_extension', value: config.value.system_conf.safe_conf.allow_extension },
      { key: 'system_conf.data_conf.database.path', value: config.value.system_conf.data_conf.database.path },
      { key: 'system_conf.data_conf.image.path', value: config.value.system_conf.data_conf.image.path },
      { key: 'system_conf.data_conf.image.conv_to', value: config.value.system_conf.data_conf.image.conv_to },
      { key: 'system_conf.data_conf.resource.path', value: config.value.system_conf.data_conf.resource.path },
      { key: 'system_conf.crawler_conf.concurrency', value: config.value.system_conf.crawler_conf.concurrency }
    ]

    for (const update of updates) {
      await updateSystemConfig(update.key, update.value)
    }

    ElMessage.success('配置保存成功')
  } catch (error) {
    ElMessage.error('保存配置失败')
    console.error(error)
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.settings-container {
  padding: 20px;
}

.settings-card {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}

:deep(.el-divider__text) {
  font-weight: 600;
  color: #303133;
}
</style>