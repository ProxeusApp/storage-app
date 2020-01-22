<template>
  <div class="card costs">
    <ul class="list-group list-group-flush">
      <li v-if="storageProviders" class="list-group-item storage">
        <h5 class="mb-1">{{ $t( 'filebrowser.costs.storagePrice', 'Storage price') }}</h5>

        <div v-for="storageProvider in storageProviders" :key="storageProvider.address" class="row small storage-provider">
          <div class="col-5 font-weight-bold">{{ storageProvider.name }}</div>
          <div class="col-7">{{ storageProvider.priceTotal | weiToXes }} XES</div>
        </div>
      </li>

      <li v-if="gasEstimate === false" class="list-group-item list-group-item-warning transaction">
        {{ $t( 'filebrowser.costs.gasEstimate.estimationFailed', 'Gas estimation failed') }}
      </li>
      <li v-else-if="gasEstimate !== undefined && gasEstimate.gasLimit === 0" class="list-group-item list-group-item-light transaction">
        {{ $t( 'filebrowser.costs.gasEstimate.free', 'No transaction costs') }}
      </li>
      <li v-else-if="gasEstimate !== undefined" class="list-group-item transaction">
        <h5 class="mb-1">{{ $t( 'filebrowser.costs.gasEstimate', 'Gas estimate') }}</h5>
        <div class="row small max-gas-price">
          <div class="col-5 font-weight-bold">{{ $t( 'filebrowser.costs.gasEstimate.maxGasPrice', 'Max gas price') }}</div>
          <div class="col-7 font-weight-bold">{{ gasEstimate.gasLimit * gasEstimate.gasPrice | weiToEth }} ETH</div>
        </div>
        <div class="row small gas-price">
          <div class="col-5">{{ $t( 'filebrowser.costs.gasEstimate.gasPrice', 'Gas price') }}</div>
          <div class="col-7">{{ gasEstimate.gasPrice | weiToEth }} ETH / {{ $t( 'filebrowser.costs.unit', 'Unit') }}</div>
        </div>
        <div class="row small gas-limit">
          <div class="col-5">{{ $t( 'filebrowser.costs.gasEstimate.gasLimit', 'Gas limit') }}</div>
          <div class="col-7">{{ gasEstimate.gasLimit }} {{ $t( 'filebrowser.costs.units', 'Units') }}</div>
        </div>
      </li>
    </ul>
  </div>
</template>

<script>
import web3Utils from 'web3-utils'

export default {
  name: 'costs',
  props: {
    storageProviders: Array,
    gasEstimate: [Object, Boolean]
  },
  filters: {
    weiToXes: function (val) {
      if (val === undefined) {
        return 0
      }

      let xes = web3Utils.fromWei(val.toString())
      return parseFloat(xes).toFixed(4)
    },
    weiToEth: function (val) {
      if (val === undefined) {
        return 0
      }

      return web3Utils.fromWei(val.toString())
    }
  }
}
</script>

<style lang="scss" scoped>
  .list-group-item {
    &:hover,
    &:focus {
      z-index: 0; // Prevent overlapping of multiselect__content-wrapper
    }
  }
</style>
