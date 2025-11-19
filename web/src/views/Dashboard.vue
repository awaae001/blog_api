<template>
  <div class="dashboard-container">
    <el-card class="welcome-card">
      <h3>欢迎使用管理面板</h3>
      <p>当前登录用户: <strong>{{ username }}</strong></p>
      <el-divider />
      <div class="info-section">
        <h4>快速开始</h4>
        <ul>
          <li>左侧菜单可以切换不同的管理模块</li>
          <li>友链管理：查看、创建、编辑、删除友链</li>
          <li>RSS文章：查看爬取的RSS文章内容</li>
          <li>系统设置：配置CORS、数据库等参数</li>
        </ul>
      </div>
      <el-divider />
      <div class="stats-section">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-statistic title="友链总数" :value="stats.status_data.friend_link_count" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="RSS文章总数" :value="stats.status_data.rss_post_count" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="在线时长" :value="stats.uptime" />
          </el-col>
        </el-row>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { statsApi, type SystemStatus } from '@/api/stats'

const username = ref('')
const stats = ref<SystemStatus>({
  uptime: '0s',
  status_data: {
    friend_link_count: 0,
    rss_count: 0,
    rss_post_count: 0
  },
  time: ''
})

onMounted(async () => {
  username.value = localStorage.getItem('username') || '管理员'
  try {
    const res = await statsApi.getSystemStatus()
    if (res.code === 200) {
      stats.value = res.data
    } else {
      ElMessage.error(res.message || '获取状态信息失败')
    }
  } catch (error) {
    ElMessage.error('请求状态信息时出错')
  }
})
</script>

<style scoped>
.dashboard-container {
  padding: 20px;
}
.welcome-card {
  max-width: 1200px;
  margin: 0 auto;
}
.welcome-card h3 {
  margin: 0 0 16px 0;
  color: #303133;
  font-size: 24px;
}
.welcome-card p {
  color: #606266;
  margin: 0 0 16px 0;
}
.info-section h4 {
  color: #303133;
  margin: 0 0 12px 0;
}
.info-section ul {
  margin: 0;
  padding-left: 20px;
  color: #606266;
  line-height: 1.8;
}
.stats-section {
  margin-top: 16px;
}
</style>