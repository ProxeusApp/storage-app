<template>
  <div class="overlay-container">
    <!--<transition name="slide">-->
    <!--<div class="notifications py-3 pl-4 pr-3" v-if="modal">-->
    <b-modal :id="'notificationModal' + _uid"
             :title="title"
             :lazy="true"
             ok-only
             :hide-footer="true"
             class="notification-modal"
             :ok-title="$t('close', 'Close')"
             :visible="modal"
             @hidden="$emit('modalClosed')">
      <div slot="modal-header" class="d-flex flex-column w-100 pl-1">
        <div class="d-flex flex-row align-items-center">
          <div class="d-flex flex-column">
            <button class="close-modal btn btn-round btn-primary p-2"
                    @click="$emit('modalClosed')">
              <i class="mdi mdi-arrow-right mdi-24px"></i>
            </button>
            <h2 class="mb-1">
              {{ $t('filebrowser.notifications.notifications', 'Notifications') }}
            </h2>
            <small>You have
              <strong>{{ unreadCount }}</strong>
              unread notifications and
              <strong>{{ pendingCount }}</strong>
              actions pending
            </small>
          </div>
          <div class="btn-group text-right ml-auto">
            <button type="button"
                    class="btn btn-primary btn-sm ml-auto"
                    @click="onShowAll"
                    :disabled="showPending === false">All
            </button>
            <button type="button"
                    class="btn btn-primary btn-sm ml-auto"
                    @click="onShowPending"
                    :disabled="showPending === true">Pending
            </button>
          </div>
        </div>
        <search-box class="w-100 mt-2" :placeholder="$t('filebrowser.search.searchFiles', 'Search…')" :term="searchTerm"
                    @search="search"></search-box>
      </div>
      <div class="pl-1">
        <div v-if="hasToday">
          <div class="d-flex align-items-end pb-1">
            <div v-if="showCategory('today')">{{ getCategoryHeader('today') }}</div>
            <button v-if="hasToday" type="button" class="btn btn-link btn-sm ml-auto" @click="markAllAsRead">
              {{ $t('filebrowser.notifications.mark_all_read', 'Mark all as read') }}
            </button>
          </div>
          <component v-bind:is="componentType(n.type)"
                     v-for="n in notificationsToday"
                     :notification="n"
                     :key="n.id"></component>
          <infinite-loading identifier="today" @infinite="infiniteHandlerToday">
            <div slot="no-more"></div>
            <div slot="no-results"></div>
          </infinite-loading>
        </div>
        <div v-if="hasYesterday">
          <div class="d-flex align-items-center pb-1">
            <div v-if="showCategory('yesterday')">{{ getCategoryHeader('yesterday') }}</div>
            <button v-if="!hasToday" type="button" class="btn btn-link btn-sm ml-auto" @click="markAllAsRead">
              {{ $t('filebrowser.notifications.mark_all_read', 'Mark all as read') }}
            </button>
          </div>
          <component v-bind:is="componentType(n.type)"
                     v-for="n in notificationsYesterday"
                     :notification="n"
                     :key="n.id"></component>
          <infinite-loading identifier="yesterday" @infinite="infiniteHandlerYesterday">
            <div slot="no-more"></div>
            <div slot="no-results"></div>
          </infinite-loading>
        </div>
        <div v-if="hasOlder">
          <div class="d-flex align-items-center pb-1">
            <div v-if="showCategory('older')">{{ getCategoryHeader('older') }}</div>
            <button v-if="!hasToday && !hasYesterday" type="button" class="btn btn-link btn-sm ml-auto"
                    @click="markAllAsRead">{{ $t('filebrowser.notifications.mark_all_read', 'Mark all as read') }}
            </button>
          </div>
          <component v-bind:is="componentType(n.type)"
                     v-for="n in olderNotifications"
                     :notification="n"
                     :key="n.id"></component>
          <infinite-loading identifier="Older" @infinite="infiniteHandlerOlder">
            <div slot="no-more">{{ $t('filebrowser.notifications.no_more_notifications', 'No more notifications') }}</div>
            <div slot="no-results">
              <div class="text-muted py-2" v-if="!hasOlder">
                <h2>{{ $t('filebrowser.notifications.no_notification', 'No notifications') }}</h2>
              </div>
            </div>
          </infinite-loading>
        </div>
      </div>
    </b-modal>
  </div>
</template>

<script>
import Backdrop from '@/components/Backdrop'
import BaseModal from '@/components/Modal/BaseModal'
import NotificationEntry from '@/components/File/NotificationEntry'
import SigningEntry from '@/components/File/SigningEntry'
import WorkflowEntry from '@/components/File/WorkflowEntry'
import SearchBox from '@/components/SearchBox'

export default {
  name: 'notification-modal',
  extends: BaseModal,
  components: {
    Backdrop,
    NotificationEntry,
    SigningEntry,
    WorkflowEntry,
    SearchBox
  },
  props: ['modal', 'title', 'text'],
  data () {
    return {
      pageInfo: {
        'today': { 'page': 1, 'end': 10 },
        'yesterday': { 'page': 1, 'end': 10 },
        'older': { 'page': 1, 'end': 10 }
      },
      itemsPerPage: 10
    }
  },
  computed: {
    hasToday () {
      return this.notificationsToday && this.notificationsToday.length > 0
    },
    hasYesterday () {
      return this.notificationsYesterday && this.notificationsYesterday.length > 0
    },
    hasOlder () {
      return this.olderNotifications && this.olderNotifications.length > 0
    },
    notificationsToday () {
      return this.$store.getters.notificationsToday({ 'end': this.pageInfo.today.end })
    },
    notificationsYesterday () {
      return this.$store.getters.notificationsYesterday({ 'end': this.pageInfo.yesterday.end })
    },
    olderNotifications () {
      return this.$store.getters.olderNotifications({ 'end': this.pageInfo.older.end })
    },
    unreadCount () {
      return this.$store.state.notification.notifications.filter(s => s.unread).length
    },
    pendingCount () {
      return this.$store.getters.pendingNotificationsCount
    },
    notificationsTodayCount () {
      return this.$store.getters.notificationsTodayCount
    },
    notificationsYesterdayCount () {
      return this.$store.getters.notificationsYesterdayCount
    },
    olderNotificationsCount () {
      return this.$store.getters.olderNotificationsCount
    },
    totalNotifications () {
      return this.$store.getters.totalNotifications
    },
    filterNotificationByOpts () {
      return this.$store.getters.filterNotificationByOpts
    },
    searchTerm () {
      return this.$store.state.notification.searchTerm
    },
    showPending () {
      return this.$store.state.notification.showPending
    }
  },
  methods: {
    showCategory (category) {
      let showCategory
      switch (category) {
        case 'today':
          showCategory = this.hasToday
          return showCategory
        case 'yesterday':
          showCategory = this.hasYesterday
          return showCategory
        case 'older' :
          showCategory = this.hasYesterday === false
          return showCategory
        default:
          showCategory = false
      }
      return showCategory
    },
    getCategoryHeader (category) {
      if (category === 'today') {
        return this.$t('filebrowser.notifications.today', 'Today')
      }
      if (category === 'yesterday') {
        return this.showCategory('today') ? this.$t('filebrowser.notifications.yesterday', 'Yesterday') : this.$t(
          'filebrowser.notifications.latest', 'Latest')
      }
      if (category === 'older') {
        return this.showCategory('today') ? this.$t('filebrowser.notifications.older', 'Older') : this.$t(
          'filebrowser.notifications.latest', 'Latest')
      }
      return ''
    },
    onShowPending () {
      this.$store.commit('SET_SHOW_PENDING', true)
    },
    onShowAll () {
      this.$store.commit('SET_SHOW_PENDING', false)
    },
    async markAllAsRead () {
      if (this.showPending === false && this.searchTerm === '') {
        this.$store.dispatch('SET_ALL_NOTIFICATIONS_AS', { unread: false })
      } else {
        this.$store.dispatch('SET_FILTERED_NOTIFICATIONS_AS', { unread: false })
      }
    },
    nextPage (category, pageInfo) {
      pageInfo.page = pageInfo.page === undefined ? 1 : pageInfo.page
      pageInfo.start = (pageInfo.start === undefined) ? pageInfo.page * this.itemsPerPage : pageInfo.start
      let items = this.itemsPerPage
      let remaining
      if (category === 'today') {
        remaining = this.notificationsTodayCount - pageInfo.start
      }
      if (category === 'yesterday') {
        remaining = this.notificationsYesterdayCount - pageInfo.start
      }
      if (category === 'older') {
        remaining = this.olderNotificationsCount - pageInfo.start
      }
      if (remaining < items && remaining > 0) {
        items = remaining
      }
      const end = pageInfo.start + items
      return { 'start': pageInfo.start, 'end': end, 'remaining': remaining, 'page': pageInfo.page }
    },
    updatePage (scrollState, pageInfo) {
      if (pageInfo.remaining < 1) {
        scrollState.complete()
        return pageInfo
      } else {
        pageInfo.page += 1
        scrollState.loaded()
        return pageInfo
      }
    },
    infiniteHandlerToday (scrollState) {
      this.pageInfo.today = this.updatePage(scrollState, this.nextPage('today', { 'page': this.pageInfo.today.page }))
    },
    infiniteHandlerYesterday (scrollState) {
      this.pageInfo.yesterday = this.updatePage(scrollState,
        this.nextPage('yesterday', { 'page': this.pageInfo.yesterday.page }))
    },
    infiniteHandlerOlder (scrollState) {
      this.pageInfo.older = this.updatePage(scrollState, this.nextPage('older', { 'page': this.pageInfo.older.page }))
    },
    componentType: function (type) {
      if (type === 'signing_request') {
        return 'signing-entry'
      } else if (type === 'workflow_request') {
        return 'workflow-entry'
      }
      return 'notification-entry'
    },
    search (term) {
      term = term === '' ? undefined : term
      this.$store.commit('SET_NOTIFICATION_SEARCH_TERM', term)
    }
  }
}

</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .close-modal {
    position: absolute;
    margin-left: -63px;
    margin-top: -4px;
  }

  /deep/ .notification-modal {
    .modal {
      .modal-dialog {
        position: fixed;
        width: 50%;
        min-width: 750px;
        margin: auto;
        right: 0;
        top: 0;
        bottom: 0;
        border-radius: 0;

        .modal-header {
          border-radius: 0;
          width: 100%;
          display: block;
        }

        .modal-content {
          background: white;
          border-radius: 0;
          height: 100%;
          overflow-x: visible;
          overflow-y: auto;
          position: static;
        }

        .modal-body {
          padding-top: $spacer / 2;
        }
      }

      &.fade {
        .modal-dialog {
          transition: all 0.2s ease-out;
          transform: translate(20%, 0);
          opacity: 0;
        }
      }

      &.show {
        .modal-dialog {
          transform: translate(0, 0);
          opacity: 1;
        }
      }
    }
  }

  .notifications {
    position: fixed;
    right: 0;
    top: 0;
    height: 100%;
    width: 50%;
    min-width: 700px;
    background: white;
    z-index: 1040;
    box-shadow: 0 40px 100px rgba(0, 0, 0, 0.35);
  }

  .slide-enter-active {
    transition: all 0.1s ease;
    background: white;
  }

  .slide-leave-active {
    transition: all 0.1s cubic-bezier(1, 0.5, 0.8, 1);
    background: white;
  }

  .slide-enter,
  .slide-leave-to {
    transform: translateX(100px);
    opacity: 0;
  }
</style>
