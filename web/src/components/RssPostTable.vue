<template>
  <el-table :data="posts" v-loading="loading" :max-height="maxHeight" style="width: 100%">
    <el-table-column prop="title" label="文章标题">
      <template #default="{ row }">
        <a :href="row.link" target="_blank" rel="noopener noreferrer" class="post-link">{{ row.title }}</a>
      </template>
    </el-table-column>
    <el-table-column v-if="showAuthor" prop="author" label="作者" width="160" />
    <el-table-column prop="time" label="发布时间" width="180">
      <template #default="{ row }">
        {{ formatDate(row.time) }}
      </template>
    </el-table-column>
    <!-- @vue-generic {RssPost} -->
    <el-table-column label="操作" width="80">
      <template #default="{ row }">
        <el-button type="danger" link :icon="Delete" @click="emit('delete', row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
import { Delete } from '@element-plus/icons-vue'
import type { RssPost } from '@/api/rss'
import { formatDate } from '@/utils/date'

// 每个折叠项仍使用页面维护的文章状态，只保留一份分页和请求逻辑。
defineProps<{
  posts: RssPost[]
  loading: boolean
  showAuthor?: boolean
  maxHeight?: number
}>()

const emit = defineEmits<{
  delete: [post: RssPost]
}>()
</script>

<style scoped>
.post-link {
  text-decoration: none;
  color: #409eff;
}

.post-link:hover {
  text-decoration: underline;
}
</style>
