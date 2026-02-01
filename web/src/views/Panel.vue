<template>
  <el-container class="panel-container">
    <el-header class="panel-header">
      <div class="header-left">
        <h2>Blog API 管理面板</h2>
      </div>
      <div class="header-right">
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-icon><User /></el-icon>
            <span>{{ username }}</span>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">
                <el-icon><SwitchButton /></el-icon>
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <el-container>
      <el-aside width="200px" class="panel-aside">
        <el-menu
          :default-active="activeMenu"
          class="sidebar-menu"
          @select="handleMenuSelect"
        >
          <el-menu-item index="dashboard">
            <el-icon><HomeFilled /></el-icon>
            <span>仪表板</span>
          </el-menu-item>
          <el-menu-item index="moments">
            <el-icon><ChatLineRound /></el-icon>
            <span>我的动态</span>
          </el-menu-item>
          <el-menu-item index="friend">
            <el-icon><Link /></el-icon>
            <span>友链管理</span>
          </el-menu-item>
          <el-menu-item index="rss">
            <el-icon><Document /></el-icon>
            <span>RSS 管理</span>
          </el-menu-item>
          <el-menu-item index="image">
            <el-icon><Picture /></el-icon>
            <span>图片管理</span>
          </el-menu-item>
          <el-menu-item index="resource">
            <el-icon><FolderOpened /></el-icon>
            <span>本地资源</span>
          </el-menu-item>
          <el-menu-item index="settings">
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <el-main class="panel-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import {
  User,
  SwitchButton,
  HomeFilled,
  ChatLineRound,
  Link,
  Document,
  Setting,
  Picture,
  FolderOpened
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const username = ref('')

const activeMenu = computed(() => {
  return route.path.substring(1) // e.g., /friend-link -> friend-link
})

onMounted(() => {
  username.value = localStorage.getItem('username') || '管理员'
})

const handleMenuSelect = (index: string) => {
  router.push(`/${index}`)
}

const handleCommand = (command: string) => {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      ElMessage.success('已退出登录')
      router.push('/login')
    })
  }
}
</script>

<style scoped>
.panel-container {
  width: 100%;
  height: 100vh;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 20px;
}

.header-left h2 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 4px;
  transition: background 0.3s;
}

.user-info:hover {
  background: #f5f7fa;
}

.panel-aside {
  background: #fff;
  border-right: 1px solid #e4e7ed;
}

.sidebar-menu {
  border-right: none;
}

.panel-main {
  background: #f0f2f5;
  padding: 6px;
  overflow-y: auto;
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
