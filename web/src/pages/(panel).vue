<template>
  <el-container class="panel-container">
    <el-container class="panel-body">
      <el-aside width="200px" class="panel-aside">
        <el-menu :default-active="route.path" class="sidebar-menu" router>
          <el-menu-item
            v-for="item in menus"
            :key="item.path"
            :index="item.path"
            :class="{ 'sidebar-bottom': item.bottom }"
          >
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.title }}</span>
          </el-menu-item>
        </el-menu>
        <el-dropdown class="sidebar-account" trigger="click" placement="top-start" @command="handleCommand">
          <button type="button" class="user-info">
            <el-icon><User /></el-icon>
            <span>{{ username }}</span>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">
                <el-icon><SwitchButton /></el-icon>
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-aside>

      <el-main class="panel-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { User, SwitchButton } from '@element-plus/icons-vue'

definePage({ redirect: '/dashboard' })

const router = useRouter()
const route = useRoute()
const username = localStorage.getItem('username') || '管理员'

const menus = computed(() =>
  router
    .getRoutes()
    .filter((r) => r.meta.title)
    .sort((a, b) => (a.meta.order ?? 0) - (b.meta.order ?? 0))
    .map((r) => ({ path: r.path, ...r.meta }))
)

const handleCommand = async (command: string) => {
  if (command !== 'logout') return
  await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
  localStorage.removeItem('token')
  localStorage.removeItem('username')
  ElMessage.success('已退出登录')
  router.push('/login')
}
</script>

<style scoped>
.panel-container {
  width: 100%;
  height: 100vh;
}

.sidebar-account {
  flex-shrink: 0;
}

.user-info {
  display: flex;
  align-items: center;
  width: 100%;
  min-height: var(--el-menu-item-height);
  padding: 0 var(--el-menu-base-level-padding);
  border: none;
  background: transparent;
  color: #606266;
  font: inherit;
  text-align: left;
  cursor: pointer;
  transition: background 0.3s;
}

.user-info .el-icon {
  flex-shrink: 0;
  width: var(--el-menu-icon-width);
  margin-right: 5px;
  font-size: 18px;
}

.user-info:hover {
  background: #f5f7fa;
}

.panel-body {
  min-height: 0;
}

.panel-aside {
  display: flex;
  flex-direction: column;
  background: #fff;
  border-right: 1px solid #e4e7ed;
}

.sidebar-menu {
  display: flex;
  flex: 1;
  flex-direction: column;
  border-right: none;
}

.sidebar-bottom {
  margin-top: auto;
  border-top: 1px solid #e4e7ed;
}

.panel-main {
  background: #f0f2f5;
  padding: 6px;
  overflow-y: auto;
}
</style>
