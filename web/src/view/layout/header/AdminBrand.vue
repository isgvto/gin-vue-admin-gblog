<template>
  <button
    type="button"
    class="admin-brand"
    :class="{ 'admin-brand--compact': compact }"
    :aria-label="`${name}，返回首页`"
    title="返回首页"
    @click="$emit('navigate')"
  >
    <span class="admin-brand__icon" aria-hidden="true"><Logo /></span>
    <span v-if="!compact" class="admin-brand__copy">
      <span class="admin-brand__name">{{ name }}</span>
      <span class="admin-brand__caption">内容管理后台</span>
    </span>
  </button>
</template>

<script setup>
import Logo from '@/components/logo/index.vue'

defineProps({
  name: { type: String, default: 'GBlog' },
  compact: { type: Boolean, default: false }
})
defineEmits(['navigate'])
</script>

<style scoped>
.admin-brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  min-width: 192px;
  height: 48px;
  padding: 6px 10px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--admin-text, var(--el-text-color-primary));
  font-family: inherit;
  text-align: left;
  cursor: pointer;
  flex-shrink: 0;
  transition: background-color .15s ease;
}
.admin-brand:hover { background: var(--admin-selected, var(--el-fill-color-light)); }
.admin-brand:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 2px; }
.admin-brand__icon { display: flex; align-items: center; justify-content: center; width: 32px; height: 32px; flex-shrink: 0; }
.admin-brand__copy { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.admin-brand__name { max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 21px; line-height: 23px; font-weight: 750; letter-spacing: .025em; }
.admin-brand__caption { color: var(--admin-muted, var(--el-text-color-secondary)); font-size: 10px; line-height: 14px; font-weight: 400; letter-spacing: .14em; white-space: nowrap; }
.admin-brand--compact { min-width: 48px; width: 48px; padding: 8px; }
@media (prefers-reduced-motion: reduce) { .admin-brand { transition: none; } }
</style>
