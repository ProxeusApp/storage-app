<template>
  <form enctype="multipart/form-data" novalidate>
    <button v-if="uploadModalLoading" class="btn btn-primary spinner-wrapper">
      <tiny-spinner></tiny-spinner>
      <div class="invisible">{{ $t('filebrowser.fileupload.upload_file', 'Upload File') }}</div><!-- ensure same button size as upload-button for all translations -->
    </button>
    <div v-else class="btn btn-primary" data-tour-step="4">
      {{ $t('filebrowser.fileupload.upload_file', 'Upload File') }}
      <input class="btn-file-input" type="file" name="file"
             @click="clicked"
             @change="filesChange($event.target.name, $event.target.files); fileCount = $event.target.files.length"/>
    </div>
  </form>
</template>

<script>
import TinySpinner from '@/components/TinySpinner'

export default {
  name: 'file-upload-button',
  components: {
    TinySpinner
  },
  data () {
    return {
      newFileInput: undefined
    }
  },
  computed: {
    uploadModalLoading () {
      return this.$store.state.file.uploadModalLoading
    }
  },
  methods: {
    clicked (e) {
      this.$el.reset()
    },
    filesChange (fieldName, fileList) {
      this.$emit('dropped', fileList[0])
    }
  }
}
</script>

<style scoped>
  .btn.btn-primary {
    position: relative;
  }

  .btn-file-input {
    position: absolute;
    font-size: 22px;
    opacity: 0;
    right: 0;
    top: 0;
    cursor: pointer;
  }

  .spinner-wrapper {
    cursor: default;
    pointer-events: none;
    height: 2.45em;
    overflow: hidden;
  }
</style>
