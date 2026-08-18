// Lightweight i18n dictionary (zh/en). Kept dependency-free on purpose:
// the site has few strings and a full i18n module would be overkill.
// Add a locale by extending `Dict` and `locales` below.

export type Locale = 'zh' | 'en'

export const locales: Locale[] = ['zh', 'en']

export const localeLabels: Record<Locale, string> = {
  zh: '中文',
  en: 'English',
}

type Dict = Record<string, string>

const zh: Dict = {
  'app.title': 'Lolicount',
  'app.desc': '萌系可换肤 SVG 访问计数器,往 README 贴一行链接即可计数。',

  'nav.lang': '语言',
  'nav.top': '顶部',

  'hero.title': 'Lolicount',

  'loli.title': '立绘主题',
  'loli.desc': '该主题由多个部件组成,如服装、腮红、表情等',
  'loli.loading': '加载中…',
  'loli.rolling': '生成中…',
  'loli.reroll': '点击换一个孩子',

  'themes.title': '卡片主题',
  'themes.desc': '卡片主题每张展示只有一张图片构成,点击图片可重新加载',
  'themes.reload': '重新加载',

  'playground.title': '🎨 Playground',
  'playground.params': '参数',
  'playground.preview': '预览',
  'playground.previewPlaceholder': '选择参数后点击 Generate it! 生成预览',
  'playground.generate': 'Generate it!',
  'playground.regenerate': '重新生成',

  'embed.title': '📦 嵌入方式',
  'embed.svg': 'SVG 地址',
  'embed.img': 'Img 标签',
  'embed.markdown': 'Markdown',
  'embed.copy': '复制',
  'embed.copied': '已复制',

  'param.name': '计数器名称',
  'param.kind': '主题类型',
  'param.kindFrame': '卡片主题',
  'param.kindCharacter': '立绘主题',
  'param.theme': '主题',
  'param.fontStyle': '字体样式',
  'param.fontDefault': '默认',
  'param.mode': '帧模式 mode',
  'param.modeSeq': '顺序 seq',
  'param.modeRandom': '随机 random',
  'param.modeHint': '顺序模式随计数循环帧,随机模式每次请求随机抽帧。',
  'param.characterHint': '立绘主题固定随机模式,每次请求重新组合服装与表情。',
  'param.fsize': '字号 fsize',
  'param.scale': '图片缩放 scale',
  'param.px': '像素 x',
  'param.py': '像素 y',
  'param.rx': '比例 rx',
  'param.ry': '比例 ry',
  'param.unshowf': '隐藏字体 (unshowf)',

  'footer.credits': '感谢所有贡献者',
  'footer.poweredBy': '由 ❤ 与 Go + Nuxt 驱动',
  'footer.repo': 'GitHub 仓库',
  'footer.author': '作者',
  'footer.donate': '赞助本项目',
  'footer.donateHint': '如果 Lolicount 对你有帮助,可以请作者喝杯奶茶 🧋',
  'footer.contributors': '贡献者',
  'footer.tagline': '萌系可换肤 SVG 访问计数器',
  'footer.author': '作者',
  'footer.repo': 'GitHub 仓库',
  'footer.thanks': '致谢项目',
}

const en: Dict = {
  'app.title': 'Lolicount',
  'app.desc': 'A cute, themeable SVG visitor counter — paste one link in your README and watch it count.',

  'nav.lang': 'Language',
  'nav.top': 'Top',

  'hero.title': 'Lolicount',

  'loli.title': 'Character Themes',
  'loli.desc': 'A character theme is composed of multiple layers such as outfit, blush, and expression.',
  'loli.loading': 'Loading…',
  'loli.rolling': 'Generating…',
  'loli.reroll': 'Click to reroll another girl',

  'themes.title': 'Card Themes',
  'themes.desc': 'Each card theme shows a single image. Click the image to reload it.',
  'themes.reload': 'Reload',

  'playground.title': '🎨 Playground',
  'playground.params': 'Parameters',
  'playground.preview': 'Preview',
  'playground.previewPlaceholder': 'Pick parameters then click Generate it! to preview.',
  'playground.generate': 'Generate it!',
  'playground.regenerate': 'Regenerate',

  'embed.title': '📦 Embed Formats',
  'embed.svg': 'SVG address',
  'embed.img': 'Img tag',
  'embed.markdown': 'Markdown',
  'embed.copy': 'Copy',
  'embed.copied': 'Copied',

  'param.name': 'Counter name',
  'param.kind': 'Theme type',
  'param.kindFrame': 'Card',
  'param.kindCharacter': 'Character',
  'param.theme': 'Theme',
  'param.fontStyle': 'Font style',
  'param.fontDefault': 'Default',
  'param.mode': 'Frame mode',
  'param.modeSeq': 'Sequential',
  'param.modeRandom': 'Random',
  'param.modeHint': 'Sequential cycles frames with the count; random picks a frame per request.',
  'param.characterHint': 'Character themes are always random — outfit and expression are recomposed per request.',
  'param.fsize': 'Font size',
  'param.scale': 'Image scale',
  'param.px': 'Pixel x',
  'param.py': 'Pixel y',
  'param.rx': 'Ratio rx',
  'param.ry': 'Ratio ry',
  'param.unshowf': 'Hide font (unshowf)',

  'footer.credits': 'Thanks to all contributors',
  'footer.poweredBy': 'Powered by ❤ and Go + Nuxt',
  'footer.repo': 'GitHub repo',
  'footer.author': 'Author',
  'footer.donate': 'Sponsor this project',
  'footer.donateHint': 'If Lolicount helps you, buy the author a milk tea 🧋',
  'footer.contributors': 'Contributors',
  'footer.tagline': 'A cute, themeable SVG visitor counter',
  'footer.author': 'Author',
  'footer.repo': 'GitHub repo',
  'footer.thanks': 'Acknowledgements',
}

export const dictionaries: Record<Locale, Dict> = { zh, en }
