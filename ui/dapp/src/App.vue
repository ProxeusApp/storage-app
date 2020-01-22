<template>
  <div id="app" @click="activityHandler" @mousemove="activityHandler">
    <notifications classes="alert" position="top center" :duration="3000" :width="500"/>
    <transition name="component-fade" mode="out-in">
      <component v-bind:is="view"></component>
    </transition>
  </div>
</template>

<script>
import Login from './wallet/components/Auth/Login'
import translationEN from './i18n/en.js'
// import translationJA from './i18n/ja.js'
import SplashScreen from '@/components/SplashScreen'
import axios from 'axios'

export default {
  name: 'App',
  components: {
    Login,
    SplashScreen
  },
  computed: {
    authenticated () {
      return this.$store.state.wallet.authenticated
    },
    unlocked () {
      return this.$store.state.wallet.unlocked
    },
    confirmedSplashScreen () {
      return this.$store.state.confirmedSplashScreen
    },
    view () {
      if (this.confirmedSplashScreen === false) {
        return 'splash-screen'
      }
      return this.authenticated === true || this.unlocked === true ? 'router-view' : Login
    }
  },
  async created () {
    try {
      let res = await axios.get('/api/isUnlocked')
      this.$store.commit('SET_UNLOCKED', res.data)

      const blockchainNet = await axios.get('/api/config/blockchainnet')
      this.$store.commit('SET_BLOCKCHAIN_NET', blockchainNet.data)
    } catch (e) {
      console.log(e)
    }
    this.$store.dispatch('FETCH_APP_VERSION')
    setInterval(() => {
      this.$store.dispatch('FETCH_APP_VERSION')
    }, 60000 * 10)
  },
  methods: {
    activityHandler (event) {
      lastActivity = new Date()
    },
    async activityPing () {
      if ((new Date()).getTime() - lastActivity.getTime() < 1800000) {
        try {
          const response = await axios.post('/api/ping')
          if (response.status !== 200) {
            this.$store.dispatch('LOCK_WALLET')
            this.$store.dispatch('CLOSE_CHANNEL_HUB')
            this.$store.dispatch('LOAD_ACCOUNTS_AND_SET_FIRST_ACTIVE')
          }
        } catch (e) {
          console.log(e)
        }
      } else {
        this.$store.dispatch('LOCK_WALLET')
        this.$store.dispatch('CLOSE_CHANNEL_HUB')
        this.$store.dispatch('LOAD_ACCOUNTS_AND_SET_FIRST_ACTIVE')
      }
    }
  },
  async mounted () {
    setInterval(() => this.activityPing(), 120000) // 2 minutes

    // TODO use this once the backend supports json storage again
    // fetch the translations from backend
    // const {data} = await axios.get(`http://127.0.0.1:8081/api/i18n/all`)
    // this.$i18n.add('en', data);

    // fetch the translations from json file
    this.$i18n.add('en', translationEN)
    // this.$i18n.add('ja', translationJA)
  }
}

let lastActivity = new Date() // we don't want to use vuex, it would only clutter mutations view
</script>

<style lang="scss">
  @import "assets/styles/fonts.scss";
  @import "assets/styles/variables.scss";
  @import "assets/styles/sidebar.scss";
  @import "assets/styles/tooltips.scss";
  @import "assets/styles/tour.scss";
  @import "~bootstrap/scss/bootstrap";
  @import "assets/styles/vue-bootstrap.scss";

  $mdi-font-path: "~@mdi/font/fonts";

  @import "~@mdi/font/scss/materialdesignicons.scss";

  * {
    -webkit-user-drag: none;
  }

  body {
    overflow-x: hidden;
  }

  p a {
    text-decoration: underline;
  }

  h2,
  .h2 {
    font-weight: 500;
    margin-bottom: 1rem;
  }

  h3,
  .h3 {
    font-size: 1rem;
    font-weight: 600;
  }

  h5 {
    font-weight: 500;
  }

  .text-break {
    word-break: break-word !important; // IE & < Edge 18
    overflow-wrap: break-word !important;
  }

  .form-group label {
    color: $primary;
  }

  .mdi {
    /* stylelint-disable-next-line font-family-no-missing-generic-family-keyword */
    font-family: 'Material Icons';
    font-weight: normal;
    font-style: normal;
    line-height: 1;
    letter-spacing: normal;
    text-transform: none;
    display: inline-block;
    white-space: nowrap;
    word-wrap: normal;
    direction: ltr;
    -webkit-font-feature-settings: 'liga';
    -webkit-font-smoothing: antialiased;

    .btn & {
      vertical-align: middle;
    }

    &.md-18 {
      font-size: 18px;
    }

    &.md-20 {
      font-size: 20px;
    }

    &.md-24 {
      font-size: 24px;
    }

    &.md-36 {
      font-size: 36px;
    }

    &.md-48 {
      font-size: 48px;
    }
  }

  .btn-info {
    color: white !important;
  }

  .btn-light {
    color: theme-color("primary") !important;
  }

  .btn-link:hover {
    text-decoration: none !important;
  }

  .btn.btn-round {
    border-radius: 100%;
    padding: 1rem;
    line-height: 1;
  }

  .btn-link-muted {
    background: transparent;
    color: $text-muted;

    &:hover {
      color: $primary;
    }
  }

  .btn-link-light {
    color: theme-color("light");
    background: transparent;

    &:hover {
      color: white;
    }
  }

  .bg-alert-warning {
    background-color: theme-color-level("warning", $alert-bg-level) !important;
  }

  .bg-alert-danger {
    background-color: theme-color-level("danger", $alert-bg-level) !important;
  }

  .dropdown-menu {
    border: 0;
    box-shadow: $dropdown-box-shadow;

    .dropdown-item {
      border-bottom: 1px solid $gray-200;
      cursor: pointer;
    }

    .dropdown-item:last-child {
      border-bottom: 0;
    }
  }

  .badge {
    vertical-align: text-bottom;

    &.badge-num {
      width: 20px;
      line-height: 20px;
      padding: 0;
      margin-top: 8px;
      margin-left: -10px;
      position: absolute;
      display: inline-block;
      border-radius: 50%;
    }
  }

  .badge-info {
    color: white;
  }

  .btn-link:focus {
    text-decoration: none;
  }

  .component-fade-enter-active,
  .component-fade-leave-active {
    transition: all 0.3s ease;
  }

  .component-fade-enter,
  .component-fade-leave-to {
    opacity: 0;
    transform: translateY(-300px);
  }

  .modal-backdrop {
    opacity: 1;
    display: block;
    position: fixed;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.4);
    transition: all 150ms;
  }

  .modal-content {
    box-shadow: 0 40px 100px rgba(0, 0, 0, 0.35);
  }

  .modal-footer {
    border-bottom-left-radius: $border-radius;
    border-bottom-right-radius: $border-radius;
    background: $light;
  }

  .modal-dialog {
    margin-top: 2rem;

    .modal-header {
      background: $gray-200;
      padding-top: 1rem;
      padding-bottom: 1rem;
    }

    .modal-body {
      padding-top: $spacer;
    }

    .modal-body,
    .modal-header,
    .modal-footer {
      padding-left: $spacer;
      padding-right: $spacer;
    }
  }

  .notification-wrapper {
    overflow: visible !important;
  }

  .alert {
    margin-top: 1.5rem;
    padding: 1rem 2rem;
    box-shadow: 0 40px 100px rgba(0, 0, 0, 0.35);
    text-align: center;
    opacity: 0.97;

    .notification-content {
      word-wrap: break-word;
    }

    &.success {
      background-color: $info;
      border-color: transparent;
      color: white;
    }

    &.warn {
      color: white;
      background-color: #ffc04d;
      border-color: #ffa500;
    }

    &.error {
      color: white;
      background-color: red;
      border: transparent;
    }

    &.info {
      background: $info;
    }
  }
</style>
