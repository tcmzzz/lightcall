import MarkdownIt from 'markdown-it'
import DOMPurify from 'dompurify'

// Create markdown parser with basic options
const md = new MarkdownIt({
  html: false, // Disable HTML tags in source
  xhtmlOut: false, // Use '/' to close single tags
  breaks: true, // Convert line breaks to <br>
  linkify: true, // Autoconvert URL-like text to links
  typographer: true // Enable some language-neutral replacement
})

/**
 * Convert markdown text to HTML with XSS protection
 * @param {string} text - Markdown text to convert
 * @returns {string} Sanitized HTML string
 */
export function renderMarkdown(text) {
  if (!text) return ''
  const rawHtml = md.render(text)
  return DOMPurify.sanitize(rawHtml, { USE_PROFILES: { html: true } })
}

/**
 * Convert markdown text to HTML inline with XSS protection
 * @param {string} text - Markdown text to convert
 * @returns {string} Sanitized HTML string
 */
export function renderMarkdownInline(text) {
  if (!text) return ''
  const rawHtml = md.renderInline(text)
  return DOMPurify.sanitize(rawHtml, { USE_PROFILES: { html: true } })
}
