<template>
  <div class="account-login">
    <spinner v-if="authenticating" :margin="85" background="transparent" style="position: relative;"></spinner>
    <div v-else-if="currentAccount">
      <h5 class="login-text-hint">{{ $t('loginfullscreen.loginExisting', 'Login with your existing account') }}</h5>
      <account-entry @click.native="currentAddress = ''" :account="currentAccount"
                     @changedName="updateAccountEntryName"/>
      <form @submit.prevent="authenticate">
        <div class="d-flex flex-row">
          <div class="input-group mb-2 mt-2">
            <input id="pwEl" type="password" class="form-control" v-model="password" name="password"
                   ref="passwordInput" :placeholder="$t('loginfullscreen.password', 'Password')">
            <div class="input-group-append">
              <button type="submit" :disabled="password === ''"
                      class="btn btn-primary">{{ $t('loginfullscreen.login', 'Login') }}
              </button>
            </div>
          </div>
        </div>
        <div class="auth-error pb-3" v-if="authError">
          <span class="text-danger">{{authError}}</span>
        </div>
      </form>
    </div>
    <account-list v-if="!currentAccount" :accounts="accounts"></account-list>
  </div>
</template>

<script>
import AccountList from '@/wallet/components/Auth/AccountList'
import AccountEntry from '@/wallet/components/Auth/AccountEntry'
import Spinner from '@/components/Spinner'

export default {
  name: 'AccountLogin',

  components: {
    AccountList,
    AccountEntry,
    Spinner
  },
  watch: {
    currentAccount: function (account) {
      this.$nextTick(() => {
        if (typeof account !== 'undefined' && account !== null &&
          typeof this.$refs.passwordInput !== 'undefined') {
          this.$refs.passwordInput.focus()
        }
      })
    },

    loginTabIndex: function (index) {
      this.$nextTick(() => {
        if (index === 0 && this.currentAccount !== null &&
          typeof this.$refs.passwordInput !== 'undefined') {
          this.$refs.passwordInput.focus()
        }
      })
    }
  },
  computed: {
    currentAccount () {
      return this.$store.getters.currentAccount
    },
    accounts () {
      return this.$store.state.wallet.accounts
    },
    authError: {
      get () {
        return this.$store.state.wallet.authError
      },
      set (flag) {
        this.$store.commit('SET_AUTH_ERROR', flag)
      }
    },
    loginTabIndex () {
      return this.$store.state.wallet.loginTabIndex
    },
    authenticating: {
      get () {
        return this.$store.state.wallet.authenticating
      },
      set (flag) {
        this.$store.commit('SET_AUTHENTICATING', flag)
      }
    },
    password: {
      get () {
        return this.$store.state.wallet.password
      },
      set (password) {
        return this.$store.commit('SET_PASSWORD', password)
      }
    },
    currentAddress: {
      get () {
        return this.$store.state.wallet.currentAddress
      },
      set (address) {
        this.$store.commit('SET_CURRENT_ADDRESS', address)
      }
    }
  },
  methods: {
    async authenticate () {
      this.$store.dispatch('AUTHENTICATE', { password: this.password, currentAddress: this.currentAddress })
      this.password = ''
    },
    updateAccountEntryName ({ address, name }) {
      this.$store.dispatch('UPDATE_ACCOUNT_NAME', { address, name })
    }
  }
}
</script>
