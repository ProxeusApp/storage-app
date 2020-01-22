import Vue from 'vue'
import { fromWei, toWei } from 'web3-utils'
import axios from 'axios'

const getDefaultState = () => {
  return {
    accounts: undefined, // initialize with undefined so watch gets notified
    createAccountError: false,
    authenticated: false,
    logoutInProgress: false,
    authError: false,
    authenticating: false,
    approving: false,
    allowance: 0,
    connectionStatus: 'starting',
    balance: 0,
    creatingAccount: false,
    currentAddress: undefined,
    ethBalance: 0,
    password: '',
    importWalletError: undefined,
    unlocked: false,
    loginTabIndex: undefined,
    activePane: 'overview'
  }
}

const state = getDefaultState()

const mutations = {
  UPDATE_ACTIVE_WALLET_PANE (state, pane) {
    state.activePane = pane
  },
  RESET_WALLET_STATE (state) {
    Object.assign(state, getDefaultState())
  },
  ADD_ACCOUNT_INFO (state, info) {
    state.balance = fromWei(info.balance)
    state.ethBalance = fromWei(info.ethBalance)
    // Todo: Clean this up
    state.currentAddress = info.address
    state.allowance = parseFloat(fromWei(info.allowance))
  },
  SET_CONNECTION_STATUS (state, connStatus) {
    state.connectionStatus = connStatus
  },
  SET_UNLOCKED (state, unlocked) {
    state.unlocked = unlocked
  },
  SET_ACCOUNTS (state, accounts) {
    state.accounts = accounts
  },
  SET_LOGIN_TAB_INDEX (state, loginTabIndex) {
    state.loginTabIndex = loginTabIndex
  },
  SET_ALLOWANCE (state, allowance) {
    state.allowance = parseInt(allowance)
  },
  SET_BALANCE (state, balance) {
    state.balance = balance
  },
  SET_ETH_BALANCE (state, balance) {
    state.ethBalance = balance
  },
  SET_LOGOUT_IN_PROCESS (state, logoutInProgress) {
    state.logoutInProgress = logoutInProgress
  },
  SET_AUTHENTICATED (state, authenticated) {
    state.authenticated = authenticated
    state.unlocked = authenticated
    if (authenticated === true) {
      state.authError = false
    }
  },
  SET_AUTHENTICATING (state, authenticating) {
    state.authenticating = authenticating
  },
  SET_AUTH_ERROR (state, hasError) {
    state.authError = hasError
  },
  SET_CREATING_ACCOUNT (state, creating) {
    state.creatingAccount = creating
  },
  SET_APPROVING (state, approving) {
    state.approving = approving
  },
  SET_CURRENT_ADDRESS (state, currentAddress) {
    state.currentAddress = currentAddress
  },
  SET_PASSWORD (state, password) {
    state.password = password
  },
  REMOVE_ACCOUNT (state, ethAddress) {
    const accountIndex = state.accounts.findIndex(
      acc => acc.address === ethAddress)
    state.accounts.splice(accountIndex, 1)
  },
  UPDATE_ACCOUNT (state, { address, name }) {
    const account = state.accounts.find(account => account.address === address)
    if (account) {
      account.name = name
    }
  }
}

const getters = {
  currentAccount: state => {
    if (state.accounts === undefined) {
      return null
    }
    let accounts = state.accounts.filter(
      account => account.address === state.currentAddress)
    return accounts.length > 0 ? accounts[0] : null
  },
  hasXesBalance: state => {
    return parseFloat(state.balance) > 0
  },
  hasEtherBalance: state => {
    return parseFloat(state.ethBalance) > 0
  },
  hasAllowance: state => {
    return parseFloat(state.allowance) > 0
  },
  isApproving: (state, getters, rootState) => {
    return state.approving === true ||
      undefined !== rootState.file.transactionQueue.find(
        tx => tx.name === 'xes-approve'
      )
  }
}

const actions = {
  async LOAD_ACCOUNTS ({ commit, dispatch }) {
    try {
      let res = await axios.get('/api/accounts')
      commit('SET_ACCOUNTS', res.data.reverse())
    } catch (e) {
      console.log(e)
    }
  },
  async LOAD_ACCOUNTS_AND_SET_FIRST_ACTIVE ({ commit, dispatch }) {
    await dispatch('LOAD_ACCOUNTS')
    let firstAccount = this.state.wallet.accounts[0]
    if (typeof firstAccount === 'object') {
      commit('SET_CURRENT_ADDRESS', firstAccount.address)
    }
  },

  LOAD_BALANCE ({ commit }) {
    try {
      axios.get('/api/account/balance')
    } catch (e) {
      console.log(e)
    }
  },

  async APPROVE_ESTIMATE_GAS ({ commit }, xesValue) {
    try {
      return await axios.post('/api/approveXESToContract/estimateGas', xesValue)
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },

  async APPROVE ({ commit }, xesValue) {
    commit('SET_APPROVING', true)
    try {
      await axios.post('/api/approveXESToContract', xesValue)
      commit('SET_APPROVING', false)
    } catch (e) {
      commit('SET_APPROVING', false)
      return { 'status': false, 'msg': e.response.data.message }
    }
    return { 'status': true, 'msg': null }
  },

  async CREATE_ACCOUNT ({ commit, dispatch }, { name, password }) {
    commit('SET_AUTHENTICATING', true)
    commit('SET_AUTH_ERROR', '')
    commit('SET_CREATING_ACCOUNT', true)
    try {
      let res = await axios.put('/api/account', { name: name, pw: password })
      if (res.data) {
        commit('AIRDROP_HINT_OUTSTANDING', res.data)
        commit('SET_CURRENT_ADDRESS', res.data)
        commit('SET_AUTHENTICATED', true)
        dispatch('LOAD_ACCOUNTS')
      }
    } catch (e) {
      commit('SET_AUTH_ERROR', Vue.i18n.translate(
        'walletJS.importPK.privateKeyNotImportable',
        'Could not import Private Key.'))
    }
    commit('SET_CREATING_ACCOUNT', false)
    commit('SET_AUTHENTICATING', false)
  },

  async AUTHENTICATE ({ commit, dispatch }, { password, currentAddress }) {
    commit('SET_AUTHENTICATING', true)
    commit('SET_AUTH_ERROR', '')

    try {
      let res = await axios.post('/api/login', { pw: password, ethAddr: currentAddress })
      if (res.data) {
        commit('SET_CURRENT_ADDRESS', currentAddress)
        commit('SET_AUTHENTICATED', true)
      }
    } catch (e) {
      console.log(e)
      commit('SET_AUTH_ERROR', Vue.i18n.translate(
        'loginfullscreen.invalidPassword',
        'You have entered an invalid password.'))
    }
    commit('SET_AUTHENTICATING', false)
  },

  async EXPORT_KEYSTORE_TO_FILE ({ state }) {
    try {
      const wallet = await axios.get('/api/account/export')

      if (wallet.data) {
        const link = document.createElement('a')
        link.href = window.URL.createObjectURL(new Blob([wallet.data], {
          type: 'application/zip'
        }))
        link.download = state.currentAddress + '.proxeusks'
        link.click()
      }
    } catch (e) {
      console.log(e)
    }
  },

  async EXPORT_KEYSTORE_BY_ADDRESS ({ state }, { ethAddress, password }) {
    try {
      const wallet = await axios.post('/api/account/export',
        { pw: password, address: ethAddress })

      if (wallet.data) {
        const link = document.createElement('a')
        link.href = window.URL.createObjectURL(new Blob([wallet.data], {
          type: 'application/zip'
        }))
        link.download = ethAddress + '.proxeusks'
        link.click()
        return { 'status': true, 'msg': '' }
      }
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },

  async IMPORT_KEYSTORE_FROM_FILE ({ commit, dispatch }, { walletFile, password }) {
    commit('SET_AUTH_ERROR', '')
    commit('SET_AUTHENTICATING', true)
    try {
      let formData = new FormData()
      formData.append('files', walletFile)
      formData.append('password', password)

      let res = await axios.post('/api/account/import', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
      if (res && res.data) {
        commit('SET_CURRENT_ADDRESS', res.data)
        commit('SET_AUTHENTICATED', true)
        dispatch('LOAD_ACCOUNTS')
      }
    } catch (e) {
      switch (e.response.data) {
        case 'could not decrypt key with given passphrase':
          commit('SET_AUTH_ERROR', Vue.i18n.translate(
            'loginfullscreen.invalidPassword',
            'You have entered an invalid password.'))
          break

        default:
          commit('SET_AUTH_ERROR', Vue.i18n.translate(
            'walletJS.importKS.walletFileNotImportable',
            'Could not import wallet file.'))
          break
      }
    }
    commit('SET_AUTHENTICATING', false)
  },

  async IMPORT_PRIVATE_KEY ({ commit, dispatch }, { accountName, privateKey, password }) {
    commit('SET_AUTHENTICATING', true)
    commit('SET_AUTH_ERROR', '')

    try {
      let res = await axios.post('/api/account/import/eth', {
        ethPriv: privateKey,
        accountName: accountName,
        pw: password
      })

      if (res && res.data) {
        commit('SET_CURRENT_ADDRESS', res.data)
        commit('SET_AUTHENTICATED', true)
        commit('SET_AUTHENTICATING', false)
        dispatch('LOAD_ACCOUNTS')
        return true
      }
    } catch (e) {
      switch (e.response.data) {
        // todo: here the user doesn't know which account is the one that already exists, show address in error msg
        case 'account already exists':
          commit('SET_AUTH_ERROR', Vue.i18n.translate(
            'walletJS.importPK.privateKeyAlreadyExists',
            'Account with given Private Key already exists. Either login with the existing account or remove account before importing.'))
          break

        default:
          commit('SET_AUTH_ERROR', Vue.i18n.translate(
            'walletJS.importPK.privateKeyNotImportable',
            'Could not import Private Key.'))
          break
      }
      commit('SET_AUTHENTICATING', false)
      return false
    }
  },

  HANDLE_WALLET_LOCK ({ commit, dispatch }) {
    Vue.nextTick(() => {
      dispatch('RESET_MODULE_STATE')
      commit('ADD_ACCOUNT_INFO',
        { balance: '0', ethBalance: '0', address: '0', allowance: '0' })
    })
  },
  NOTIFY_WALLET_LOCK ({ commit, dispatch }) {
    commit('SET_LOGOUT_IN_PROCESS', true)

    // PCO-1105: wait until all modal (and their background) closed
    Vue.nextTick(() => {
      dispatch('RESET_MODULE_STATE')
      commit('ADD_ACCOUNT_INFO',
        { balance: '0', ethBalance: '0', address: '0', allowance: '0' })

      commit('SET_LOGOUT_IN_PROCESS', false)
    })
  },
  async LOCK_WALLET ({ dispatch, commit, state }) {
    try {
      if (state.logoutInProgress === true) {
        return
      }
      commit('SET_LOGOUT_IN_PROCESS', true)
      await dispatch('HANDLE_WALLET_LOCK')
      await axios.post('/api/logout')
      commit('SET_LOGOUT_IN_PROCESS', false)
    } catch (e) {
      console.log(e)
    }
  },

  async REMOVE_ACCOUNT ({ commit }, { ethAddress, password }) {
    try {
      let res = await axios.post('/api/account/remove',
        { pw: password, address: ethAddress })

      if (res.status === 200) {
        commit('REMOVE_ACCOUNT', ethAddress)
        return { 'status': true, 'msg': '' }
      }
    } catch (e) {
      console.log(e)
      return { 'status': false, 'msg': e.response.data }
    }
  },

  async UPDATE_ACCOUNT_NAME ({ commit }, { address, name }) {
    // commit here & assume renaming won't fail to prevent delay when updating UI
    commit('UPDATE_ACCOUNT', { address, name })
    try {
      let res = await axios.post('/api/account/' + address, {
        name
      })
      if (res.status !== 204) {
        return { 'status': false, 'msg': '' }
      }
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },

  async SEND_XES_ESTIMATE_GAS ({ commit }, { amount, address }) {
    try {
      return await axios.post('/api/sendXES/estimateGas', {
        xesAmount: toWei(amount.toString()),
        ethAddressTo: address
      })
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },

  async SEND_XES ({ commit }, { amount, address }) {
    try {
      let res = await axios.post('/api/sendXES', {
        xesAmount: toWei(amount.toString()),
        ethAddressTo: address
      })

      if (res.status === 200) {
        return { 'status': true, 'msg': '' }
      } else {
        return { 'status': false, 'msg': 'Could not process transaction.' }
      }
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },

  async SEND_ETH_ESTIMATE_GAS ({ commit }, { amount, address }) {
    try {
      return await axios.post('/api/sendETH/estimateGas', {
        ethAmount: toWei(amount.toString()),
        ethAddressTo: address
      })
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },

  async SEND_ETH ({ commit }, { amount, address }) {
    try {
      let res = await axios.post('/api/sendETH', {
        ethAmount: toWei(amount.toString()),
        ethAddressTo: address
      })

      if (res.status === 200) {
        return { 'status': true, 'msg': '' }
      } else {
        return { 'status': false, 'msg': 'Could not process transaction.' }
      }
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  }
}

export default {
  state, mutations, getters, actions
}
