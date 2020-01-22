<template>
  <div class="spinner" v-bind="$attrs" v-bind:class="[cls ? cls: '']"
       :style="{'min-height': (margin*2 || 100) + 'px', background, color}">
    <div class="sk-circle" :style="{top: margin/2 || 25 + 'px', margin: margin + 'px auto'}">
      <div class="sk-circle1 sk-child"></div>
      <div class="sk-circle2 sk-child"></div>
      <div class="sk-circle3 sk-child"></div>
      <div class="sk-circle4 sk-child"></div>
      <div class="sk-circle5 sk-child"></div>
      <div class="sk-circle6 sk-child"></div>
      <div class="sk-circle7 sk-child"></div>
      <div class="sk-circle8 sk-child"></div>
      <div class="sk-circle9 sk-child"></div>
      <div class="sk-circle10 sk-child"></div>
      <div class="sk-circle11 sk-child"></div>
      <div class="sk-circle12 sk-child"></div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'spinner',
  props: {
    cls: String,
    margin: {
      type: Number,
      default: 100
    },
    background: {
      type: String,
      default: '#333'
    },
    color: {
      type: String,
      default: '#fff'
    }
  }
}
</script>

<style lang="scss" scoped>
  .spinner {
    position: absolute;
    bottom: 0;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    min-height: 100px;
    z-index: 2;

    &.position-sticky {
      padding-top: 50px;
    }
  }

  .sk-circle {
    margin: 100px auto;
    width: 40px;
    height: 40px;
    position: relative;

    .sk-child {
      width: 100%;
      height: 100%;
      position: absolute;
      left: 0;
      top: 0;

      &:before {
        content: '';
        display: block;
        margin: 0 auto;
        width: 15%;
        height: 15%;
        background-color: #aaaaaa;
        border-radius: 100%;
        animation: sk-circleBounceDelay 1.2s infinite ease-in-out both;
      }
    }

    .sk-circle2 {
      transform: rotate(30deg);

      &:before {
        animation-delay: -1.1s;
      }
    }

    .sk-circle3 {
      transform: rotate(60deg);

      &:before {
        animation-delay: -1s;
      }
    }

    .sk-circle4 {
      transform: rotate(90deg);

      &:before {
        animation-delay: -0.9s;
      }
    }

    .sk-circle5 {
      transform: rotate(120deg);

      &:before {
        animation-delay: -0.8s;
      }
    }

    .sk-circle6 {
      transform: rotate(150deg);

      &:before {
        animation-delay: -0.7s;
      }
    }

    .sk-circle7 {
      transform: rotate(180deg);

      &:before {
        animation-delay: -0.6s;
      }
    }

    .sk-circle8 {
      transform: rotate(210deg);

      &:before {
        animation-delay: -0.5s;
      }
    }

    .sk-circle9 {
      transform: rotate(240deg);

      &:before {
        animation-delay: -0.4s;
      }
    }

    .sk-circle10 {
      transform: rotate(270deg);

      &:before {
        animation-delay: -0.3s;
      }
    }

    .sk-circle11 {
      transform: rotate(300deg);

      &:before {
        animation-delay: -0.2s;
      }
    }

    .sk-circle12 {
      transform: rotate(330deg);

      &:before {
        animation-delay: -0.1s;
      }
    }
  }

  .small-spinner {
    min-height: 0 !important;

    .sk-circle {
      margin: 0 !important;
      height: 40px;
    }
  }

  @keyframes sk-circleBounceDelay {
    0%,
    80%,
    100% {
      transform: scale(0);
    }

    40% {
      transform: scale(1);
    }
  }
</style>
