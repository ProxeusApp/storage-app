<template>
  <notification-entry :notification="notification"
                      :key="notification.id">
    <button type="button" class="btn btn-light btn-sm px-1 py-1"
            @click="open">
      {{ openActionText }}
    </button>
    <!--<form ref="form" name="datatransfer" action="http://localhost:3005/froparea" method="post">-->
    <!--<input type="text" name="file" id="file" v-bind="fileBlob"/>-->
    <!--</form>-->
  </notification-entry>
</template>

<script>
import NotificationEntry from '@/components/File/NotificationEntry'

export default {
  name: 'workflow-entry',
  props: ['notification'],
  components: {
    NotificationEntry
  },
  computed: {
    openActionText () {
      return this.$t('filebrowser.notifications.notification_workflow_open', 'Open in Workflow Manager')
    }
  },
  methods: {
    async open () {
      const response = await this.$store.dispatch('OPEN_PROCESS', this.notification.data.hash)
      if (response === true) {
        this.$store.dispatch('SET_NOTIFICATION_AS', { notification: this.notification, unread: false })
          .catch(err => {
            console.log(err)
          })
      }
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables.scss";

  form[name=datatransfer] input {
    display: none;
  }

  .cursor-default {
    cursor: default;
  }

  .file-name,
  .address-name {
    font-size: 0.8rem;
  }

  .trim {
    word-wrap: break-word;
    overflow: hidden;
    min-width: 0;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .spinner {
    min-width: 1rem;
    height: 1rem;
    display: inline-block;
  }

  .tinyspinner {
    display: inline-block;
    width: 1rem;
    height: 1rem;
    border: 0.15rem solid $gray-500;
    border-bottom: 0.15rem solid rgba(0, 0, 0, 0);
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

  @-webkit-keyframes rotating {
    from {
      -webkit-transform: rotate(0deg);
    }

    to {
      -webkit-transform: rotate(360deg);
    }
  }
</style>
