const path = require('path')

module.exports = {
  assetsDir: 'static/',
  outputDir: path.resolve(__dirname, '../../artifacts/dapp/dist'),
  runtimeCompiler: true,
  productionSourceMap: false,
  pages: {
    app: {
      entry: 'src/main.js',
      template: 'public/view/index.html',
      filename: 'view/index.html'
    }
  },
  devServer: {
    port: 8080,
    historyApiFallback: {
      rewrites: [
        {
          from: /.*/,
          to: '/view/index.html'
        }]
    },
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        secure: false
      },
      '/pipe/*': {
        target: 'ws://localhost:8081',
        secure: false,
        ws: true,
        changeOrigin: false
      }
    }
  }
}
