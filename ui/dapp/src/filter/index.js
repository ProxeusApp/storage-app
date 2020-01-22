import * as filter from './filters'

export default {
  install (Vue) {
    Vue.filter('urlToHostname', filter.urlToHostname)
  }
}
