<template>
  <div class="pt-0 pb-3">
    <div class="border-bottom">
      <button class="btn btn-link py-2" @click="backToOverview">
        <span class="mdi mdi-chevron-left md-24 d-inline-block border-right pr-1"></span> <span
        class="text-muted d-inline-block pl-1">{{ $t('account.wallet.backAccountOverview', 'Back to account overview') }}</span>
      </button>
    </div>
    <div class="deposit px-3 pt-3">
      <h2 class="text-center">{{ $t('wallet.deposit.title', 'Deposit XES/ETH') }}</h2>

      <div class="wallet-box p-3 text-center">
        <h5>{{ $t('wallet.deposit.depositFromOtherWallet', 'Deposit from other wallet') }}</h5>
        <p>
          {{ $t('wallet.deposit.depositFromOtherWallet.description', 'If you already have some Ether and XES, the quickest way to get them into your wallet is by sending them to.') }}
        </p>
        <ul class="list-group">
          <li class="list-group-item">
            <a :href="`${etherscanUrl}/address/${accountAddress}`"
               target="_blank"
               onclick="window.openInBrowser(event, this);"
               v-tooltip="$t('wallet.deposit.checkWallet', 'View wallet on etherscan')">
              {{ accountAddress }} <i class="mdi mdi-launch"></i>
            </a>
          </li>
        </ul>
      </div>

      <div class="wallet-box text-center p-3 mt-3">
        <h5>{{ $t('wallet.deposit.buyXESETH.title', 'Get XES/ETH')}}</h5>
        <p>{{ $t('wallet.deposit.buyXESETH.description', 'If you don’t have any XES or Ether yet, click the links below for instructions:') }}</p>

        <a href="https://testfaucet.proxeus.com/" class="btn btn-primary" onclick="window.openInBrowser(event, this);" target="_blank">{{ $t('wallet.deposit.xesFaucet', 'XES Faucet')}}
          <i class="mdi mdi-launch"></i>
        </a>
        <a href="http://faucet.ropsten.be" class="ml-3 btn btn-primary" onclick="window.openInBrowser(event, this);" target="_blank">{{ $t('wallet.deposit.ethFaucet', 'Ether Faucet')}}
          <i class="mdi mdi-launch"></i>
        </a>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'deposit',
  computed: {
    accountAddress () {
      return this.$store.getters.currentAccount.address
    },
    etherscanUrl () {
      return this.$store.getters.etherscanUrl
    }
  },
  methods: {
    backToOverview () {
      this.$store.commit('UPDATE_ACTIVE_WALLET_PANE', 'overview')
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .wallet-box {
    border-radius: $border-radius;
    background: $light;
  }
</style>
