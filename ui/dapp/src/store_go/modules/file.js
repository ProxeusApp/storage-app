import Vue from 'vue'
import FILE_CONSTANTS from '../../lib/FileConstants'
import axios from 'axios'

const getDefaultState = () => {
  return {
    walletSettingsDropDown: false,
    walletModal: false,
    processModal: false,
    processInfo: undefined,
    userProfileDropDown: false,
    categorySidebar: { toggled: true, size: 'normal' },
    activeCategory: 'all-files',
    files: [],
    filesLoadingError: false,
    doNotShowFileRemoveWarning: false,
    filesQueue: [],
    transactionQueue: [],
    filesLoading: false,
    searchTerm: '',
    uploadModalLoading: false
  }
}

const state = getDefaultState()

const mutations = {
  RESET_FILE_STATE (state) {
    Object.assign(state, getDefaultState())
  },
  SET_WALLET_SETTINGS (state, status) {
    state.walletSettingsDropDown = status
  },
  SET_WALLET_MODAL (state, status) {
    state.walletModal = status
  },
  SET_PROCESS_MODAL (state, status) {
    state.processModal = false
  },
  SET_USER_PROFILE_DROPDOWN (state, status) {
    state.userProfileDropDown = status
  },
  SET_ACTIVE_CATEGORY (state, category) {
    state.activeCategory = category
  },
  TOGGLE_CATEGORY_SIDEBAR (state, categorySidebar) {
    state.categorySidebar = { ...state.categorySidebar, ...categorySidebar }
  },
  SET_SEARCH_TERM (state, searchTerm) {
    state.searchTerm = searchTerm
  },
  SET_DO_NOT_SHOW_FILE_REMOVE_WARNING (state, doNotShowFileRemoveWarning) {
    state.doNotShowFileRemoveWarning = doNotShowFileRemoveWarning
  },
  SET_FILES_LOADING (state, loading) {
    state.filesLoading = loading
  },
  ADD_FILE (state, file) {
    const storedFile = state.files.find(f => file.id === f.id)

    if (file.filename === undefined || file.filename === '.' ||
      file.filename === '') {
      file.filename = file.id
    }

    if (storedFile) {
      file = { ...file }
      Vue.set(state.files, state.files.indexOf(storedFile), file)
      return
    }
    state.files.push(file)
  },
  SHARE_PROCESS (state, shareinfo) {
    state.processModal = true
    state.processInfo = shareinfo
    // FIXME: This should not be a global var in window scope!!!
    window.shareprocesslink = shareinfo.link
  },
  REMOVE_FILE (state, file) {
    state.files = state.files.filter(f => f.id !== file.id)
  },
  SET_FILES_LOADING_ERROR (state, error) {
    state.filesLoadingError = error
  },
  UPDATE_TX_QUEUE_WITH_TRANSACTION (state, transaction) {
    const tx = state.transactionQueue.find(t => t.txHash === transaction.txHash)
    if (tx) {
      const foundIndex = state.transactionQueue.indexOf(tx)
      // Remove transaction on success
      if (transaction.status && transaction.status === 'success') {
        state.transactionQueue.splice(foundIndex, 1)
        return
      }

      // Replace transaction if it's already defined
      Vue.set(state.transactionQueue, foundIndex,
        { ...transaction, error: false, syncing: false })
      return
    }

    if (transaction.status && transaction.status === 'success') {
      return
    }

    // Otherwise add it to the list
    state.transactionQueue.push({ ...transaction, error: false, syncing: false })
  },
  REMOVE_TRANSACTION_FROM_QUEUE (state, transaction) {
    const txIndex = state.transactionQueue.findIndex(
      t => t.txHash === transaction.txHash)
    if (txIndex) {
      state.transactionQueue.splice(txIndex, 1)
    }
  },
  CLEAN_FILE_LIST (state) {
    state.files = []
  },
  SET_UPLOAD_MODAL_LOADING (state, loading) {
    state.uploadModalLoading = loading
  }
}

const getters = {
  filesLoading: (state, getters) => {
    if (state.filesLoading === true) {
      return true
    }
    return state.files.length > 0 && getters.loadedFiles.length === 0
  },
  processInfo: (state) => {
    return state.processInfo
  },
  filteredFiles: (state) => {
    return state.files.filter(
      file => file.removed !== true && file.loaded === true
    ).sort((a, b) => {
      if (a.scOrder > b.scOrder) {
        return -1
      }
      if (a.scOrder < b.scOrder) {
        return 1
      }
      return 0
    })
  },
  loadedFiles: (state) => {
    return state.files.filter(
      file => file.loaded === true
    )
  },
  filesBySearchTerm: (state, getters) => (term) => {
    if (term === undefined || term === '') {
      return getters.filteredFiles
    }
    let lTerm = term.toLowerCase()
    return getters.filteredFiles.filter(
      file => file.filename &&
        (file.filename.toLowerCase().indexOf(lTerm) !== -1 || file.id.toLowerCase().indexOf(lTerm) !== -1)
    )
  },
  fileByHash: (state) => (hash) => {
    return state.files.find(f => hash === f.id)
  },
  fileByHashFromQueue: (state) => (hash) => {
    if (hash === undefined || hash === '') {
      return []
    }
    return state.filesQueue.filter(
      file => file.removed !== true && file.hash === hash)
  },
  sharedWithMe: (state, getters) => (address, searchTerm) => {
    if (searchTerm === undefined || searchTerm === '') {
      return getters.filteredFiles.filter(
        file => file.owner && file.owner.address !== address)
    }
    let lTerm = searchTerm.toLowerCase()
    return getters.filteredFiles.filter(
      file => file.owner && file.owner.address !== address &&
        file.filename.toLowerCase().indexOf(lTerm) !== -1
    )
  },
  isFileSharedWithMe: (state, getters, rootState) => (file) => {
    return file.owner && file.owner.address !== rootState.wallet.currentAddress
  },
  isFileRemoving: (state) => (file) => {
    const txQueueFile = state.transactionQueue.find(tx =>
      tx.hash && tx.hash === file.id && tx.status === 'pending' && tx.name ===
      'remove')
    return txQueueFile !== undefined
  },
  isFileHashPendingInTxQueue: (state) => (fileHash) => {
    const txQueueFile = state.transactionQueue.find(tx =>
      tx.hash && tx.hash === fileHash && tx.status === 'pending')
    return txQueueFile !== undefined
  },
  isFilePendingInTxQueue: (state, getters) => (file) => {
    return getters.isFileHashPendingInTxQueue(file.id)
  },
  fileSignStatus: (state) => (file) => {
    switch (file.signatureStatus) {
      case 1:
        return FILE_CONSTANTS.NO_SIGNERS_REQUIRED
      case 2:
        return FILE_CONSTANTS.UNSIGNED
      case 3:
        return FILE_CONSTANTS.SIGNED
    }
  },
  signersList: (state) => (file) => {
    let signersList = []
    if (file.definedSigners === null || file.definedSigners === undefined) {
      return signersList
    }
    file.definedSigners.forEach(ds => {
      if (file.signers === undefined || file.signers === null ||
        file.signers.find(s => s.address === ds.address) === undefined) {
        signersList.push({
          signer: ds,
          signed: false
        })
      } else {
        signersList.push({
          signer: ds,
          signed: true
        })
      }
    })
    return signersList
  },
  missingSignersInfo: (state, getters) => (file) => {
    let missingSigners = ''
    if (file.definedSigners === null || file.definedSigners === undefined) {
      return missingSigners
    }
    file.definedSigners.forEach(ds => {
      if (file.signers === undefined || file.signers === null ||
        file.signers.find(s => s.address === ds.address) === undefined) {
        if (missingSigners === '') {
          if (ds.address === getters.myself.address) {
            missingSigners += getters.myself.name
          } else {
            missingSigners += ds.name || ds.address
          }
        } else {
          if (ds.address === getters.myself.address) {
            missingSigners += getters.myself.name
          } else {
            missingSigners += ', ' + ds.name || ds.address
          }
        }
      }
    })
    return missingSigners
  },
  readAccessList: (state) => (file) => {
    let readAccessList = []

    if (file.readAccess !== undefined && file.readAccess !== null) {
      readAccessList = file.readAccess.map(
        ra => {
          return { name: ra.name, address: ra.address }
        }
      )
    }

    return readAccessList
  },
  isUndefinedSigner: () => (file, address) => {
    if (file.sentSignRequestsFileUndefinedSigners === null) {
      return false
    }
    return file.sentSignRequestsFileUndefinedSigners.map((s) => { return s.address }).indexOf(address) !== -1
  },
  isDefinedSigner: () => (file, address) => {
    if (file.definedSigners === null) {
      return false
    }
    return file.definedSigners.map((s) => { return s.address }).indexOf(address) !== -1
  },
  readAccessButNoSignatureRequestList: (state, getters) => (file) => {
    return getters.readAccessList(file)
      .filter(contact => !getters.isUndefinedSigner(file, contact.address))
      .filter(contact => !getters.isDefinedSigner(file, contact.address))
  },
  thumbnailSrc: () => (fileHash) => {
    return '/api/file/thumb/' + fileHash
  },
  defaultThumbnailClass: () => (file) => {
    let fileExtension = file.filename ? file.filename.split('.').pop().toLowerCase() : null

    switch (fileExtension) {
      case 'pdf':
        return 'mdi-file-pdf'
      case 'rtf':
      case 'odt':
      case 'txt':
        return 'mdi-file-document'
      case 'doc':
      case 'docx':
        return 'mdi-file-word'
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'bmp':
      case 'gif':
      case 'tiff':
        return 'mdi-file-image'
      case 'avi':
      case 'mpg':
      case 'mov':
      case 'wmv':
      case 'mp4':
        return 'mdi-file-movie'
      case 'xls':
        return 'mdi-file-chart'
      default:
        return 'mdi-file'
    }
  }
}

const actions = {
  async LOAD_FILES ({ state, commit }) {
    commit('SET_FILES_LOADING', true)
    commit('SET_FILES_LOADING_ERROR', false)
    try {
      commit('CLEAN_FILE_LIST')

      let params = {
        params: { 'filter': state.searchTerm }
      }

      if (state.activeCategory === 'my-files') {
        params.params['myFiles'] = true
      }

      if (state.activeCategory === 'shared-with-me') {
        params.params['sharedWithMe'] = true
      }

      if (state.activeCategory === 'signed-by-me') {
        params.params['signedByMe'] = true
      }

      if (state.activeCategory === 'expired-files') {
        params.params['expiredFiles'] = true
      }

      // files are not in response, files get pushed through channelhub as `fileInfo`
      let response = await axios.get('/api/file/list', params)

      if (response.status !== 200) {
        commit('SET_FILES_LOADING_ERROR', true)
      }
    } catch (e) {
      commit('SET_FILES_LOADING_ERROR', true)
    }

    commit('SET_FILES_LOADING', false)
  },
  async QUOTE ({ commit, dispatch }, { formData }) {
    try {
      return await axios.post('/api/file/quote', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async UPLOAD_ESTIMATE_GAS ({ commit, dispatch }, { formData }) {
    try {
      return await axios.post('/api/file/new/estimateGas', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async UPLOAD ({ commit, dispatch }, formData) {
    try {
      await axios.post('/api/file/new', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }

    return { 'status': true, 'msg': null }
  },
  async SHARE_PROCESS ({ commit, dispatch }, formData) {
    try {
      await axios.post('/api/process/share', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
    } catch (e) {
      console.log('process/share error:', e.response.data)
      return { 'status': false, 'msg': e.response.data }
    }

    return { 'status': true, 'msg': null }
  },
  async OPEN_PROCESS ({ commit, dispatch }, filehash) {
    try {
      console.log(filehash)
      const response = await axios.get('/api/process/drop/' + filehash)
      if (response.status === 200) {
        var relocate = response.data
        console.log(relocate)
        // eslint-disable-next-line no-undef
        electron.remote.shell.openExternal(relocate)
        return true
      } else {
        return false
      }
    } catch (e) {
      Vue.prototype.$notify({
        title: 'Error',
        text: 'Failed',
        type: 'error'
      })
    }

    return { 'status': true, 'msg': null }
  },
  async DOWNLOAD_FILE ({ commit, dispatch, state }, { file, preview }) {
    try {
      let res = await axios({
        url: '/api/file/download/' + file.id,
        method: 'GET',
        responseType: 'blob'
      })
      if (preview === true) {
        return new Blob([res.data])
      }
      let link = document.createElement('a')
      link.href = window.URL.createObjectURL(new Blob([res.data]))
      link.download = file.filename
      link.click()
    } catch (e) {
      Vue.prototype.$notify({
        title: Vue.i18n.translate('fileJS.download.couldNotDownload',
          'Could not download the file from the storage provider'),
        text: '',
        type: 'error'
      })
    }
  },
  async SHARE_FILES_ESTIMATE_GAS ({ commit, dispatch }, { file, addresses }) {
    try {
      return await axios.post('/api/file/share/estimateGas/' + file.id, addresses)
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async SHARE_FILES ({ commit, dispatch }, { file, addresses }) {
    try {
      await axios.post('/api/file/share/' + file.id, addresses)
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async SEND_SIGN_REQUEST_ESTIMATE_GAS ({ commit, dispatch }, { file, addresses }) {
    try {
      return await axios.post('/api/file/sendSigningRequest/estimateGas/' + file.id, addresses)
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async SEND_SIGN_REQUEST ({ commit, dispatch }, { file, addresses }) {
    try {
      return await axios.post('/api/file/sendSigningRequest/' + file, addresses)
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async UNSHARE_FILE_ESTIMATE_GAS ({ commit, dispatch }, { file, addresses }) {
    try {
      return await axios.post('/api/file/revoke/estimateGas/' + file.id, addresses)
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async UNSHARE_FILE ({ commit, dispatch }, { file, addresses }) {
    try {
      await axios.post('/api/file/revoke/' + file.id, addresses)
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async SIGN_HASH_ESTIMATE_GAS ({ commit }, hash) {
    try {
      return await axios.get('/api/file/sign/estimateGas/' + hash)
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async SIGN_HASH ({ commit }, hash) {
    try {
      await axios.get('/api/file/sign/' + hash)
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async REMOVE_FILE_ESTIMATE_GAS ({ commit, dispatch }, file) {
    try {
      return await axios.get('/api/file/remove/estimateGas/' + file.id)
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async REMOVE_FILE ({ commit, dispatch }, file) {
    try {
      const response = await axios.post('/api/file/remove/' + file.id)

      if (response.status === 200) {
        return true
      }
    } catch (e) { console.log(e) }

    return false
  },
  async REMOVE_FILE_LOCAL ({ commit, dispatch }, file) {
    try {
      return await axios.post('/api/file/removeLocal/' + file.id)
    } catch (e) {
      return { 'status': false, 'msg': e.response.data }
    }
  },
  async REMOVE_FILE_DISK_KEEP_META ({ commit }, fileHash) {
    try {
      await axios.post('/api/file/removeDiskKeepMeta/' + fileHash)
    } catch (e) { console.log(e) }
  }
}

export default {
  state, mutations, getters, actions
}
