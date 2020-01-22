export function urlToHostname (url) {
  let parser = document.createElement('a')
  parser.href = url
  // returns 'localhost' or '127.0.0.1' if passed param it not a valid url
  if ((parser.hostname === 'localhost' && url.indexOf('localhost') === -1) ||
      (parser.hostname === '127.0.0.1' && url.indexOf('127.0.0.1') === -1)) {
    return ''
  }
  return parser.hostname
}
