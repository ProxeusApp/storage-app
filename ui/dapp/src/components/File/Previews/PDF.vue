<template>
  <object class="pdf" v-if="src" :data="src+'#toolbar=1&navpanes=0'" type="application/pdf" id="mypdf"></object>
</template>

<script>
import BasePreview from './BasePreview'

export default {
  name: 'preview-pdf',
  extends: BasePreview,
  mounted () {
    console.log('preview-pdf src = ' + this.src)
  }
}
</script>

<style scoped>
  .pdf {
    width: 100%;
    height: 100%;
    display: block;
    overflow: hidden;
  }
</style>
