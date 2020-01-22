<template>
  <div class="upload-thumbnail image-holder p-1 bg-light">
    <img :src="src" width="100%" alt="" v-if="src">
  </div>
</template>

<script>
import BasePreview from './BasePreview'

export default {
  name: 'preview-image',
  extends: BasePreview
}
</script>
