// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'

import App from './App'
import router from './router'
import store from './store_go'
import filters from './filter/index'

import axios from 'axios'
import vuexI18n from 'vuex-i18n'

import Notifications from 'vue-notification'
import VueClipboard from 'vue-clipboard2'
import VTooltip from 'v-tooltip'
import VueTour from 'vue-tour'

import { Alert, Dropdown, Modal, Popover, Progress, Tabs } from 'bootstrap-vue/es/components'

import './lib/customElectron'
import InfiniteLoading from 'vue-infinite-loading'

Vue.use(Alert)
Vue.use(Modal)
Vue.use(Dropdown)
Vue.use(Tabs)
Vue.use(Progress)
Vue.use(Popover)

Vue.use(VueTour)
Vue.use(filters)
Vue.use(Notifications)
Vue.use(VTooltip)
VueClipboard.config.autoSetContainer = true
Vue.use(VueClipboard)
Vue.use(vuexI18n.plugin, store)
Vue.use(InfiniteLoading, { /* options */ })

// set the start locale to use
Vue.i18n.set(store.state.language)

// convenience method for showing notifications to the user
Vue.prototype.$showNotification = (
  title, text, type = 'notification', transParams = {}, duration) => {
  let transParamTitle = transParams.title !== undefined ? transParams.title : {}
  let transParamText = transParams.text !== undefined ? transParams.text : {}
  let notificationDuration = duration !== undefined ? duration : null

  Vue.prototype.$notify({
    title: Vue.i18n.translate(title, '', transParamTitle),
    text: Vue.i18n.translate(text, '', transParamText),
    type: type,
    duration: notificationDuration
  })
}

Vue.config.productionTip = false

// axios setup
axios.defaults.headers.common['X-Requested-With'] = 'XMLHttpRequest'

// eslint-disable-next-line no-new
new Vue({
  el: '#app',
  router,
  store,
  components: { App },
  template: '<App/>'
})
