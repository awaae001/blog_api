<template>
  <span class="secret-label">
    <slot />
    <el-tooltip content="敏感信息，会写入配置文件" placement="top">
      <el-icon class="secret-label__lock"><Lock /></el-icon>
    </el-tooltip>
  </span>
</template>

<script setup lang="ts">
import { Lock } from '@element-plus/icons-vue'
</script>

<style scoped>
.secret-label {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.secret-label__lock {
  font-size: 12px;
  color: var(--el-color-warning);
}
</style>
