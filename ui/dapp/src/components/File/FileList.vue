<template>
  <div class="file-list-container">
    <div class="text-center mt-3" v-if="files.length === 0 && filesLoading === false">
      <h2 class="text-muted">{{ $t('filelist.no_files_found', 'No files found') }}</h2>
    </div>
    <spinner background="transparent" v-if="filesLoading === true"></spinner>
    <div class="file-grid-view" v-if="files.length > 0 && filesLoading === false && fileListViewType === 'grid'">
      <file-grid-view v-for="(file, index) in files" :index="index" :file="file" :key="file.id"/>
    </div>
    <div class="file-list-view list-group" v-if="files.length > 0 && filesLoading === false && fileListViewType === 'list'">
      <file-list-view v-for="(file, index) in files" :index="index" :file="file" :key="file.id"/>
    </div>
  </div>
</template>

<script>
import FileListView from './FileListView'
import FileGridView from './FileGridView'
import Spinner from '@/components/Spinner'

export default {
  name: 'file-list',
  components: {
    FileListView,
    FileGridView,
    Spinner
  },
  props: {
    files: {
      type: Array,
      required: true
    },
    filesLoading: {
      type: Boolean
    },
    view: {
      type: String,
      default: 'grid'
    }
  },
  computed: {
    fileListViewType () {
      return this.$store.state.fileListViewType
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .file-list-container {
    position: relative;
  }

  .file-grid-view {
    display: grid;
    grid-auto-columns: 180px;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    grid-gap: 1.25rem;
  }
</style>
