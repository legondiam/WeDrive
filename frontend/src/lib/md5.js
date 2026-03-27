function add32(a, b) {
  return (a + b) >>> 0
}

function cmn(q, a, b, x, s, t) {
  return add32(((add32(add32(a, q), add32(x, t)) << s) | ((add32(add32(a, q), add32(x, t)) >>> (32 - s)))), b)
}

function ff(a, b, c, d, x, s, t) {
  return cmn((b & c) | (~b & d), a, b, x, s, t)
}

function gg(a, b, c, d, x, s, t) {
  return cmn((b & d) | (c & ~d), a, b, x, s, t)
}

function hh(a, b, c, d, x, s, t) {
  return cmn(b ^ c ^ d, a, b, x, s, t)
}

function ii(a, b, c, d, x, s, t) {
  return cmn(c ^ (b | ~d), a, b, x, s, t)
}

function md5cycle(state, block) {
  let [a, b, c, d] = state

  a = ff(a, b, c, d, block[0], 7, 0xd76aa478)
  d = ff(d, a, b, c, block[1], 12, 0xe8c7b756)
  c = ff(c, d, a, b, block[2], 17, 0x242070db)
  b = ff(b, c, d, a, block[3], 22, 0xc1bdceee)
  a = ff(a, b, c, d, block[4], 7, 0xf57c0faf)
  d = ff(d, a, b, c, block[5], 12, 0x4787c62a)
  c = ff(c, d, a, b, block[6], 17, 0xa8304613)
  b = ff(b, c, d, a, block[7], 22, 0xfd469501)
  a = ff(a, b, c, d, block[8], 7, 0x698098d8)
  d = ff(d, a, b, c, block[9], 12, 0x8b44f7af)
  c = ff(c, d, a, b, block[10], 17, 0xffff5bb1)
  b = ff(b, c, d, a, block[11], 22, 0x895cd7be)
  a = ff(a, b, c, d, block[12], 7, 0x6b901122)
  d = ff(d, a, b, c, block[13], 12, 0xfd987193)
  c = ff(c, d, a, b, block[14], 17, 0xa679438e)
  b = ff(b, c, d, a, block[15], 22, 0x49b40821)

  a = gg(a, b, c, d, block[1], 5, 0xf61e2562)
  d = gg(d, a, b, c, block[6], 9, 0xc040b340)
  c = gg(c, d, a, b, block[11], 14, 0x265e5a51)
  b = gg(b, c, d, a, block[0], 20, 0xe9b6c7aa)
  a = gg(a, b, c, d, block[5], 5, 0xd62f105d)
  d = gg(d, a, b, c, block[10], 9, 0x02441453)
  c = gg(c, d, a, b, block[15], 14, 0xd8a1e681)
  b = gg(b, c, d, a, block[4], 20, 0xe7d3fbc8)
  a = gg(a, b, c, d, block[9], 5, 0x21e1cde6)
  d = gg(d, a, b, c, block[14], 9, 0xc33707d6)
  c = gg(c, d, a, b, block[3], 14, 0xf4d50d87)
  b = gg(b, c, d, a, block[8], 20, 0x455a14ed)
  a = gg(a, b, c, d, block[13], 5, 0xa9e3e905)
  d = gg(d, a, b, c, block[2], 9, 0xfcefa3f8)
  c = gg(c, d, a, b, block[7], 14, 0x676f02d9)
  b = gg(b, c, d, a, block[12], 20, 0x8d2a4c8a)

  a = hh(a, b, c, d, block[5], 4, 0xfffa3942)
  d = hh(d, a, b, c, block[8], 11, 0x8771f681)
  c = hh(c, d, a, b, block[11], 16, 0x6d9d6122)
  b = hh(b, c, d, a, block[14], 23, 0xfde5380c)
  a = hh(a, b, c, d, block[1], 4, 0xa4beea44)
  d = hh(d, a, b, c, block[4], 11, 0x4bdecfa9)
  c = hh(c, d, a, b, block[7], 16, 0xf6bb4b60)
  b = hh(b, c, d, a, block[10], 23, 0xbebfbc70)
  a = hh(a, b, c, d, block[13], 4, 0x289b7ec6)
  d = hh(d, a, b, c, block[0], 11, 0xeaa127fa)
  c = hh(c, d, a, b, block[3], 16, 0xd4ef3085)
  b = hh(b, c, d, a, block[6], 23, 0x04881d05)
  a = hh(a, b, c, d, block[9], 4, 0xd9d4d039)
  d = hh(d, a, b, c, block[12], 11, 0xe6db99e5)
  c = hh(c, d, a, b, block[15], 16, 0x1fa27cf8)
  b = hh(b, c, d, a, block[2], 23, 0xc4ac5665)

  a = ii(a, b, c, d, block[0], 6, 0xf4292244)
  d = ii(d, a, b, c, block[7], 10, 0x432aff97)
  c = ii(c, d, a, b, block[14], 15, 0xab9423a7)
  b = ii(b, c, d, a, block[5], 21, 0xfc93a039)
  a = ii(a, b, c, d, block[12], 6, 0x655b59c3)
  d = ii(d, a, b, c, block[3], 10, 0x8f0ccc92)
  c = ii(c, d, a, b, block[10], 15, 0xffeff47d)
  b = ii(b, c, d, a, block[1], 21, 0x85845dd1)
  a = ii(a, b, c, d, block[8], 6, 0x6fa87e4f)
  d = ii(d, a, b, c, block[15], 10, 0xfe2ce6e0)
  c = ii(c, d, a, b, block[6], 15, 0xa3014314)
  b = ii(b, c, d, a, block[13], 21, 0x4e0811a1)
  a = ii(a, b, c, d, block[4], 6, 0xf7537e82)
  d = ii(d, a, b, c, block[11], 10, 0xbd3af235)
  c = ii(c, d, a, b, block[2], 15, 0x2ad7d2bb)
  b = ii(b, c, d, a, block[9], 21, 0xeb86d391)

  state[0] = add32(state[0], a)
  state[1] = add32(state[1], b)
  state[2] = add32(state[2], c)
  state[3] = add32(state[3], d)
}

function bytesToWords(bytes, offset) {
  const words = new Array(16).fill(0)
  for (let i = 0; i < 64; i += 1) {
    words[i >> 2] |= bytes[offset + i] << ((i % 4) * 8)
  }
  return words
}

function wordsToHex(state) {
  const hex = []
  for (let i = 0; i < state.length; i += 1) {
    const value = state[i]
    for (let j = 0; j < 4; j += 1) {
      const byte = (value >> (j * 8)) & 0xff
      hex.push(byte.toString(16).padStart(2, '0'))
    }
  }
  return hex.join('')
}

export async function calculateFileMD5(file) {
  const buffer = await file.arrayBuffer()
  const input = new Uint8Array(buffer)
  const originalLength = input.length
  const bitLength = originalLength * 8
  const totalLength = (((originalLength + 8) >> 6) + 1) * 64
  const bytes = new Uint8Array(totalLength)

  bytes.set(input)
  bytes[originalLength] = 0x80

  for (let i = 0; i < 8; i += 1) {
    bytes[totalLength - 8 + i] = (bitLength >>> (i * 8)) & 0xff
  }

  const state = [0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476]

  for (let offset = 0; offset < bytes.length; offset += 64) {
    md5cycle(state, bytesToWords(bytes, offset))
  }

  return wordsToHex(state)
}
