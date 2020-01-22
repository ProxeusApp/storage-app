import Vue from 'vue'
import Router from 'vue-router'
import FileBrowser from '../views/FileBrowser'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'FileBrowser',
      component: FileBrowser,
      props: true
    }
  ]
})
