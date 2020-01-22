<template>
  <div class="address-list text-left w-100 mt-2">
    <h5 class="login-text-hint">{{ $t('loginfullscreen.chooseAccount', 'Choose an account') }}</h5>

    <div class="address-list">
      <account-entry v-for="account in accounts"
                     :key="account.address"
                     v-bind:account="account"
                     @removeAccount="setRemoveAccountWarning(account.address)"
                     @changedName="updateAccountEntryName"/>
    </div>
    <forget-wallet-warning :accountToRemove="accountToRemove" :modal="forgetWalletModal"
                           @modalClosed="removeAccountWarningClosed"></forget-wallet-warning>
  </div>
</template>

<script>
import AccountEntry from '@/wallet/components/Auth/AccountEntry'
import ForgetWalletWarning from '@/components/Modal/ForgetWalletWarning'

export default {
  name: 'AccountList',
  components: {
    AccountEntry,
    ForgetWalletWarning
  },
  data () {
    return {
      forgetWalletModal: false,
      accountToRemove: undefined,
      password: undefined
    }
  },
  computed: {
    accounts () {
      return this.$store.state.wallet.accounts
    }
  },
  methods: {
    removeAccountWarningClosed () {
      this.accountToRemove = undefined
      this.forgetWalletModal = false
    },
    setRemoveAccountWarning (ethAddress) {
      this.forgetWalletModal = true
      this.accountToRemove = ethAddress
    },
    updateAccountEntryName ({ address, name }) {
      this.$store.dispatch('UPDATE_ACCOUNT_NAME', { address, name })
    }
  }
}
</script>
