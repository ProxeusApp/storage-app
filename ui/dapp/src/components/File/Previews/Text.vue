<template>
  <object v-if="src" :data="src+'#toolbar=0&navpanes=0'"></object>
</template>

<script>
import BasePreview from './BasePreview'

export default {
  name: 'preview-text',
  extends: BasePreview
}
</script>

<style scoped>
  .preview-image {
    max-width: 100%;
  }
</style>
