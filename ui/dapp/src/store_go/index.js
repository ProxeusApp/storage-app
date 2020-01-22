import Vue from 'vue'
import Vuex from 'vuex'

import addressModule from './modules/address'
import channelhubModule from './modules/channelhub'
import fileModule from './modules/file'
import storageProviderModule from './modules/storage-provider'
import walletModule from './modules/wallet'
import notificationModule from './modules/notification'

import VuexPersist from 'vuex-persist'
import axios from 'axios'

Vue.use(Vuex)

const vuexPersist = new VuexPersist({
  key: 'vuex_go',
  storage: window.localStorage,
  strictMode: true,
  reducer: state => ({
    language: state.language,
    confirmedSplashScreen: state.confirmedSplashScreen,
    fileListViewType: state.fileListViewType,
    airdropHintOutstanding: state.airdropHintOutstanding,
    productTourCompleted: state.productTourCompleted,
    loginTourCompleted: state.loginTourCompleted,
    file: {
      doNotShowFileRemoveWarning: state.file.doNotShowFileRemoveWarning
    },
    storageProvider: {
      defaultSPAddress: state.storageProvider.defaultSPAddress
    }
  })
  // reducer: (state) => ({file: state.file}),
  // Function that passes the state and returns the state with only the objects you want to store.
  // reducer: state => state,
  // Function that passes a mutation and lets you decide if it should update the state in localStorage.
  // filter: mutation => (true)
})

// const debug = process.env.NODE_ENV !== 'production'

// ToDO: Use Namespacing: https://github.com/vuejs/vuex/issues/335

export default new Vuex.Store({
  // actions,
  // getters,
  strict: true,
  state: {
    language: 'en',
    fileListViewType: 'list',
    availableLanguages: ['en'],
    confirmedSplashScreen: false,
    productTourCompleted: [],
    loginTourCompleted: false,
    version: undefined,
    airdropHintOutstanding: [], // set to true for existing wallets, will be set to false when new createNewWalletAccount method is called
    blockchainNet: '',
    upgradeDismissed: false
  },
  modules: {
    address: addressModule,
    channelhub: channelhubModule,
    file: fileModule,
    storageProvider: storageProviderModule,
    wallet: walletModule,
    notification: notificationModule
  },
  mutations: {
    RESTORE_MUTATION: vuexPersist.RESTORE_MUTATION,
    SET_VERSION (state, version) {
      state.update = 'block'
      state.version = version
    },
    CHANGE_LANGUAGE (state, newLanguage) {
      state.language = newLanguage
    },
    CONFIRM_SPLASH_SCREEN (state) {
      state.confirmedSplashScreen = true
    },
    AIRDROP_HINT_DISPLAYED (state, account) {
      state.airdropHintOutstanding = state.airdropHintOutstanding.filter(
        a => a !== account)
    },
    SET_LOGIN_TOUR_COMPLETED (state) {
      state.loginTourCompleted = true
    },
    SET_PRODUCT_TOUR_COMPLETED (state, account) {
      state.productTourCompleted.push(account)
    },
    AIRDROP_HINT_OUTSTANDING (state, account) {
      state.airdropHintOutstanding.push(account)
    },
    SET_FILE_LIST_VIEW_TYPE (state, viewType) {
      state.fileListViewType = viewType
    },
    SET_BLOCKCHAIN_NET (state, value) {
      state.blockchainNet = value
    },
    UPGRADE_DISMISSED (state) {
      state.upgradeDismissed = true
    }
  },
  getters: {
    etherscanUrl: (state) => {
      switch (state.blockchainNet) {
        case 'ropsten':
          return 'https://ropsten.etherscan.io'
        case 'mainnet':
        default:
          return 'https://etherscan.io'
      }
    }
  },
  actions: {
    RESET_MODULE_STATE ({ commit }) {
      commit('RESET_ADDRESS_STATE')
      commit('RESET_FILE_STATE')
      commit('RESET_NOTIFICATION_STATE')
      commit('RESET_STORAGE_PROVIDER_STATE')
      commit('RESET_WALLET_STATE')
    },
    async FETCH_APP_VERSION ({ commit }) {
      try {
        const response = await axios.get('/api/versions')
        if (response.status !== 200 || !response.data) {
          console.log('Could not fetch app version')
          return
        }
        commit('SET_VERSION', response.data)
      } catch (e) {
        console.log(e)
      }
    }
  },
  plugins: [vuexPersist.plugin]
})
