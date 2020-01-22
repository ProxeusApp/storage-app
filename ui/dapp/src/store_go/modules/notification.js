import Vue from 'vue'
import moment from 'moment'
import unionBy from 'lodash/fp/unionBy'
import axios from 'axios'

const getDefaultState = () => {
  return {
    notifications: [],
    notificationTranslations: {
      'signing_request': {
        'name': 'filebrowser.notifications.notification_signing_request',
        'desc': 'filebrowser.notifications.notification_signing_request_desc'
      },
      'signing_request_removed': {
        'name': 'filebrowser.notifications.notification_signing_request_removed',
        'desc': 'filebrowser.notifications.notification_signing_request_removed_desc'
      },
      'share_request': {
        'name': 'filebrowser.notifications.notification_share_request',
        'desc': 'filebrowser.notifications.notification_share_request_desc'
      },
      'workflow_request': {
        'name': 'filebrowser.notifications.notification_workflow_request',
        'desc': 'filebrowser.notifications.notification_workflow_request_desc'
      },
      'tx_xes_approval': {
        'name': 'filebrowser.notifications.notification_tx_xes_approve',
        'desc': 'filebrowser.notifications.notification_tx_xes_approve_desc'
      },
      'tx_register': {
        'name': 'filebrowser.notifications.notification_tx_register',
        'desc': 'filebrowser.notifications.notification_tx_register_desc'
      },
      'tx_remove': {
        'name': 'filebrowser.notifications.notification_tx_remove',
        'desc': 'filebrowser.notifications.notification_tx_remove_desc'
      },
      'tx_share': {
        'name': 'filebrowser.notifications.notification_tx_share',
        'desc': 'filebrowser.notifications.notification_tx_share_desc'
      },
      'tx_sign': {
        'name': 'filebrowser.notifications.notification_tx_sign',
        'desc': 'filebrowser.notifications.notification_tx_sign_desc'
      },
      'ev_notifysign': {
        'name': 'filebrowser.notifications.notification_ev_notifysign',
        'desc': 'filebrowser.notifications.notification_ev_notifysign_desc'
      },
      'tx_signRequest': {
        'name': 'filebrowser.notifications.notification_tx_signRequest',
        'desc': 'filebrowser.notifications.notification_tx_signRequest_desc'
      },
      'tx_revoke': {
        'name': 'filebrowser.notifications.notification_tx_revoke',
        'desc': 'filebrowser.notifications.notification_tx_revoke_desc'
      },
      'tx_success': {
        'name': 'filebrowser.notifications.notification_tx_success',
        'desc': 'filebrowser.notifications.notification_tx_success'
      },
      'file_about_to_expire': {
        'name': 'filebrowser.notifications.notification_file_about_to_expire',
        'desc': 'filebrowser.notifications.notification_file_about_to_expire_desc'
      },
      'file_grace_period': {
        'name': 'filebrowser.notifications.notification_file_grace_period',
        'desc': 'filebrowser.notifications.notification_file_grace_period_desc'
      },
      'file_expired': {
        'name': 'filebrowser.notifications.notification_file_expired',
        'desc': 'filebrowser.notifications.notification_file_expired_desc'
      },
      'tx_xes_send': {
        'name': 'filebrowser.notifications.notification_tx_xes_send',
        'desc': 'filebrowser.notifications.notification_tx_xes_send_desc'
      },
      'tx_xes_receive': {
        'name': 'filebrowser.notifications.notification_tx_xes_receive',
        'desc': 'filebrowser.notifications.notification_tx_xes_receive_desc'
      },
      'tx_eth_increase': {
        'name': 'filebrowser.notifications.notification_tx_eth_increase',
        'desc': 'filebrowser.notifications.notification_tx_eth_increase_desc'
      },
      'tx_eth_decrease': {
        'name': 'filebrowser.notifications.notification_tx_eth_decrease',
        'desc': 'filebrowser.notifications.notification_tx_eth_decrease_desc'
      }
    },

    insufficientGasEstimationModal: false,
    insufficientXesModal: false,
    insufficientXesAllowanceModal: false,

    showPending: false,
    searchTerm: '',
    signingRequests: []
  }
}

const state = getDefaultState()

const mutations = {
  RESET_NOTIFICATION_STATE (state) {
    Object.assign(state, getDefaultState())
  },
  ADD_NOTIFICATION (state, notification) {
    const n = state.notifications.find(
      s => s.id === notification.id)
    const index = state.notifications.indexOf(n)
    if (index === -1 && !notification.dismissed) {
      state.notifications.push(notification)
    } else if (n.actionInProgress === undefined) {
      Vue.set(state.notifications, index, notification)
    }
  },
  ADD_SIGNING_REQUEST (state, requestSign) {
    const sign = state.signingRequests.find(
      s => s.fileHash === requestSign.fileHash)
    const index = state.signingRequests.indexOf(sign)

    Vue.set(requestSign, 'inProgress', false)

    if (index === -1) {
      state.signingRequests.push(requestSign)
    } else if (sign.inProgress === false) {
      Vue.set(state.signingRequests, index, requestSign)
    }
  },
  REMOVE_NOTIFICATION (state, notification) {
    state.notifications = state.notifications.filter(
      s => s.id !== notification.id)
  },
  SET_NOTIFICATIONS_MARK_ALL_READ (state) {
    state.notifications.filter(s => s.unread !== 0).forEach(n => {
      n.unread = false
    })
  },

  ADD_NOTIFICATION_ACTION_IN_PROGRESS (state, { notification, action }) {
    const res = state.notifications.filter(
      s => s.id === notification.id)
    if (res.length !== 0) {
      Vue.set(res[0], 'actionInProgress', action)
    }
  },
  REMOVE_NOTIFICATION_ACTION_IN_PROGRESS (state, { fileHash, type }) {
    const res = state.notifications.filter(
      s => s.data.hash === fileHash && s.type === type)
    if (res.length !== 0) {
      Vue.set(res[0], 'actionInProgress', undefined)
    }
  },
  SET_NOTIFICATION_SEARCH_TERM (state, term) {
    state.searchTerm = term
  },
  SET_SHOW_PENDING (state, showPending) {
    state.showPending = showPending
  },
  SET_INSUFFICIENT_GAS_MODAL (state, showModal) {
    state.insufficientGasEstimationModal = showModal
  },
  SET_INSUFFICIENT_XES_ALLOWANCE_MODAL (state, showModal) {
    state.insufficientXesAllowanceModal = showModal
  },
  SET_INSUFFICIENT_XES_MODAL (state, showModal) {
    state.insufficientXesModal = showModal
  }
}

const getters = {
  notificationsToday: (state, getters) => ({ end = 10 }) => {
    return getters.filterNotificationByOpts
      .filter(s => isNotificationToday(s.timestamp))
      .slice(0, end)
  },
  notificationsYesterday: (state, getters) => ({ end = 10 }) => {
    return getters.filterNotificationByOpts
      .filter(s => isNotificationYesterday(s.timestamp))
      .slice(0, end)
  },
  olderNotifications: (state, getters) => ({ end = 10 }) => {
    return getters.filterNotificationByOpts
      .filter(s => isNotificationOlder(s.timestamp))
      .slice(0, end)
  },
  sortNotifications: (state) => {
    return state.notifications.slice().sort((a, b) => {
      if (a.timestamp > b.timestamp) {
        return -1
      }
      if (a.timestamp < b.timestamp) {
        return 1
      }
      return 0
    })
  },
  filterNotificationByOpts: (state, getters) => {
    let n = getters.sortNotifications
    if (state.showPending) {
      n = n.filter(s => s.pending)
    }
    let nMatchingSearchTerm = []

    // translation keys matching searchTerm
    let transKeysMatching
    if (state.searchTerm !== undefined) {
      let quoted = false
      if ((state.searchTerm.startsWith('"') &&
        state.searchTerm.endsWith('"')) ||
        (state.searchTerm.startsWith('\'') &&
          state.searchTerm.endsWith('\''))) {
        quoted = true
      }
      nMatchingSearchTerm = n.filter(s => {
        let matched
        if (quoted) {
          const termTrim = state.searchTerm.slice(1,
            state.searchTerm.length - 1)
          matched = getters.termMatchedInNotification({ term: termTrim, n: s })
        } else {
          for (let term of state.searchTerm.split(' ')) {
            matched = getters.termMatchedInNotification({ term: term, n: s })
            if (matched) {
              break
            }
          }
        }
        return matched
      })

      transKeysMatching = getters.getTranslationKeysMatchingSearchTerm
    }

    let ntransKeysMatching = []
    if (transKeysMatching !== undefined && transKeysMatching.length > 0) {
      ntransKeysMatching = n.filter(s => {
        return transKeysMatching.indexOf(s.type) !== -1
      })
    }
    let res
    if (state.searchTerm !== undefined ||
      (transKeysMatching !== undefined && transKeysMatching.length > 0)) {
      res = unionBy(nMatchingSearchTerm, ntransKeysMatching, 'id')
    } else {
      res = n
    }
    return res
  },
  termMatchedInNotification: (state, getters) => ({ term, n }) => {
    let matched = false
    if (!containsIgnoreCase(n.id, term) ||
      !containsIgnoreCase(n.type, term)) {
      // search notification data for match
      for (let p in n.data) {
        const pVal = n.data[p]
        if (p === 'who' && pVal !== null) {
          for (let addr of pVal) {
            let name = getters.nameByAddress(addr)
            if (containsIgnoreCase(name, term)) {
              matched = true
              break
            }
          }
          if (matched) {
            break
          }
        } else if (p === 'owner') {
          let name = getters.nameByAddress(pVal)
          if (containsIgnoreCase(name, term)) {
            matched = true
            break
          }
        } else {
          if (containsIgnoreCase(pVal, term)) {
            matched = true
            break
          }
        }
      }
    } else {
      matched = true
    }
    return matched
  },
  getTranslationKeysMatchingSearchTerm: (state) => {
    let matchedTypes = []
    let regex = new RegExp(state.searchTerm, 'gi')
    Object.keys(state.notificationTranslations).forEach(t => {
      let foundMatch = false
      let tVal = state.notificationTranslations[t]
      if (translateAndMatch(regex, tVal.name)) {
        foundMatch = true
      } else {
        if (translateAndMatch(regex, tVal.desc)) {
          foundMatch = true
        }
      }
      if (foundMatch) {
        matchedTypes.push(t)
      }
    })
    return matchedTypes
  },
  pendingNotificationsCount:
    (state) => {
      return state.notifications.filter(s => s.pending).length
    },
  totalNotifications:
    (state) => {
      return state.notifications.length
    },
  notificationsTodayCount:
    (state) => {
      return state.notifications.filter(
        s => isNotificationToday(s.timestamp)).length
    },
  notificationsYesterdayCount:
    (state) => {
      return state.notifications.filter(
        s => isNotificationYesterday(s.timestamp)).length
    },
  olderNotificationsCount:
    (state) => {
      return state.notifications.filter(
        s => isNotificationOlder(s.timestamp)).length
    }
}

const actions = {
  async SET_ALL_NOTIFICATIONS_AS ({ commit, dispatch }, { unread }) {
    if (unread === false) {
      try {
        let res = await axios.put('/api/notification/markAllAsRead')
        if (res.status === 200) {
          commit('SET_NOTIFICATIONS_MARK_ALL_READ')
        }
      } catch (err) {
        console.log(err)
        dispatch('UNKNOWN_ERROR')
      }
    }
  },
  async SET_FILTERED_NOTIFICATIONS_AS (
    { getters, commit, dispatch }, { unread, pending }) {
    getters.filterNotificationByOpts.forEach(
      async n => {
        dispatch('SET_NOTIFICATION_AS',
          { notification: n, unread: unread, pending: pending })
      }
    )
  },
  async SET_NOTIFICATION_AS (
    { commit, dispatch }, { notification, unread, pending }) {
    try {
      const n = { ...notification }
      n.unread = unread !== undefined ? unread : notification.unread
      n.pending = pending !== undefined ? pending : notification.pending
      let res = await axios.put('/api/notification/update', n)
      if (res.status === 200) {
        commit('ADD_NOTIFICATION', n)
      }
    } catch (err) {
      console.log(err)
      dispatch('UNKNOWN_ERROR')
    }
  },
  async DELETE_NOTIFICATION ({ commit, dispatch }, notification) {
    try {
      let res = await axios.delete(
        '/api/notification/remove/' + notification.id)
      if (res.status === 200) {
        commit('REMOVE_NOTIFICATION', notification)
      }
    } catch (err) {
      console.log(err)
      dispatch('UNKNOWN_ERROR')
    }
  },
  // A general error on server side errors or when we don't know how to handle something anymore
  UNKNOWN_ERROR () {
    // TODO: notify it to us including logs..?
    Vue.prototype.$notify({
      title: Vue.i18n.translate('global.unknownerror', 'Unknown error'),
      text: '',
      type: 'error'
    })
  }
}

function isNotificationToday (timestamp) {
  return moment.unix(timestamp).isSame(moment(), 'd')
}

function isNotificationYesterday (timestamp) {
  const yesterday = moment().subtract(1, 'days')
  return moment.unix(timestamp).isSame(yesterday, 'd')
}

function isNotificationOlder (timestamp) {
  const yesterday = moment().subtract(1, 'days')
  return moment.unix(timestamp).isBefore(yesterday, 'd')
}

function containsIgnoreCase (t1, t2) {
  if (t1 !== undefined && t1 !== null &&
    t2 !== undefined && t2 !== null) {
    return t1.toString().toLowerCase().indexOf(t2.toLowerCase()) !== -1
  }
  return false
}

function translateAndMatch (regex, t) {
  let translated = Vue.i18n.translate(t, t,
    { 'fileName': '', 'from': '', 'xesAmount': '', 'who': '' })
  return translated.match(regex)
}

export default {
  state, mutations, getters, actions
}
