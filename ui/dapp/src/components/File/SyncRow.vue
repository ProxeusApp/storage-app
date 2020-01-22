<template>
  <div class="row sync-row" :class="{'open':syncVisible}">
    <h2 class="w-100"
        v-if="filesQueue.length === 0 && transactionQueue.length === 0">{{ $t('filebrowser.filequeue.nofiles', 'No files in queue') }}</h2>
    <div class="file-queue w-100 transaction-queue">
      <div v-for="transaction in transactionQueue" :key="transaction.txHash" class="file-queue-file p-1 w-100 d-flex flex-row align-items-center">
        <div class="meta pl-1 d-flex flex-column text-truncate mr-auto">
          <div class="percentage" :style="{width: '' + getPercentage(transaction) + ''}">&nbsp;</div>
          <div class="file--name text-truncate">
            <small>{{ humanizeTransactionName(transaction) }}</small>
          </div>
          <div class="file--status text-truncate text-danger"
               v-if="transaction.error === true || transaction.status === 'fail'">
            <small>{{ humanizeErrorMessage(transaction) }}</small>
          </div>
        </div>
        <div class="m-1" v-if="transaction.txHash">
          <a :href="blockchainNetUrl + '/tx/' + transaction.txHash"
             target="_blank" onclick="window.openInBrowser(event, this);"
             :title="$t('fileJS.transaction_queue.txlink', 'Link to Transaction')">
            <i class="mdi mdi-launch"></i>
          </a>
        </div>
        <div class="spinner m-1 text-center">
          <div class="tinyspinner" v-show="transaction.status !== 'fail'"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import web3Utils from 'web3-utils'

export default {
  name: 'sync-row',
  props: ['syncVisible'],
  computed: {
    ...mapState({
      transactionQueue: state => state.file.transactionQueue,
      filesQueue: state => state.file.filesQueue,
      fileNames: state => state.file.fileNames
    }),
    blockchainNetUrl () {
      return this.$store.getters.etherscanUrl
    }
  },
  methods: {
    formattedFilename (filename) {
      return filename.substr(0, 4) + '…' + filename.substr(filename.length - 4)
    },
    humanizeErrorMessage (tx) {
      return (tx.error.message || tx.error || this.$t('filebrowser.filequeue.failed', 'Failed'))
    },
    // TODO: Add infos to tx object
    humanizeTransactionName (tx) {
      var n

      switch (tx.name) {
        case 'xes-approve':
          n = web3Utils.fromWei(tx.xesAmount + '', 'ether')
          return this.$t('fileJS.transaction_queue.approve_xes', 'Approving XES…', {
            value: n || 0
          })
        case 'xes-transfer':
          n = web3Utils.fromWei(tx.xesAmount + '', 'ether')
          return this.$t('fileJS.transaction_queue.send_xes', 'Sending XES…', {
            value: n || 0
          })
        case 'eth-transfer':
          n = web3Utils.fromWei(tx.xesAmount + '', 'ether')
          return this.$t('fileJS.transaction_queue.send_eth', 'Sending ETH…', {
            value: n || 0
          })
        case 'upload':
          return this.$t('filebrowser.filequeue.uploading', 'Uploading {filename} to Storage Provider…',
            { filename: tx.fileName })
        case 'register':
          return this.$t('filebrowser.filequeue.registering', 'Registering', { filename: tx.fileName })
        case 'sign':
          return this.$t('fileJS.transaction_queue.sign_file', 'Signing', { filename: tx.fileName })
        case 'remove':
          return this.$t('filebrowser.filequeue.removing', 'Removing {filename}', { filename: tx.fileName })
        case 'share':
          return this.$t('fileJS.transaction_queue.share', 'Sharing {filename}', { filename: tx.fileName })
        case 'revoke':
          return this.$t('fileJS.transaction_queue.unshare', 'Unsharing {filename}', { filename: tx.fileName })
        case 'download':
          return this.$t('fileJS.transaction_queue.synchronizing', 'Synchronizing {filename}...',
            { filename: tx.fileName })
        case 'requestSign':
          return this.$t('fileJS.transaction_queue.sign_request', 'Requesting signature for {filename}',
            { filename: tx.fileName })
        default:
          return tx.name
      }
    },
    getPercentage (transaction) {
      return transaction.percentage + '%'
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .nav-link-title {
    color: darken($text-muted, 15%);
  }

  .sync-row {
    background: white;
    transform: translateY(-300px);
    opacity: 0;
    transition: all 250ms ease;
    position: fixed;
    top: calc(55px + 66px); // topnav height + secondary-nav height
    z-index: 1000;
    padding: 1.5rem;
    width: 100%;
    box-shadow: 0 40px 100px rgba(0, 0, 0, 0.35);

    &.open {
      opacity: 1;
      transform: translateY(0);
      transition: all 250ms ease;
    }
  }

  .w-90 {
    width: 90%;
  }

  .file-queue {
    display: grid;
    grid-auto-columns: 330px;
    grid-template-columns: repeat(auto-fill, minmax(330px, 1fr));
    grid-gap: 1.5rem;
  }

  .transaction-queue {
    .file--name {
      max-width: 100%;
    }

    .file--status {
      max-width: 100%;
    }
  }

  .file-queue-file {
    position: relative;
    border-radius: $border-radius;
    background: $gray-200;

    .file--name {
      max-width: 100%;
      display: block;
    }

    .percentage {
      position: absolute;
      width: 0; /* starting point */
      left: 0;
      top: 0;
      border-radius: $border-radius;
      opacity: 0.2;
      height: calc(100%);
      background-color: $cyan;
    }
  }

  .spinner {
    min-width: 1rem;
    height: 20px;
    display: block;
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
