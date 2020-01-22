<script>
import { mapState } from 'vuex'

export default {
  name: 'base-modal',

  computed: {
    ...mapState({
      logoutInProgress: state => state.wallet.logoutInProgress
    })
  },
  watch: {
    logoutInProgress: function (isLogoutInProgress) {
      if (isLogoutInProgress === true) {
        this.$emit('modalClosed')
      }
    }
  }
}
</script>
