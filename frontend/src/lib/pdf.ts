import * as pdfjs from 'pdfjs-dist/legacy/build/pdf.mjs'
import * as pdfjsWorkerModule from 'pdfjs-dist/legacy/build/pdf.worker.min.mjs'

// MAX-вебвью не держит воркеры
;(globalThis as unknown as { pdfjsWorker: unknown }).pdfjsWorker = pdfjsWorkerModule
pdfjs.GlobalWorkerOptions.workerSrc = ''

const MAX_TEXT_CHARS = 30_000

function readFile(file: File): Promise<ArrayBuffer> {
  if (typeof file.arrayBuffer === 'function') return file.arrayBuffer()

  // Старые версии WKWebView не реализуют Blob.arrayBuffer().
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      if (reader.result instanceof ArrayBuffer) resolve(reader.result)
      else reject(new Error('Не удалось прочитать PDF'))
    }
    reader.onerror = () => reject(reader.error ?? new Error('Не удалось прочитать PDF'))
    reader.onabort = () => reject(new Error('Чтение PDF отменено'))
    reader.readAsArrayBuffer(file)
  })
}

export async function extractPdfText(file: File): Promise<string> {
  const data = await readFile(file)
  const doc = await pdfjs.getDocument({ data }).promise

  const parts: string[] = []
  let total = 0
  for (let p = 1; p <= doc.numPages; p++) {
    const page = await doc.getPage(p)
    const content = await page.getTextContent()
    const pageText = content.items
      .map(x => ('str' in x ? x.str : ''))
      .join(' ')
      .replace(/\s+/g, ' ')
    parts.push(pageText)
    total += pageText.length
    if (total > MAX_TEXT_CHARS) break
  }

  const text = parts.join('\n\n').trim()
  if (text.length < 50) {
    throw new Error('PDF без текстового слоя — заполните резюме вручную')
  }
  return text.slice(0, MAX_TEXT_CHARS)
}
