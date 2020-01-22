<template>
  <v-tour name="login-tour" :steps="steps" :callbacks="callbacks">
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
              <button @click="tour.nextStep" v-if="!tour.isLast"
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
  name: 'login-tour',
  data () {
    return {
      callbacks: {
        onStart: this.start,
        onNextStep: this.next,
        onPreviousStep: this.previous,
        onStop: this.stop
      },
      steps: [
        {
          target: '#setPasswordInputCreateWallet',
          content: this.$t('proxeus.logintour.passwordfields',
            'Set a password for your new account. Use a safe password (consider using upper- and lowercase letters, numbers and special characters). Your account address will be generated automatically.'),
          params: {
            placement: 'right'
          }
        },
        {
          target: '.import',
          content: this.$t('proxeus.logintour.importdropdown',
            'Alternatively, you can import an account via private key or wallet file.'),
          params: {
            placement: 'left'
          }
        }
      ]
    }
  },
  async mounted () {
    if (this.$store.state.loginTourCompleted === false) {
      setTimeout(() => this.$tours['login-tour'].start(), 500)
    }
  },
  computed: {},
  methods: {
    start () {
      this.$store.commit('SET_LOGIN_TAB_INDEX', 1)
    },
    next (currentStep) {
      switch (currentStep) {
        case 0:
          this.$store.commit('SET_LOGIN_TAB_INDEX', 2)
          break
        case 1:
          this.$store.commit('SET_LOGIN_TAB_INDEX', 0)
          break
        default:
          break
      }
    },
    previous (currentStep) {
      switch (currentStep) {
        case 1:
          this.$store.commit('SET_LOGIN_TAB_INDEX', 1)
          break
        case 2:
          this.$store.commit('SET_LOGIN_TAB_INDEX', 2)
          break
        default:
          break
      }
    },
    stop () {
      this.$store.commit('SET_LOGIN_TAB_INDEX', 1)
      this.$store.commit('SET_LOGIN_TOUR_COMPLETED')
    }
  }
}
</script>
