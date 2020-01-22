<template>
  <div class="file-drag-overlay" :class="{dragging: isFileDragging}"></div>
</template>

<script>
export default {
  name: 'file-drag-overlay',
  props: ['isFileDragging']
}
</script>

<style scoped lang="scss">
  @import "../../assets/styles/variables";

  $drag-indicator-size: 4px;

  .file-drag-overlay {
    position: fixed;
    z-index: 10000000000;
    display: none;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    border: $drag-indicator-size dashed $gray-400;

    &.dragging {
      display: block;
    }
  }
</style>
