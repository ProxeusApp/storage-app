<template>
  <v-tour name="dapp-tour" :steps="productTourSteps" :options="options" :callbacks="productTourCallbacks">
    <template slot-scope="tour">
      <transition name="fade">
        <!-- eslint-disable -->
        <!--
        eslint disabled because this is the official implementation for this plugin:
        https://pulsar.gitbooks.io/vue-tour/customizing-the-template.html
        -->
        <v-step
          v-if="tour.currentStep === index"
          v-for="(step, index) of tour.steps"
          :key="index"
          :step="step"
          :previous-step="tour.previousStep"
          :next-step="tour.nextStep"
          :stop="tour.stop"
          :labels="tour.labels"
          :isFirst="tour.isFirst"
          :isLast="tour.isLast">
          <!-- eslint-enable -->
          <div slot="header">
            <div class="d-flex flex-row-reverse">
              <div class="mb-1">
                <span class="mr-1">{{ index + 1 }} of {{ tour.steps.length }} </span>
                <button @click="tour.stop" class="v-step__button tour-close-button">
                  <i class="icon mdi mdi-close-box-outline md-20"></i>
                </button>
              </div>
            </div>
          </div>
          <div slot="actions">
            <div class="v-step__buttons">
              <button @click="tour.stop" v-if="!tour.isLast" class="v-step__button skip_tour">
                <small>{{ $t('proxeus.tour.skip', 'Skip Tour') }}</small>
              </button>
              <button @click="tour.previousStep" v-if="!tour.isFirst"
                      class="v-step__button mr-1">{{ $t('proxeus.tour.back', 'Back') }}
              </button>
              <button @click="tour.nextStep" v-if="!tour.isLast && tour.currentStep !== 5"
                      class="v-step__button">{{ $t('proxeus.tour.next', 'Next Step') }}
              </button>
              <button @click="nextStepWallet(tour)" v-if="!tour.isLast && tour.currentStep === 5"
                      class="v-step__button">{{ $t('proxeus.tour.next', 'Next Step') }}
              </button>
              <button @click="tour.stop" v-if="tour.isLast"
                      class="v-step__button">{{ $t('proxeus.tour.finish', 'Finish') }}
              </button>
            </div>
          </div>
        </v-step>
      </transition>
    </template>
  </v-tour>
</template>

<script>
export default {
  name: 'tour',
  data () {
    return {
      options: {
        // We need this because of the ugly way we deal with the wallet modal handling
        useKeyboardNavigation: false
      },
      closeSidebarOnClose: true,
      productTourCallbacks: {
        onStart: this.start,
        onNextStep: this.next,
        onPreviousStep: this.previous,
        onStop: this.productTourStopCallback
      },
      productTourSteps: [
        {
          target: '[data-tour-step="1"]',
          content: this.$t('proxeus.tour.wallet',
            'This is your wallet. We are sending free Ropsten ETH and XES to your wallet. Please be patient before uploading your first file, as it might take a few minutes for your funds to arrive.')
        },
        {
          target: '[data-tour-step="2"]',
          content: this.$t('proxeus.tour.allowance',
            'The allowance determines how much of your XES is available to the DApp. Once your Ropsten XES arrive, you will need to set an allowance to be able to upload, share and sign files.')
        },
        {
          target: '[data-tour-step="3"]',
          content: this.$t('proxeus.tour.addressbook',
            'Add your contacts to your address book to easily share files or request signatures.')
        },
        {
          target: '[data-tour-step="4"]',
          content: this.$t('proxeus.tour.uploadfile',
            'Select files that you want to register on the blockchain, encrypt and upload to a storage provider.'),
          params: {
            placement: 'right'
          }
        },
        {
          target: '[data-tour-step="5"]',
          content: this.$t('proxeus.tour.queue',
            'Your pending actions will be displayed here. These can be registering, uploading, signing, sharing or removing files.')
        },
        {
          target: '[data-tour-step="6"]',
          content: this.$t('proxeus.tour.notification', 'This is the notification log. It includes changes of your wallet balance, information about file expiration, sharing of files and signature requests.')
        },
        // Deactivated as we have an issue with positioning of steps in the sidebar (probably because of CSS grid)
        // Also, we can't expect that the screen of the user is enough wide for the sidebar to be expanded
        // {
        //   target: '[data-tour-step="8"]',
        //   content: this.$t('proxeus.tour.handbook', 'Still have some questions? You can find the handbook here.'),
        //   params: {
        //     placement: 'bottom'
        //   }
        // },
        {
          target: '[data-tour-step="7"]',
          content: this.$t('proxeus.tour.wallet_settings',
            'Please export your keystore now and store it somewhere save. You only need to export it once and you will be able to recover the latest state of your account at any time.'),
          params: {
            placement: 'left'
          }
        }
      ]
    }
  },
  async mounted () {
    if (this.isProductTourCompleted === false) {
      setTimeout(() => this.$tours['dapp-tour'].start(), 500)
    }
  },
  computed: {
    isProductTourCompleted () {
      return this.$store.state.productTourCompleted.find(pt => pt === this.$store.state.wallet.currentAddress) !==
        undefined
    }
  },
  methods: {
    start () {
      // if sidebar is already open keep it open after tour is over
      if (this.$store.state.file.categorySidebar.toggled === false) {
        this.closeSidebarOnClose = false
        if (this.$store.state.file.categorySidebar.size !== 'wide') {
          // close it and open it again wider
          this.$store.commit('TOGGLE_CATEGORY_SIDEBAR', { toggled: true })
          this.$store.commit('TOGGLE_CATEGORY_SIDEBAR', { toggled: false, size: 'wide' })
        }
      }
    },
    productTourStopCallback () {
      let opts = { size: 'normal' }
      if (this.closeSidebarOnClose) {
        opts.toggle = true
      }
      this.$store.commit('TOGGLE_CATEGORY_SIDEBAR', opts)
      this.$store.commit('TOGGLE_CATEGORY_SIDEBAR', { size: 'normal' })
      this.$store.commit('SET_USER_PROFILE_DROPDOWN', false)
      this.$store.commit('SET_WALLET_SETTINGS', false)
      this.$store.commit('SET_WALLET_MODAL', false)
      this.$store.commit('SET_PRODUCT_TOUR_COMPLETED', this.$store.state.wallet.currentAddress)
    },
    nextStepWallet (tour) {
      this.$store.commit('SET_WALLET_MODAL', true)
      this.$store.commit('SET_WALLET_SETTINGS', true)
      setTimeout(() => tour.nextStep(), 100)
    },
    next (currentStep) {
      switch (currentStep) {
        case 5:
          this.$store.commit('SET_USER_PROFILE_DROPDOWN', true)
          break
        case 6:
          this.$store.commit('SET_USER_PROFILE_DROPDOWN', false)
          break
        default:
          break
      }
    },
    previous (currentStep) {
      switch (currentStep) {
        case 6:
          this.$store.commit('SET_USER_PROFILE_DROPDOWN', false)
          break
        default:
          break
      }
    }
  }
}
</script>
