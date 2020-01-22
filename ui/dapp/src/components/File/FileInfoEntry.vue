<template>
  <div class="form-group" v-bind="$attrs">
    <label v-if="label">{{ label }}</label>
    <div class="info shadow-none py-1 px-2 rounded" :class="(infoClass) ? infoClass : 'bg-light'">
      <slot></slot>
    </div>
    <small class="text-muted" v-if="help">{{ help }}</small>
  </div>
</template>

<script>
export default {
  name: 'file-info-entry',
  props: ['label', 'help', 'infoClass']
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .form-group {
    margin: 0.7rem 0;

    label {
      font-size: 0.9rem;
      font-weight: 500;
    }

    .info {
      // font-size: $font-size-sm;
      background: $light;
      border-radius: $border-radius;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: normal;
      word-wrap: break-word;
    }
  }
</style>
