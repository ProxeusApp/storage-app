<template>
  <div @click="showPreview" class="preview" v-bind="$attrs">
    <div v-if="loading" class="preview-placeholder d-flex align-items-center text-center">
      <div class="preview-icon w-100" :style="{'font-size':fontSize}">
        <tiny-spinner></tiny-spinner>
      </div>
    </div>
    <template v-else>
      <img v-if="hasThumbnail === true" width="100%" :src="previewSrc" :alt="file.filename">
      <div v-if="hasThumbnail === false" class="preview-placeholder d-flex align-items-center text-center">
        <div class="preview-icon w-100" :style="{'font-size':fontSize}">
          <span class="mdi" :class="defaultThumbnailClass"></span>
        </div>
      </div>
      <div v-if="getFileKind === 2" class="preview-placeholder d-flex align-items-center text-center">
        <div class="preview-icon w-100" :style="{'font-size':fontSize}">
          <span class="mdi mdi-file-chart"></span>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
import TinySpinner from '@/components/TinySpinner'

export default {
  name: 'file-icon',
  components: { TinySpinner },
  props: {
    file: {
      required: true
    },
    fontSize: {
      default: '3rem'
    },
    loading: {
      default: false
    }
  },
  computed: {
    hasThumbnail () {
      return this.file.hasThumbnail === true
    },
    getFileKind () {
      return this.file.fileKind
    },
    previewSrc () {
      return '/api/file/thumb/' + this.file.id
    },
    defaultThumbnailClass () {
      return this.$store.getters.defaultThumbnailClass(this.file)
    }
  },
  methods: {
    showPreview () {
      this.$emit('showPreview')
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .preview {
    width: 70px;
    height: 70px;
    background: #dee2e6;
    overflow: hidden;
    min-width: 70px;

    img {
      object-fit: cover;
      width: 100%;
      height: 100%;
    }

    .preview-placeholder {
      position: relative;
      width: 100%;
      height: 0;
      padding-bottom: 100%;
      background: $gray-300;
      text-align: center;
    }

    .preview-icon {
      display: block;
      overflow: hidden;
      position: absolute;
      top: 0;
      right: 0;
      bottom: 0;
      left: 0;
      color: $primary;

      .mdi {
        transform: translate(-50%, -50%);
        position: absolute;
        top: 50%;
        left: 50%;
      }
    }

    .spinner {
      transform: translate(-50%, -50%);
      position: absolute;
      top: 50%;
      left: 50%;
      margin: 0 !important;
    }

    &.preview-grid {
      position: relative;
      width: 100%;
      height: 0;
      padding-bottom: 100%;

      img {
        position: absolute;
        top: 0;
        left: 0;
      }
    }
  }
</style>
