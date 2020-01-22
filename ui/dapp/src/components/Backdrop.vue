<template>
  <div class="backdrop" @click="clicked" :class="{visible}" :style="{'z-index': zindex}"></div>
</template>

<script>
export default {
  name: 'backdrop',
  props: {
    visible: {
      default: false
    },
    zindex: {
      default: 1030
    }
  },
  methods: {
    clicked () {
      this.$emit('clicked')
    }
  }
}
</script>

<style lang="scss" scoped>
  .backdrop {
    opacity: 0;
    /*transition: all 150ms;*/

    &.visible {
      opacity: 1;
      display: block;
      position: fixed;
      left: 0;
      top: 0;
      width: 100%;
      height: 100%;
      background: rgba(0, 0, 0, 0.1);
      /*transition: all 150ms;*/
    }
  }
</style>
