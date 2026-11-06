<template>
  <el-drawer v-model="visible" :title="link ? `「${link.name}」的标签` : '标签管理'" size="400px" @open="init">
    <div class="section-title">{{ editing ? `编辑「${editing.name}」` : '新建标签' }}</div>
    <el-form label-width="48px" class="tag-form">
      <el-form-item label="名称">
        <el-input v-model="form.name" @keyup.enter="submit" />
      </el-form-item>
      <el-form-item label="颜色">
        <div class="swatches">
          <button v-for="color in presetColors" :key="color" type="button" class="swatch"
            :class="{ 'is-active': form.color.toLowerCase() === color.toLowerCase() }" :style="{ background: color }"
            :title="color" @click="form.color = color" />
          <button type="button" class="swatch swatch--none" :class="{ 'is-active': !form.color }" title="默认"
            @click="form.color = ''" />
          <el-color-picker :model-value="form.color" :predefine="presetColors" size="small"
            @update:model-value="form.color = $event ?? ''" />
        </div>
      </el-form-item>
      <el-form-item label="排序">
        <el-input-number v-model="form.sort_order" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="saving" :disabled="!form.name.trim()" @click="submit">
          {{ editing ? '保存' : link ? '新建并选中' : '新建' }}
        </el-button>
        <template v-if="editing">
          <el-button type="danger" plain :loading="saving" @click="removeTag">删除</el-button>
          <el-button @click="startEdit(null)">取消</el-button>
        </template>
      </el-form-item>
    </el-form>

    <div class="section-title">{{ link ? '点选标签' : '全部标签，点击编辑' }}</div>
    <div class="tag-list">
      <template v-if="link">
        <el-check-tag v-for="tag in tags" :key="tag.id" :checked="selected.has(tag.id)" @change="toggle(tag.id)">
          {{ tag.name }}
        </el-check-tag>
      </template>
      <template v-else>
        <el-tag v-for="tag in tags" :key="tag.id" :color="tag.color || undefined"
          :effect="editing?.id === tag.id ? 'dark' : 'light'" class="tag-item" @click="startEdit(tag)">
          {{ tag.name }}<span class="tag-count">{{ tag.link_count ?? 0 }}</span>
        </el-tag>
      </template>
      <span v-if="!tags.length" class="hint">还没有标签</span>
    </div>

    <template v-if="link" #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="saveLinkTags">保存标签</el-button>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createFriendTag,
  deleteFriendTag,
  replaceFriendLinkTags,
  updateFriendTag,
  type FriendLink,
  type FriendTag,
  type FriendTagPayload
} from '@/api/friend-links'

const props = defineProps<{ link: FriendLink | null; focus?: FriendTag | null; tags: FriendTag[] }>()
const emit = defineEmits<{ changed: [] }>()
const visible = defineModel<boolean>({ required: true })

const presetColors = ['#fde2e2', '#fdebd3', '#faf3c0', '#e1f3d8', '#d5f2ef', '#d9ecff', '#e8e0fa', '#f9e0f0', '#ebeef5']

const selected = ref(new Set<number>())
const editing = ref<FriendTag | null>(null)
const saving = ref(false)
const form = reactive<FriendTagPayload>({ name: '', color: '', sort_order: 0 })

const startEdit = (tag: FriendTag | null) => {
  editing.value = tag
  Object.assign(form, { name: tag?.name ?? '', color: tag?.color ?? '', sort_order: tag?.sort_order ?? 0 })
}

const init = () => {
  selected.value = new Set(props.link?.tags.map((t) => t.id))
  const focus = props.focus && (props.tags.find((t) => t.id === props.focus!.id) ?? props.focus)
  startEdit(focus ?? null)
}

const toggle = (id: number) => {
  if (!selected.value.delete(id)) selected.value.add(id)
}

const run = async (task: () => Promise<void>) => {
  saving.value = true
  try {
    await task()
    emit('changed')
  } catch {
    // 拦截器已经提示过了
  } finally {
    saving.value = false
  }
}

const submit = () =>
  run(async () => {
    const payload = { ...form, name: form.name.trim() }
    if (editing.value) {
      await updateFriendTag(editing.value.id, payload)
      ElMessage.success('标签已更新')
    } else {
      const { data: tag } = await createFriendTag(payload)
      if (props.link) selected.value.add(tag.id)
      ElMessage.success('标签已新建')
    }
    startEdit(null)
  })

const removeTag = async () => {
  const tag = editing.value!
  try {
    await ElMessageBox.confirm(`删除「${tag.name}」会在所有友链上移除这个 tag，确定吗？`, '警告', { type: 'warning' })
  } catch {
    return
  }
  await run(async () => {
    await deleteFriendTag(tag.id)
    ElMessage.success('标签已删除')
    startEdit(null)
  })
}

const saveLinkTags = () =>
  run(async () => {
    await replaceFriendLinkTags(props.link!.id, [...selected.value])
    ElMessage.success('标签已保存')
    visible.value = false
  })
</script>

<style scoped>
.section-title {
  margin-bottom: 8px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.tag-form {
  margin-bottom: 8px;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag-item {
  border: none;
  cursor: pointer;
}

.tag-count {
  margin-left: 4px;
  opacity: 0.6;
}

.swatches {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.swatch {
  width: 20px;
  height: 20px;
  padding: 0;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  cursor: pointer;
}

.swatch.is-active {
  outline: 2px solid var(--el-color-primary);
  outline-offset: 1px;
}

.swatch--none {
  background: linear-gradient(135deg, transparent 45%, var(--el-color-danger) 45% 55%, transparent 55%);
}

.hint {
  color: var(--el-text-color-placeholder);
  font-size: 13px;
}
</style>
