<template>
  <div class="page-shell">
    <header class="page-shell__header">
      <div class="page-shell__heading">
        <h2 class="page-shell__title">
          <slot name="title">{{ title }}</slot>
        </h2>
        <slot name="subtitle" />
      </div>
      <div v-if="$slots.actions" class="page-shell__actions">
        <slot name="actions" />
      </div>
    </header>
    <div class="page-shell__body" :class="{ 'is-flush': flush }">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  /** 内容区不留内边距，表格 / tabs 直接贴边 */
  flush?: boolean
}>()
</script>

<style scoped>
.page-shell {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.page-shell__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
  /* 固定高度，各页面 header 保持一致 */
  height: 64px;
  padding: 0 20px;
  background: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-light);
}

.page-shell__heading {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
  overflow: hidden;
}

.page-shell__title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  white-space: nowrap;
}

.page-shell__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.page-shell__body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 16px 20px;
  overflow-y: auto;
  background: var(--el-bg-color-page);
}

/* Element 把这些变量直接声明在组件根上，得命中组件本身才能覆盖 */
.page-shell__body :deep(.el-table) {
  --el-table-bg-color: var(--el-bg-color-page);
  --el-table-tr-bg-color: var(--el-bg-color-page);
  --el-table-header-bg-color: var(--el-bg-color-page);
  --el-table-expanded-cell-bg-color: var(--el-bg-color-page);
  --el-table-row-hover-bg-color: var(--el-fill-color-dark);
  --el-table-border-color: var(--el-border-color);
}

.page-shell__body :deep(.el-collapse) {
  --el-collapse-header-bg-color: var(--el-bg-color-page);
  --el-collapse-content-bg-color: var(--el-bg-color-page);
  --el-collapse-border-color: var(--el-border-color);
}

.page-shell__body.is-flush {
  padding: 0;
}
</style>
