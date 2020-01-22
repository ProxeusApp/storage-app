<template>
  <b-modal :id="file.hash"
           :title="$t('filebrowser.fileupload.download', 'Download')"
           :lazy="true"
           size="lg"
           no-fade
           hide-footer
           :visible="modal"
           @hidden="$emit('modalClosed')">
    <div slot="modal-header" class="d-flex flex-row w-100 align-items-center">
      <h5 class="modal-title text-truncate flex-fill mr-auto">{{ file.filename || file.id }}</h5>
      <file-actions class="mx-3"
                    v-if="isFileRemoving === false && isFileExpired === false"
                    in-file-preview
                    @download="download"
                    @sharePrompt="sharePrompt"
                    @signRequestPrompt="signRequestPrompt"
                    @unsharePrompt="unsharePrompt"
                    @removeFile="removeFile"
                    @removeFilePrompt="removeFilePrompt"
                    :file="file"></file-actions>
      <button type="button" class="close" @click="$emit('modalClosed')" aria-label="Close">
        <span class="mdi md-24 mdi-window-close"></span>
      </button>
    </div>
    <div class="row">
      <div class="col-sm-5 pr-0">
        <file-thumbnail :file="file"></file-thumbnail>
      </div>
      <div class="col-sm-7">
        <file-info :file="file" v-if="file"></file-info>
      </div>
    </div>
  </b-modal>
</template>

<script>
import FileInfo from './FileInfo'
import FileActions from './FileActions'
import FileThumbnail from './FileThumbnail'

export default {
  name: 'file-preview',
  props: ['modal', 'file'],
  components: {
    FileInfo,
    FileActions,
    FileThumbnail
  },
  computed: {
    isFileRemoving () {
      return this.$store.getters.isFileRemoving(this.file)
    },
    isFileExpired () {
      return this.file.expired
    }
  },
  methods: {
    download () {
      this.$emit('download')
    },
    sharePrompt () {
      this.$emit('sharePrompt')
    },
    signRequestPrompt () {
      this.$emit('signRequestPrompt')
    },
    unsharePrompt () {
      this.$emit('unsharePrompt')
    },
    removeFile () {
      this.$emit('removeFile')
    },
    removeFilePrompt () {
      this.$emit('removeFilePrompt')
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  /deep/ .modal-dialog {
    min-width: 400px;
    max-width: 800px;
  }
</style>
