import Vue from 'vue'
import axios from 'axios'

const getDefaultState = () => {
  return {
    addresses: [],
    activeStorageProviderDetail: undefined,
    storageProviderView: 'list',
    addressToRemove: undefined,
    doNotShowContactRemoveWarning: false,
    removeContactWarningModal: false
  }
}

const state = getDefaultState()

const mutations = {
  SET_STORAGE_PROVIDER_VIEW (state, { view }) {
    state.storageProviderView = view
  },
  SET_ACTIVE_STORAGE_PROVIDER_DETAIL (state, { storageProvider }) {
    state.activeStorageProviderDetail = storageProvider
  },
  RESET_ADDRESS_STATE (state) {
    Object.assign(state, getDefaultState())
  },
  SET_ADDRESSES (state, { addresses, myself }) {
    // Move myself to first position
    addresses.forEach((a, i) => {
      if (myself.address === a.address) {
        addresses.splice(i, 1)
        addresses.unshift(a)
      }
    })
    state.addresses = addresses
  },
  ADD_ADDRESS (state, { name, address, pgpPublicKey }) {
    state.addresses.unshift({ name, address, pgpPublicKey })
  },
  UPDATE_ADDRESS (
    state, { address, name = undefined, pgpPublicKey = undefined }) {
    let index = state.addresses.indexOf(address)
    if (index > -1) {
      if (name !== undefined) {
        Vue.set(state.addresses[index], 'name', name)
      }
      if (pgpPublicKey !== undefined) {
        Vue.set(state.addresses[index], 'pgpPublicKey', pgpPublicKey)
      }
    }
  },
  REMOVE_ADDRESS (state, { address }) {
    const a = state.addresses.find(addr => addr.address === address.address)

    let index = state.addresses.indexOf(a)
    if (index > -1) {
      state.addresses.splice(index, 1)
    }
  },
  UPDATE_ADDRESS_NAME (state, { address, name }) {
    const a = state.addresses.find(addr => addr.address === address.address)
    Vue.set(state.addresses[state.addresses.indexOf(a)], 'name', name)
  },
  CLEAN_ADDRESS_STATE (state) {
    state.addresses = []
  },
  SET_ADDRESS_TO_REMOVE (state, address) {
    state.addressToRemove = address
  },
  SET_DO_NOT_SHOW_CONTACT_REMOVE_WARNING (state, doNotShowContactRemoveWarning) {
    state.doNotShowContactRemoveWarning = doNotShowContactRemoveWarning
  },
  SET_REMOVE_CONTACT_WARNING_MODAL (state, removeContactWarningModal) {
    state.removeContactWarningModal = removeContactWarningModal
  }
}

const getters = {
  myself: (state, getters, rootState) => {
    return {
      // This is just for UI things - no actions should base on this entry.
      address: rootState.wallet.currentAddress,
      name: Vue.i18n.translate('addressbook.myWallet', 'My Account'),
      myself: true
    }
  },
  isMyself: (state, getters) => (address) => {
    const myAddress = getters.myself.address
    return address.address === myAddress || address === myAddress
  },
  addresses: (state, getters) => {
    return state.addresses.map(a => {
      if (getters.isMyself(a)) {
        // Create new object to not mutate state
        let b = { ...a }
        b.name = getters.myself.name
        return b
      }
      if (a.name === '' || a.name === undefined) {
        let b = { ...a }
        b.name = b.address
        return b
      }
      return a
    }).sort((a, b) => {
      if (getters.isMyself(a)) {
        return -1
      }
      if (getters.isMyself(b)) {
        return 1
      }
    })
  },
  addressesBySearchTerm: (state, getters) => (term) => {
    if (!getters.addresses) {
      return
    }
    let lTerm = term.toLowerCase()
    return getters.addresses.filter(
      address => address.name.toLowerCase().indexOf(lTerm) !== -1 ||
        address.address.toLowerCase().indexOf(lTerm) !== -1)
  },
  publicKeyByAddress: (state, getters) => (addressToFind) => {
    let lAddress = addressToFind.toLowerCase()
    let result = getters.addresses.find(
      address => address.address.toLowerCase().indexOf(lAddress) !== -1)

    if (result) {
      return result.pgpPublicKey
    }

    return undefined
  },
  namesByAddresses: (state, getters) => (addresses) => {
    let res = []
    addresses.forEach(address => {
      res.push(getters.nameByAddress(address))
    })
    return res
  },
  nameByAddress: (state, getters) => (address) => {
    if (address === null) {
      return ''
    }
    let addr = getters.addresses.find(
      a => a.address.toLowerCase() === address.toLowerCase())
    if (addr) {
      return addr.name
    }
    if (getters.myself.address === address) {
      return getters.myself.name
    }
    let storageProviderName = getters.activeStorageProviderNameByAddress(address)
    if (storageProviderName !== '') {
      return storageProviderName
    }

    let knownAddress = getters.knownAddresses(address)
    if (knownAddress !== '') {
      return knownAddress
    }
    return address
  },
  knownAddresses: () => (ethAddress) => {
    switch (ethAddress.toLowerCase()) {
      case '0x38ba9213c70bf6fe34f70cfc5c9b26707c6c1e85':
        return Vue.i18n.translate('addressbook.contact.xesFaucet', 'XES Faucet')
    }
    return ''
  },
  addressesWithPGPKey: (state, getters) => {
    if (!getters.addresses) {
      return []
    }
    return getters.addresses.filter(
      address => address.pgpPublicKey !== ''
    )
  },
  addressesWithoutMyself: (state, getters) => {
    return state.addresses.filter(addr => !getters.isMyself(addr))
  },
  addressesToRequestSignatureForFile: (state, getters) => (file) => {
    const addresses = getters.addressesWithPGPKey
    if (file.sentSignRequestsFileUndefinedSigners === undefined || file.sentSignRequestsFileUndefinedSigners === null) {
      return addresses
    }
    return addresses.filter(addr =>
      !file.sentSignRequestsFileUndefinedSigners.find(contact => contact.address === addr.address)
    )
  },
  addressesToShareFile: (state, getters) => (file) => {
    const addresses = getters.addressesWithPGPKey
    if (file.readAccess === undefined || file.readAccess === null) {
      return addresses.filter(addr => !getters.isMyself(addr))
    }
    return addresses.filter(addr =>
      !file.readAccess.find(contact => contact.address === addr.address) &&
      !getters.isMyself(addr)
    )
  }
}

const actions = {
  async ADD_ADDRESS ({ commit, dispatch }, { name, address }) {
    try {
      let response = await axios.put('/api/contact', {
        name: name,
        address: address
      })

      if (response.status === 201 && response.data) {
        // Refresh addresses from backend
        dispatch('LOAD_ADDRS')
      }
    } catch (err) { // axios weirdly treats response statuses which arent 2XX as an error
      const response = err.response
      if (response && response.status === 409) {
        Vue.prototype.$notify({
          title: Vue.i18n.translate('addressbook.contact.alreadyexists'),
          text: '',
          type: 'error'
        })
      } else {
        dispatch('UNKNOWN_ERROR')
      }
    }
  },
  async REMOVE_ADDRESS ({ commit, dispatch }, address) {
    let response = await axios.delete(`/api/contact/${address.address}`, {
      name: name,
      address: address,
      pgpPublicKey: ''
    })

    if (response.status === 200) {
      // Refresh addresses from backend
      dispatch('LOAD_ADDRS')
    }
  },
  async LOAD_ADDRS ({ commit, getters }) {
    let response = await axios.get('/api/contacts')
    if (response.status === 200 && response.data) {
      commit('SET_ADDRESSES',
        { addresses: response.data, myself: getters.myself })
    }
  },
  async UPDATE_ADDRESS_NAME ({ commit, dispatch }, { address, name }) {
    try {
      let response = await axios.post('/api/contact', {
        name: name,
        address: address.address
      })

      if (response.status === 200) {
        // Refresh addresses from backend
        dispatch('LOAD_ADDRS')
      }
    } catch (e) {
      dispatch('UNKNOWN_ERROR')
    }
  }
}

export default {
  state, mutations, getters, actions
}
