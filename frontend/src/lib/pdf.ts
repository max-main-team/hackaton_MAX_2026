import * as pdfjs from 'pdfjs-dist/legacy/build/pdf.mjs'
import workerUrl from 'pdfjs-dist/legacy/build/pdf.worker.min.mjs?url'

pdfjs.GlobalWorkerOptions.workerSrc = workerUrl

const MAX_TEXT_CHARS = 30_000

export async function extractPdfText(file: File): Promise<string> {
  const data = await file.arrayBuffer()
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
