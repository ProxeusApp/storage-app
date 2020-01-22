<template>
  <div class="password">
    <div class="Password__group">
      <input
        :type="inputType"
        ref="referanceValue"
        v-bind:value="value"
        v-on:input="emitValue($event.target.value)"
        :class="[defaultClass, disabled ? disabledClass : '']"
        :name="name"
        :id="id"
        :placeholder="placeholder"
        :disabled="disabled"
        :tabindex="inputTabindex"
      >
      <div class="Password__icons">
        <div
          v-if="toggle"
          class="Password__toggle">
          <button
            type="button"
            class="btn-clean"
            @click.prevent="togglePassword()">
            <i v-if="this.$data._showPassword" class="mdi md-20 mdi-eye-off"
               v-tooltip="$t('password.hide.tooltip', 'Hide Password')"></i>
            <i v-else class="mdi md-20 mdi-eye" v-tooltip="$t('password.show.tooltip', 'Show Password')"></i>
          </button>
        </div>
      </div>
    </div>

    <div v-if="showStrengthMeter" v-bind:class="[strengthMeterClass]">
      <div v-bind:class="[strengthMeterFillClass]" :data-score="passwordStrength"></div>
    </div>
  </div>
</template>

<script>
import zxcvbn from 'zxcvbn'
import Password from 'vue-password-strength-meter'

export default {
  extends: Password,
  props: [
    'inputTabindex'
  ],
  watch: {
    passwordStrength (score) {
      this.$emit('score', score)

      let warning = zxcvbn(this.password).feedback.warning
      if (warning === '') {
        warning = 'Please use a strong password of 7 characters or more'
      }

      const warningTranslationKey = 'password.' +
        warning.toLowerCase().replace(/-/g, '.').replace(/"/g, '').replace(/ /g, '.')
      this.$emit('warning', this.$t(warningTranslationKey, warning))
    }
  }
}
</script>

<style lang="scss">

  .password {
    margin: 0 auto;

    .Password__strength-meter:after,
    .Password__strength-meter:before {
      width: 23.5%;
    }
  }

</style>
