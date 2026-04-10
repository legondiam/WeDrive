function toHex(buffer) {
  return Array.from(new Uint8Array(buffer))
    .map((byte) => byte.toString(16).padStart(2, '0'))
    .join('')
}

const SAMPLE_CHUNK_SIZE = 1024 * 1024
export const CHUNK_IDENTITY_SIZE = 5 * 1024 * 1024

async function digestBlob(blob) {
  const buffer = await blob.arrayBuffer()
  const digest = await crypto.subtle.digest('SHA-256', buffer)
  return toHex(digest)
}

async function digestArrayBuffer(buffer) {
  return crypto.subtle.digest('SHA-256', buffer)
}

export async function calculateFileSHA256(file) {
  return digestBlob(file)
}

export async function calculateFileSampleSHA256(file) {
  const size = file.size
  const sampleSize = Math.min(size, SAMPLE_CHUNK_SIZE)

  const headBlob = file.slice(0, sampleSize)

  const midStart = size > SAMPLE_CHUNK_SIZE
    ? Math.floor((size - SAMPLE_CHUNK_SIZE) / 2)
    : 0
  const midBlob = file.slice(midStart, midStart + sampleSize)

  const tailStart = size > SAMPLE_CHUNK_SIZE
    ? size - SAMPLE_CHUNK_SIZE
    : 0
  const tailBlob = file.slice(tailStart, tailStart + sampleSize)

  const [headHash, midHash, tailHash] = await Promise.all([
    digestBlob(headBlob),
    digestBlob(midBlob),
    digestBlob(tailBlob),
  ])

  return {
    file_size: size,
    head_hash: headHash,
    mid_hash: midHash,
    tail_hash: tailHash,
  }
}

export async function calculateChunkIdentity(file, chunkSize = CHUNK_IDENTITY_SIZE) {
  const chunkCount = Math.ceil(file.size / chunkSize)
  const parts = []

  for (let index = 0; index < chunkCount; index += 1) {
    const start = index * chunkSize
    const end = Math.min(file.size, start + chunkSize)
    const blob = file.slice(start, end)
    const buffer = await blob.arrayBuffer()
    const digest = await digestArrayBuffer(buffer)

    parts.push({
      part_number: index + 1,
      chunk_hash: toHex(digest),
      chunk: blob,
      size: end - start,
    })
  }

  return {
    chunk_size: chunkSize,
    chunk_count: chunkCount,
    parts,
  }
}
