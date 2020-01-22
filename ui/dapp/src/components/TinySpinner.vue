<template>
  <div class="spinner m-1 text-center d-inline-block">
    <div class="tinyspinner"></div>
  </div>
</template>

<script>
export default {
  name: 'tiny-spinner'
}
</script>

<style lang="scss" scoped>
  @import "../assets/styles/variables";

  .spinner {
    display: block;
    width: 1em;
    height: 1em;
  }

  .tinyspinner {
    width: 100%;
    height: 100%;
    border: 0.15em solid $gray-500;
    border-bottom: 0.15em solid rgba(0, 0, 0, 0);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    z-index: 9999;
  }

  .tinyspinner--hidden {
    display: none;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  @keyframes rotating {
    from {
      transform: rotate(0deg);
    }

    to {
      transform: rotate(360deg);
    }
  }
</style>
