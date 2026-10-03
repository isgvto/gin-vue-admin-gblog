<template>
  <header class="admin-page-heading">
    <div class="admin-page-heading__copy">
      <h1>{{ title }}</h1>
      <p v-if="description">{{ description }}</p>
    </div>
    <div v-if="$slots.default" class="admin-page-heading__actions"><slot /></div>
  </header>
</template>
<script setup>
  defineProps({ title: { type: String, required: true }, description: { type: String, default: '' } })
</script>
