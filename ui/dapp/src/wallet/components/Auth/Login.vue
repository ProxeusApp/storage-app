<template>
  <div class="login-container container mt-4">
    <div class="text-center">
      <proxeus-logo width="160"></proxeus-logo>
    </div>
    <div class="d-flex flex-column login-box" v-if="accounts !== undefined">
      <b-tabs ref="loginTabHolder" class="nav-justified w-100" v-model="loginTabIndex">

        <b-tab title-item-class="tab-login" :title="$t('loginfullscreen.login', 'Login')" @click="switchTab"
               ref="loginTab">
          <account-login></account-login>
        </b-tab>

        <b-tab :active="accounts.length < 1" :title="$t('loginfullscreen.button.create', 'Create')" @click="switchTab"
               refs="createTab">
          <create-wallet></create-wallet>
        </b-tab>

        <b-tab :title="$t('loginfullscreen.button.import', 'Import')" @click="switchTab" refs="importTab">
          <import-wallet></import-wallet>
        </b-tab>
      </b-tabs>

      <div class="text-center" v-if="version !== undefined">
        <small class="text-muted mr-1">{{ $t('contract.version.info', 'Contract v.') }} {{ version.contract }}</small>
        <small class="text-muted">{{ $t('build.version.info', 'Build v.') }} {{ version.build }}</small>
      </div>
    </div>
    <upgrade-version v-if="loginTourCompleted" :modal="upgradeModal" :version="version"
                     @modalClosed="upgradeModalClose"></upgrade-version>
    <login-tour></login-tour>
  </div>
</template>

<script>
import ImportWallet from './ImportWallet'
import ProxeusLogo from '@/components/ProxeusLogo'
import CreateWallet from './CreateWallet'
import AccountLogin from '@/wallet/components/Auth/AccountLogin'
import UpgradeVersion from '@/components/Modal/UpgradeVersion'
import LoginTour from '@/components/LoginTour'

export default {
  name: 'login',
  components: {
    CreateWallet,
    ImportWallet,
    ProxeusLogo,
    AccountLogin,
    UpgradeVersion,
    LoginTour
  },
  data () {
    return {
      encryptedKeystore: undefined,
      canLoadKeystoreHandler: false,
      loginError: false,
      keystore: undefined,
      upgradeModal: false
    }
  },
  watch: {
    accounts: function (accounts) {
      this.$nextTick(() => {
        if (accounts !== undefined && accounts.length > 0) {
          this.loginTabIndex = 0
          // vanilla-js workaround in order to show/hide login-tab
          if (document.getElementsByClassName('tab-login')[0]) {
            document.getElementsByClassName('tab-login')[0].style.display = 'block'
          }
        } else {
          this.loginTabIndex = 1
          if (document.getElementsByClassName('tab-login')[0]) {
            document.getElementsByClassName('tab-login')[0].style.display = 'none'
          }
        }
      })
    },
    version: function (appVersion) {
      if (appVersion && appVersion.update !== 'none' && !this.upgradeDismissed) {
        this.upgradeModal = true
      }
    }
  },
  computed: {
    authenticating: {
      get () {
        return this.$store.state.wallet.authenticating
      },
      set (flag) {
        this.$store.commit('SET_AUTHENTICATING', flag)
      }
    },
    authError: {
      get () {
        return this.$store.state.wallet.authError
      },
      set (flag) {
        this.$store.commit('SET_AUTH_ERROR', flag)
      }
    },
    creatingAccount () {
      return this.$store.state.wallet.creatingAccount
    },
    loginTabIndex: {
      get () {
        return this.$store.state.wallet.loginTabIndex
      },
      set (index) {
        this.$store.commit('SET_LOGIN_TAB_INDEX', index)
      }
    },
    // hasAccounts () {
    //   return this.$store.state.wallet.accounts !== undefined ? this.$store.state.wallet.accounts.length > 0 : null
    // },
    accounts () {
      return this.$store.state.wallet.accounts
    },
    loginTourCompleted () {
      return this.$store.state.loginTourCompleted
    },
    version () {
      return this.$store.state.version
    },
    upgradeDismissed () {
      return this.$store.state.upgradeDismissed
    }
  },
  created () {
    this.initalizeAccounts()
  },
  methods: {
    async initalizeAccounts () {
      try {
        await this.$store.dispatch('LOAD_ACCOUNTS_AND_SET_FIRST_ACTIVE')
      } catch (e) {
        console.error('accounts could not be loaded')
        console.error(e)
      }
    },
    switchTab () {
      this.authError = ''
    },
    setAccount (account) {
      this.currentAddress = account ? account.address : null
    },
    upgradeModalClose () {
      this.upgradeModal = false
      this.$store.commit('UPGRADE_DISMISSED')
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../../assets/styles/variables";

  .login-container {
    max-width: 600px;
  }

  .login-box {
    margin-top: 2.5rem;
    border-radius: $border-radius;
    //border: 1px solid $gray-300;

    /deep/ .tab-content {
      padding: 3rem;
      margin-bottom: 2.5rem;
      border-left: 1px solid $gray-300;
      border-bottom: 1px solid $gray-300;
      border-right: 1px solid $gray-300;

      .login-text-hint {
        text-align: center;
        margin-bottom: 2.25rem;
        color: $primary;
      }
    }
  }

  /deep/ .nav-tabs {
    //background-color: $gray-300;

    .nav-item {
      .nav-link {
        border-radius: $border-radius $border-radius 0 0;

        //margin: .5rem .5rem 0 .5rem;
        color: $primary;
        cursor: pointer;
        border-bottom: 1px solid $gray-300;
        padding: 0.8rem 0.5rem;

        &.active {
          color: $gray-700;
          border: 1px solid $gray-300;
          background: $white;
          border-bottom: 1px solid white;
        }
      }

      &:first-child {
        .nav-link {
          margin-left: 0;
        }
      }

      &:last-child {
        .nav-link {
          margin-right: 0;
        }
      }
    }

    .nav-link-create {
      font-weight: bold;
    }

    .active {
      background: $gray-300;
    }
  }

  .col-account-nav {
    width: 250px;
    border-top-left-radius: $border-radius;
    border-bottom-left-radius: $border-radius;
    background: $gray-300;
  }

  .col-main {
    border-top-right-radius: $border-radius;
    border-bottom-right-radius: $border-radius;
    overflow-y: auto;
  }

  .col-account-nav,
  .col-main {
    background: $light;
    //height: 390px;
  }

  .wallet-new {
    transition: all 200ms;
    border-radius: $border-radius;
  }

  .wallet-login {
    background: $light;
    border-radius: $border-radius;
  }

  .tab-title-bar {
    height: 55px;
    vertical-align: middle;
    background: lighten($gray-300, 4%);
    border-top-right-radius: $border-radius;
  }

  .tab-container {
    border-bottom-left-radius: $border-radius;
    border-bottom-right-radius: $border-radius;
    background: $light;
  }

  .tab {
    border-radius: $border-radius;
  }

  .account-hash {
    border-radius: $border-radius;
  }

  .trim {
    word-wrap: break-word;
    overflow: hidden;
    min-width: 0;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
