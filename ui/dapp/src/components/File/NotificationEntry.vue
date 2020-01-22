<template>
  <div class="notification d-flex flex-row py-1 pl-2 pr-1 bg-light mb-2 w-100 align-items-center" v-if="notification">
    <div class="d-flex w-25 flex-column mr-auto flex-wrap flex-fill">
      <div class="d-flex justify-content-between text-muted" data-togle="tooltip" :title="formatDate">
        <small>{{ i18nNotification.name }}</small>
        <small>{{ momentFromNow }}</small>
      </div>
      <div class="description">{{ i18nNotification.desc }}</div>
    </div>
    <slot></slot>
    <button v-if="notification.type != 'signing_request'"
            v-bind:title="dismissTooltip"
            v-tooltip="{content: dismissTooltip}"
            type="button"
            class="btn btn-light btn-sm px-1 py-1"
            @click="dismiss">
      <i class="icon mdi mdi-close md-18 ml-auto"></i>
    </button>
    <button v-bind:title="unreadTooltip"
            v-tooltip="{content: dismissTooltip}"
            type="button"
            class="btn btn-light btn-sm px-1 py-1"
            @click="toggleRead">
      <i class="icon mdi md-18 ml-auto" v-bind:class="{
        'mdi-checkbox-blank-circle': notification.unread,
        'mdi-checkbox-blank-circle-outline': !notification.unread }"></i>
    </button>
  </div>
</template>

<script>
import moment from 'moment'
import { fromWei } from 'web3-utils'

export default {
  name: 'notification-entry',
  props: ['notification'],
  computed: {
    i18nNotification () {
      let t = { name: 'unknown', desc: 'unknown notification type' }
      switch (this.notification.type) {
        case 'signing_request':
          t.name = this.$t('filebrowser.notifications.notification_signing_request', '***Signing request')
          t.desc = this.$t('filebrowser.notifications.notification_signing_request_desc',
            '***You received a signing request from {from} to sign file {fileName}', {
              from: this.owner,
              fileName: this.fileName
            })
          break
        case 'signing_request_removed':
          t.name = this.$t('filebrowser.notifications.notification_signing_request_removed', '***Signing request')
          t.desc = this.$t('filebrowser.notifications.notification_signing_request_removed_desc',
            '***A signing request from {from} is no longer available to sign file {fileName}', {
              from: this.owner,
              fileName: this.fileName
            })
          break
        case 'share_request':
          t.name = this.$t('filebrowser.notifications.notification_share_request', '***Share request')
          t.desc = this.$t('filebrowser.notifications.notification_share_request_desc',
            '***File {fileName} was shared with you by {from}', {
              from: this.owner,
              fileName: this.fileName
            })
          break
        case 'workflow_request':
          t.name = this.$t('filebrowser.notifications.notification_workflow_request', '***Workflow pending')
          t.desc = this.$t('filebrowser.notifications.notification_workflow_request_desc',
            'Process {fileName} awaits your action', {
              from: this.owner,
              fileName: this.fileName
            })
          break
        case 'workflow_shared':
          t.name = this.$t('filebrowser.notifications.notification_workflow_shared', '***Workflow shared')
          t.desc = this.$t('filebrowser.notifications.notification_workflow_shared_desc',
            '***Workflow {workflowName} has been sent to {recipient}', {
              recipient: this.who,
              workflowName: this.workflowName
            })
          break
        case 'tx_xes_approval':
          t.name = this.$t('filebrowser.notifications.notification_tx_xes_approve',
            '***You successfully changed your XES allowance')
          t.desc = this.$t('filebrowser.notifications.notification_tx_xes_approve_desc',
            '***You successfully changed your XES allowance to {xesAmount} XES',
            { xesAmount: this.xesAmount })
          break
        case 'tx_register':
          t.name = this.$t('filebrowser.notifications.notification_tx_register', '***Transaction: register')
          t.desc = this.$t('filebrowser.notifications.notification_tx_register_desc',
            '***File {fileName} as been successfully registered on the blockchain', {
              fileName: this.fileName
            })
          break
        case 'tx_remove':
          t.name = this.$t('filebrowser.notifications.notification_tx_remove', '***Transaction: remove')
          t.desc = this.$t('filebrowser.notifications.notification_tx_remove_desc',
            '***File {fileName} has been removed',
            { fileName: this.fileName })
          break
        case 'tx_share':
          t.name = this.$t('filebrowser.notifications.notification_tx_share', '***Transaction: share')
          t.desc = this.$t('filebrowser.notifications.notification_tx_share_desc',
            '***File {fileName} has been shared with {who}',
            { fileName: this.fileName, who: this.who })
          break
        case 'tx_sign':
          t.name = this.$t('filebrowser.notifications.notification_tx_sign', '***Transaction: sign')
          t.desc = this.$t('filebrowser.notifications.notification_tx_sign_desc',
            '***File {fileName} has been signed by you',
            { fileName: this.fileName })
          break
        case 'ev_notifysign':
          t.name = this.$t('filebrowser.notifications.notification_ev_notifysign', '***Event: sign')
          t.desc = this.$t('filebrowser.notifications.notification_ev_notifysign_desc',
            '***File {fileName} has been signed by {who}',
            { fileName: this.fileName, who: this.who })
          break
        case 'tx_signRequest':
          t.name = this.$t('filebrowser.notifications.notification_tx_signRequest', '***Transaction: sign request')
          t.desc = this.$t('filebrowser.notifications.notification_tx_signRequest_desc',
            '***You successfully sent signing request for file {fileName} to {who}', {
              fileName: this.fileName, who: this.who
            })
          break
        case 'tx_revoke':
          t.name = this.$t('filebrowser.notifications.notification_tx_revoke', '***Transaction: revoke')
          t.desc = this.$t('filebrowser.notifications.notification_tx_revoke_desc',
            '***File {fileName} has been revoked',
            { fileName: this.fileName })
          break
        // Not sure we need this generic tx notification success
        case 'tx_success':
          t.name = this.$t('filebrowser.notifications.notification_tx_success', '***Transaction successful')
          t.desc = this.$t('filebrowser.notifications.notification_tx_success', '***Transaction successful')
          break
        case 'file_about_to_expire':
          t.name = this.$t('filebrowser.notifications.notification_file_about_to_expire', '***File is about to expire')
          t.desc = this.$t('filebrowser.notifications.notification_file_about_to_expire_desc',
            '***File {fileName} is going to expire on {expiryDate}.',
            { fileName: this.fileName, expiryDate: this.fileExpiryDate })
          break
        case 'file_grace_period':
          t.name = this.$t('filebrowser.notifications.notification_file_grace_period', '***File is in grace period')
          t.desc = this.$t('filebrowser.notifications.notification_file_grace_period_desc',
            '***File {fileName} has expired and is going to be deleted from the Storage Provider on {gracePeriodEndDate}.',
            { fileName: this.fileName, gracePeriodEndDate: this.fileGracePeriodEndDate })
          break
        case 'file_expired':
          t.name = this.$t('filebrowser.notifications.notification_file_expired', '***File has expired')
          t.desc = this.$t('filebrowser.notifications.notification_file_expired_desc',
            '***File {fileName} has expired and was deleted from the Storage Provider.',
            { fileName: this.fileName })
          break
        case 'tx_xes_send':
          t.name = this.$t('filebrowser.notifications.notification_tx_xes_send', 'XES sent')
          t.desc = this.$t('filebrowser.notifications.notification_tx_xes_send_desc',
            'You sent {xes} XES to {to}.',
            { xes: this.xesAmount, to: this.who })
          break
        case 'tx_xes_receive':
          t.name = this.$t('filebrowser.notifications.notification_tx_xes_receive', 'XES received')
          t.desc = this.$t('filebrowser.notifications.notification_tx_xes_receive_desc',
            'You received {xes} XES from {from}.',
            { xes: this.xesAmount, from: this.who })
          break
        case 'tx_eth_increase':
          t.name = this.$t('filebrowser.notifications.notification_tx_eth_increase', 'ETH received')
          t.desc = this.$t('filebrowser.notifications.notification_tx_eth_increase_desc',
            'ETH holdings increased by {eth} ETH.',
            { eth: this.ethAmount, to: this.who })
          break
        case 'tx_eth_decrease':
          t.name = this.$t('filebrowser.notifications.notification_tx_eth_decrease', 'ETH sent')
          t.desc = this.$t('filebrowser.notifications.notification_tx_eth_decrease_desc',
            'ETH holdings decreased by {eth} ETH.',
            { eth: this.ethAmount, from: this.who })
          break
        default:
          break
      }
      return t
    },
    fileName () {
      return this.notification.data.fileName || ''
    },
    workflowName () {
      return this.notification.data.workflowName !== undefined ? this.notification.data.workflowName : ''
    },
    owner () {
      if (this.notification.data.owner === null) {
        return ''
      }
      let res = this.notification.data.owner !== undefined ? this.$store.getters.nameByAddress(
        this.notification.data.owner)
        : ''
      return res !== undefined ? res : ''
    },
    xesAmount () {
      return this.notification.data.xesAmount !== undefined ? fromWei(String(this.notification.data.xesAmount)) : ''
    },
    ethAmount () {
      return this.notification.data.ethAmount !== undefined ? fromWei(String(this.notification.data.ethAmount)) : ''
    },
    who () {
      if (this.notification.data.who !== undefined) {
        const a = this.$t('filebrowser.notifications.notification_and', 'and')
        return this.$store.getters.namesByAddresses(this.notification.data.who)
          .join(', ')
          .replace(/,(?=[^,]*$)/, ` ${a}`)
      }
      return []
    },
    dismissTooltip () {
      return this.$t('filebrowser.notifications.dismiss', 'Dismiss')
    },
    unreadTooltip () {
      return this.notification.unread ? this.$t('filebrowser.notifications.mark_as_read', 'Mark as read') : this.$t(
        'filebrowser.notifications.mark_as_unread', 'Mark as unread')
    },
    momentFromNow () {
      return moment.unix(this.notification.timestamp).fromNow()
    },
    formatDate () {
      let mDate = moment.unix(this.notification.timestamp)
      let date = mDate.format('dddd MMMM Do')
      let time = mDate.format('h:mm a')
      return date + ' at ' + time
    },
    fileExpiryDate () {
      return this.notification.data.expiry !== undefined ? moment.unix(this.notification.data.expiry).format('DD. MMMM YYYY') : ''
    },
    fileGracePeriodEndDate () {
      return this.notification.data.graceSeconds !== undefined ? moment.unix(this.notification.data.expiry).add(this.notification.data.graceSeconds, 'seconds').format('DD. MMMM YYYY') : ''
    }
  },
  methods: {
    dismiss () {
      this.$store.dispatch('DELETE_NOTIFICATION', this.notification).catch(err => {
        console.log(err)
      })
    },
    toggleRead () {
      this.$store.dispatch('SET_NOTIFICATION_AS', { notification: this.notification, unread: !this.notification.unread })
        .catch(err => {
          console.log(err)
        })
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables.scss";

  .notification {
    background: $gray-110;
    position: relative;
    margin-bottom: 0.5rem;

    &:hover {
      background-color: $gray-100;
    }
  }

  .description {
    font-size: 0.85rem;
    width: 100%;
    word-wrap: break-word;
  }

  .trim {
    word-wrap: break-word;
    overflow: hidden;
    min-width: 0;
    text-overflow: ellipsis;
    white-space: nowrap;
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
