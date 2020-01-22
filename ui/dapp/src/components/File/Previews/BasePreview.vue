<script>
export default {
  name: 'base-preview',
  props: ['fileBlob'],
  data () {
    return {
      src: null
    }
  },
  mounted () {
    if (this.fileBlob) {
      let t = this.fileBlob.slice(0, this.fileBlob.size, this.mime)
      this.src = URL.createObjectURL(t)
    }
  }
}
</script>
