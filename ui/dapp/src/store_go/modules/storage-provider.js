import Vue from 'vue'
import axios from 'axios'

const getDefaultState = () => {
  return {
    storageProviders: [],
    defaultSPAddress: {}
  }
}

const state = getDefaultState()

const mutations = {
  RESET_STORAGE_PROVIDER_STATE (state) {
    // Save defaultSPAddress
    let persistDefaultSPAddress = Object.assign({}, state.defaultSPAddress)

    Object.assign(state, getDefaultState())

    // Don't reset defaultSPAddress (restore from saved)
    state.defaultSPAddress = persistDefaultSPAddress
  },
  CLEAN_STORAGE_PROVIDER (state) {
    state.storageProviders = []
  },
  ADD_STORAGE_PROVIDER (state, storageProvider) {
    const index = state.storageProviders.findIndex(
      sp => sp.address === storageProvider.address
    )
    if (index !== -1) {
      Vue.set(state.storageProviders, index, storageProvider)
    } else {
      state.storageProviders.push(storageProvider)
    }
  },
  SET_DEFAULT_SP_ADDRESS (state, payload) {
    Vue.set(state.defaultSPAddress, payload.accountAddress, payload.spAddress)
  }
}

const getters = {
  defaultSPOrFirst: (state, getters, rootState) => () => {
    if (getters.activeStorageProviders.length === 1 || state.defaultSPAddress[rootState.wallet.currentAddress] === undefined) {
      return getters.activeStorageProviders[0]
    }
    return getters.defaultSP()
  },
  defaultSP: (state, getters, rootState) => () => {
    return getters.activeStorageProviders.find(
      sp => sp.address === state.defaultSPAddress[rootState.wallet.currentAddress]
    )
  },
  activeStorageProviders (state) {
    return state.storageProviders
  },
  activeStorageProviderNameByAddress: (state, getters) => (ethAddress) => {
    let sp = getters.activeStorageProviders.find(
      sp => sp.address === ethAddress
    )
    if (sp === undefined) {
      return ''
    }
    if (sp.name === '') {
      return Vue.i18n.translate('addressbook.storageProvider', 'Storage Provider')
    }

    return sp.name
  }
}

const actions = {
  async LOAD_STORAGE_PROVIDERS ({ commit }) {
    commit('CLEAN_STORAGE_PROVIDER')

    let response = await axios.get('/api/providers')

    if (response && response.data) {
      for (let i in response.data) {
        let spInfo = response.data[i]

        commit('ADD_STORAGE_PROVIDER',
          {
            address: spInfo.address,
            name: spInfo.name,
            description: spInfo.description,
            logoUrl: spInfo.logoUrl,
            jurisdictionCountry: spInfo.jurisdictionCountry,
            dataCenter: spInfo.dataCenter,
            termsAndConditionsUrl: (spInfo.termsAndConditionsUrl) ? spInfo.termsAndConditionsUrl : null,
            privacyPolicyUrl: (spInfo.privacyPolicyUrl) ? spInfo.privacyPolicyUrl : null,
            maxStorageDays: spInfo.maxStorageDays,
            maxFileSizeByte: spInfo.maxFileSizeByte,
            graceSeconds: spInfo.graceSeconds,
            priceByte: spInfo.priceByte,
            priceDay: spInfo.priceDay
          })
      }
    }
  },
  SET_DEFAULT_SP_ADDRESS ({ state, commit, rootState }, spAddress) {
    commit('SET_DEFAULT_SP_ADDRESS', { accountAddress: rootState.wallet.currentAddress, spAddress: spAddress })
  }
}

export default {
  state, mutations, getters, actions
}
