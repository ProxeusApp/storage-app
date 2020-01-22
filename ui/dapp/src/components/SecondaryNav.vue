<template>
  <div class="col-12 secondary-nav d-flex flex-row justify-content-center align-items-center fixed-top">
    <div class="d-flex flex-row mr-4">
      <file-upload-button @dropped="dropped"></file-upload-button>
      <button class="btn btn-light text-primary px-2 ml-3 mr-3" @click="$emit('toggleSyncRow')" data-tour-step="5">
        <span class="mdi md-24 mdi-sync"></span> <span class="badge badge-num badge-queue badge-info text-white ml-0"
                                                       :class="{visible:filesQueue.length > 0 || transactionQueue.length > 0}">{{ filesQueue.length + transactionQueue.length}}</span>
      </button>
    </div>
    <div class="search-box-container mx-auto container-fluid">
      <search-box :placeholder="$t('filebrowser.search.searchFiles', 'Search…')" :term="searchTerm"
                  @search="search"></search-box>
    </div>
    <div class="btn-group ml-auto" role="group" aria-label="Basic example">
      <button @click="setFileListViewType('list')" type="button" class="btn btn-light"
              :disabled="fileListViewType === 'list'">
        <i class="mdi mdi-view-list md-24"></i>
      </button>
      <button @click="setFileListViewType('grid')" type="button" class="btn btn-light"
              :disabled="fileListViewType === 'grid'">
        <i class="mdi mdi-view-grid md-18"></i>
      </button>
    </div>
  </div>
</template>

<script>
import SearchBox from './SearchBox'
import FileUploadButton from '@/components/FileUploadButton'

export default {
  name: 'secondary-nav',
  components: {
    SearchBox,
    FileUploadButton
  },
  computed: {
    transactionQueue () {
      return this.$store.state.file.transactionQueue
    },
    filesQueue () {
      return this.$store.state.file.filesQueue
    },
    fileListViewType () {
      return this.$store.state.fileListViewType
    },
    searchTerm () {
      return this.$store.state.file.searchTerm
    }
  },
  methods: {
    setFileListViewType (viewType) {
      this.$store.commit('SET_FILE_LIST_VIEW_TYPE', viewType)
    },
    dropped (file) {
      // TODO: remove, handle drag n drop on top level
      this.$emit('dropped', file)
    },
    async search (term) {
      this.$store.commit('SET_SEARCH_TERM', term)
      this.$store.dispatch('LOAD_FILES')
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../assets/styles/variables";

  .secondary-nav {
    top: 55px; // topnav height
    background: lighten($gray-300, 2%);
    padding-top: 8px;
    padding-bottom: 8px;
    vertical-align: middle;
    z-index: 1010;
    box-shadow: 0 1px 5px rgba(0, 0, 0, 0.03);
  }

  .sb-container {
    max-width: 500px;
  }

  .search-box-container {
    width: 100%;
    max-width: 500px;
  }

  /deep/ .search-box .form-control {
    height: 50px !important;
  }

  .badge-queue {
    opacity: 0;
    transform: translate(50px, -50px) scale(4);
    transition: all 350ms ease-in;
    position: absolute;

    &.visible {
      opacity: 1;
      transform: translate(0, 0) scale(1);
      transition: all 350ms ease;
    }
  }

  .mdi-view-grid {
    margin-bottom: 2px;
  }
</style>
