<template>
  <nav class="sidebar sidebar-light side-bar-tour" :class="{toggled:sidebarToggled, 'sidebar-wide':wideSidebar}">
    <nav class="collapse show sidebar-sticky">
      <nav class="nav main-nav flex-column">
        <li class="nav-item">
          <button @click="changeActiveCategory('all-files')"
                  class="btn btn-link nav-link"
                  :class="{active: activeCategory === 'all-files'}"
                  data-toggle="tooltip"
                  data-placement="right"
                  data-boundary="window"
                  :title="$t('filebrowser.sidebar.all_files', 'All files')"><span
            class="material-icons mdi mdi-folder"></span><span
            class="nav-link-title">{{ $t('filebrowser.sidebar.all_files', 'All Files') }}</span>
          </button>
        </li>
        <li class="nav-item">
          <button @click="changeActiveCategory('my-files')"
                  class="btn btn-link nav-link"
                  :class="{active: activeCategory === 'my-files'}"
                  data-toggle="tooltip"
                  data-placement="right"
                  data-boundary="window"
                  :title="$t('filebrowser.sidebar.my_files', 'My files')"><span
            class="material-icons mdi icon mdi-folder-account"></span><span
            class="nav-link-title">{{ $t('filebrowser.sidebar.my_files', 'My files') }}</span>
          </button>
        </li>
        <li class="nav-item">
          <button @click="changeActiveCategory('shared-with-me')"
                  class="btn btn-link nav-link"
                  :class="{active: activeCategory === 'shared-with-me'}"
                  data-toggle="tooltip"
                  data-placement="right"
                  data-boundary="window"
                  :title="$t('filebrowser.sidebar.shared_with_me', 'Shared with me')"><span
            class="material-icons mdi mdi-account-group"></span><span
            class="nav-link-title">{{ $t('filebrowser.sidebar.shared_with_me', 'Shared with me') }}</span>
          </button>
        </li>
        <li class="nav-item">
          <button @click="changeActiveCategory('signed-by-me')"
                  class="btn btn-link nav-link"
                  :class="{active: activeCategory === 'signed-by-me'}"
                  data-toggle="tooltip"
                  data-placement="right"
                  data-boundary="window"
                  :title="$t('filebrowser.sidebar.signed_by_me', 'Signed by me')"><span
            class="material-icons mdi mdi-feather"></span><span
            class="nav-link-title">{{ $t('filebrowser.sidebar.signed_by_me', 'Signed by me') }}</span>
          </button>
        </li>
        <li class="nav-item">
          <button @click="changeActiveCategory('expired-files')"
                  class="btn btn-link nav-link"
                  :class="{active: activeCategory === 'expired-files'}"
                  data-toggle="tooltip"
                  data-placement="right"
                  data-boundary="window"
                  :title="$t('filebrowser.sidebar.expired_files', 'Expired files')">
            <span class="material-icons mdi mdi-folder-clock"></span>
            <span class="nav-link-title">{{ $t('filebrowser.sidebar.expired_files', 'Expired files') }}</span>
          </button>
        </li>
      </nav>
      <ul class="nav secondary-sidebar-nav flex-column">
        <li class="nav-item">
          <a
            :href="$t('filebrowser.sidebar.handbook_url', 'https://docs.google.com/document/d/1-YRJAoB_t7IK7IAYBFWEKW7vhci6hewS5AKnX5BUafw/preview')"
            class="nav-link" data-toggle="tooltip" data-placement="right"
            data-boundary="window" :title="$t('filebrowser.sidebar.handbook', 'Recent')" target="_blank"><span
            class="material-icons mdi mdi-help-circle-outline"></span><span
            class="nav-link-title">{{ $t('filebrowser.sidebar.handbook', 'Handbook') }}</span>
          </a>
        </li>
        <li class="nav-item nav-item-toc">
          <a
            :href="$t('filebrowser.sidebar.terms_and_conditions_url', 'https://docs.google.com/document/d/1bLIyQi16L6DGv7N5n_UxS8BidFLc--LtGo0Sf7qanN0/preview')"
            class="nav-link" data-toggle="tooltip" data-placement="right"
            data-boundary="window" :title="$t('filebrowser.sidebar.terms_and_conditions', 'Recent')"
            target="_blank"><span
            class="material-icons mdi mdi-file-document-outline"></span><span
            class="nav-link-title">{{ $t('filebrowser.sidebar.terms_and_conditions', 'TOC') }}</span>
          </a>
        </li>
        <li v-if="version !== undefined" v-show="!sidebarToggled" class="nav-item nav-item-slider py-1 px-3">
          <small class="text-muted mr-1">{{ $t('contract.version.info', 'Contract v.') }} {{ version.contract }}</small>
          <small class="text-muted">{{ $t('build.version.info', 'Build v.') }} {{ version.build }}</small>
        </li>
        <li class="nav-item nav-item-slider">
          <button class="nav-link-slider nav-link btn btn-block btn-link text-left border-top"
                  @click="toggleSidebar">
                  <span class="material-icons mdi mdi-chevron-double-left"
                        :class="[sidebarToggled ? 'mdi-chevron-double-right' : 'mdi-chevron-double-left']"></span>
          </button>
        </li>
      </ul>
    </nav>
  </nav>
</template>

<script>
export default {
  name: 'category-sidebar',
  computed: {
    activeCategory () {
      return this.$store.state.file.activeCategory
    },
    sidebarToggled () {
      return this.$store.state.file.categorySidebar.toggled
    },
    wideSidebar () {
      return this.$store.state.file.categorySidebar.size === 'wide'
    },
    version () {
      return this.$store.state.version
    }
  },
  created () {
    this.$store.commit('TOGGLE_CATEGORY_SIDEBAR', { toggled: true, size: 'normal' })
  },
  methods: {
    changeActiveCategory (view) {
      this.$emit('changeActiveCategory', view)
    },
    toggleSidebar () {
      this.$store.commit('TOGGLE_CATEGORY_SIDEBAR', { toggled: !this.sidebarToggled, size: 'normal' })
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../assets/styles/variables";

  .sidebar {
    grid-area: sidebar;
  }

  .sidebar-wide {
    max-width: 350px;
    width: 350px;

    .secondary-sidebar-nav {
      max-width: 350px;
      width: 350px;
    }

    .nav-link-slider {
      max-width: 350px;
      width: 350px;
    }
  }

  .nav-link.active {
    color: $primary;
  }
</style>
