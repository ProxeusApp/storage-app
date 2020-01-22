<template>
  <div class="thumbnail" v-bind="$attrs">
    <div v-if="hasThumbnail === true" class="thumbnail-image">
      <img :src="thumbnailSrc" :alt="file.filename">
    </div>
    <div v-if="hasThumbnail === false" class="thumbnail-icon">
      <span class="mdi" :class="defaultThumbnailClass"></span>
    </div>
    <div v-if="getFileKind === 2" class="thumbnail-icon">
      <span class="mdi mdi-file-chart"></span>
    </div>
  </div>
</template>

<script>
export default {
  name: 'file-thumbnail',
  props: {
    file: {
      required: true
    }
  },
  computed: {
    hasThumbnail () {
      return this.file.hasThumbnail === true
    },
    getFileKind () {
      return this.file.fileKind
    },
    thumbnailSrc () {
      return this.$store.getters.thumbnailSrc(this.file.id)
    },
    defaultThumbnailClass () {
      return this.$store.getters.defaultThumbnailClass(this.file)
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .thumbnail {
    overflow: hidden;
    position: relative;
    width: 100%;
    height: 0;
    padding-bottom: 100%;
    background: $gray-300;
    text-align: center;

    .thumbnail-image {
      position: absolute;
      top: 0;
      right: 0;
      bottom: 0;
      left: 0;
      border: 5px solid $gray-300;

      img {
        object-fit: cover;
        width: 100%;
        height: 100%;
      }
    }

    .thumbnail-icon {
      position: absolute;
      left: 0;
      top: 50%;
      transform: translateY(-50%);
      width: 100%;
      color: $primary;
      font-size: 15rem;
    }
  }
</style>
