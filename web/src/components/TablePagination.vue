<template>
  <el-pagination
    class="table-pagination"
    background
    :layout="layout"
    :total="total"
    :page-sizes="pageSizes"
    :page-size="pageSize"
    :current-page="currentPage"
    @size-change="emit('size-change', $event)"
    @current-change="emit('current-change', $event)"
  />
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  total: number
  pageSize: number
  currentPage: number
  pageSizes?: number[]
  layout?: string
}>(), {
  pageSizes: () => [10, 20, 50, 100],
  layout: 'total, sizes, prev, pager, next, jumper'
})

const emit = defineEmits<{
  'size-change': [size: number]
  'current-change': [page: number]
}>()
</script>

<style scoped>
.table-pagination {
  display: flex;
  flex-shrink: 0;
  justify-content: flex-end;
  margin-top: 20px;
}
</style>
