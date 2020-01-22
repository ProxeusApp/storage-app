import Vue from 'vue'
import ChannelHub from '../../lib/ChannelHub'

const websocketEndpoint = 'pipe'
let channelHub

const mutations = {}

const getters = {}

const actions = {
  INIT_CHANNEL_HUB ({ rootState, state, commit, dispatch }) {
    if (!channelHub || channelHub.isClosed()) {
      channelHub = new ChannelHub(websocketEndpoint) // On logout and other cases we close the connection with the websocket. Reopen it if closed
      channelHub.Channel({
        'global': {
          onMessage: (msg) => {
            console.log('Message from go:', msg)

            switch (msg.type) {
              case 'fileInfo':
                handleFileInfo(msg)
                break
              case 'account':
                handleAccount(msg)
                break
              case 'connectionStatus':
                commit('SET_CONNECTION_STATUS', msg.data.status)
                break
              case 'fileUpload':
              case 'fileDownload':
                handleFileUploadAndDownload(msg)
                break
              case 'tx':
                handleTransaction(msg)
                break
              case 'storageProvider':
                handleStorageProvider(msg)
                break
              case 'addressBook':
                handleAddressBook(msg)
                break
              case 'add_signing_request':
                commit('ADD_SIGNING_REQUEST', msg.data)
                break
              case 'remove_signing_request':
                commit('REMOVE_NOTIFICATION', msg.data)
                break
              case 'notification':
              case 'workflow_request':
                handleNotification(msg)
                break
              case 'share_process':
                handleShareProcess(msg)
                break
              case 'session-timeout':
                if (channelHub) {
                  channelHub.close()
                }
                dispatch('NOTIFY_WALLET_LOCK')
                break
              default:
                console.log('message of type ' + msg.type + ' is not handled')
                break
            }
          }
        }
      })
    }
    const handleFileInfo = (msg) => {
      const activeCat = rootState.file.activeCategory
      const myFiles = activeCat === 'my-files'
      const shared = activeCat === 'shared-with-me'
      const signed = activeCat === 'signed-by-me'
      const expiredFiles = activeCat === 'expired-files'
      const searchTerm = rootState.file.searchTerm
      const grpId = `${myFiles}-${shared}-${signed}-${expiredFiles}-${searchTerm}`

      // discard files that do not correspond to current locally queried files
      if (grpId !== msg.grpID) {
        return
      }

      commit('ADD_FILE', {
        id: msg.data.id,
        filename: msg.data.filename,
        fileKind: msg.data.fileKind,
        hasThumbnail: msg.data.hasThumbnail,
        owner: msg.data.owner,
        removed: msg.data.removed,
        expiry: new Date(msg.data.expiry * 1000),
        graceSeconds: msg.data.graceSeconds,
        expired: msg.data.expired,
        inGracePeriod: msg.data.inGracePeriod,
        aboutToExpire: msg.data.aboutToExpire,
        fileType: msg.data.fileType,
        definedSigners: msg.data.definedSigners,
        signers: msg.data.signers,
        scOrder: msg.data.scOrder,
        undefinedSigners: msg.data.undefinedSigners,
        undefinedSignersLeft: msg.data.undefinedSignersLeft,
        signatureStatus: msg.data.signatureStatus,
        loaded: true,
        readAccess: msg.data.readAccess,
        sentSignRequestsFileUndefinedSigners: msg.data.sentSignRequestsFileUndefinedSigners
      })
    }

    const handleAccount = (msg) => {
      commit('ADD_ACCOUNT_INFO', msg.data)
    }

    const handleFileUploadAndDownload = (msg) => {
      commit('UPDATE_TX_QUEUE_WITH_TRANSACTION', msg.data)
    }

    const handleTransaction = (msg) => {
      const tx = msg.data

      switch (tx.name) {
        case 'register':
          switch (tx.status) {
            case 'error':
              Vue.prototype.$showNotification(
                'fileJS.transaction_queue.notify.error_title',
                'fileJS.upload.couldNotUpload',
                'error'
              )
              break
          }
          break
        case 'remove':
          switch (tx.status) {
            case 'error':
              Vue.prototype.$showNotification(
                'fileJS.transaction_queue.notify.error_title',
                'fileJS.transaction_queue.remove_file.error',
                'error',
                // TODO: Use filename from tx obj
                { text: { filename: tx.fileName } }
              )
              break
          }
          break
        case 'sign':
          switch (tx.status) {
            case 'success':
              commit('REMOVE_NOTIFICATION', tx)
              commit('REMOVE_TRANSACTION_FROM_QUEUE', tx)
              commit('REMOVE_NOTIFICATION_ACTION_IN_PROGRESS',
                { 'fileHash': tx.hash, 'type': 'signing_request' })
              break
            case 'fail':
              Vue.prototype.$showNotification(
                'fileJS.transaction_queue.notify.error_title',
                'fileJS.transaction_queue.sign_file.error',
                'error',
                { text: { filename: tx.fileName } }
              )
              commit('REMOVE_TRANSACTION_FROM_QUEUE', tx)
              commit('REMOVE_NOTIFICATION_ACTION_IN_PROGRESS',
                { 'fileHash': tx.hash, 'type': 'signing_request' })

              break
            default:
              break
          }
          break
        case 'requestSign':
          switch (tx.status) {
            case 'success':
              commit('REMOVE_TRANSACTION_FROM_QUEUE', tx)
              break
            case 'fail':
              Vue.prototype.$showNotification(
                'fileJS.transaction_queue.notify.error_title',
                'fileJS.transaction_queue.sign_request.error',
                'error',
                { text: { filename: tx.fileName } }
              )
              commit('REMOVE_TRANSACTION_FROM_QUEUE', tx)
              break
            default:
              break
          }
          break
        case 'revoke':
          break
        default:
          break
      }
      commit('UPDATE_TX_QUEUE_WITH_TRANSACTION', tx)
    }

    const handleStorageProvider = (msg) => {
      commit('ADD_STORAGE_PROVIDER', msg.data)
    }

    const handleAddressBook = (msg) => {
      commit('ADD_ADDRESS', msg.data)
    }

    const handleNotification = (msg) => {
      commit('ADD_NOTIFICATION', msg.data)
    }
    const handleShareProcess = (msg) => {
      commit('SHARE_PROCESS', msg.data)
    }
  },
  CLOSE_CHANNEL_HUB (state) {
    if (channelHub !== undefined) {
      channelHub.close()
    }
  }
}

export default {
  mutations, getters, actions
}
