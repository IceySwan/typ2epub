# FAQ

### 1. 编译报错 `unknown variable: xxx`

原因：`Contents/*.typ` 与 `main.typ` 作用域独立，子文件未导入模板。

解决：在每个子文件首行添加：

```typst
#import "../template.typ": *
```

### 2. 如何设置封面

按优先级：

1. 命令行指定：`go run ./scripts -c path/to/cover.png`（PNG/JPG/SVG/PDF 均可）
2. 在 `assets/` 下放置 `cover.png`、`cover.jpg` 或 `cover.pdf`，构建器自动检测
3. 以上都没有时，自动生成一张白底黑字的简洁封面（含书名、副标题、作者、版本号）

> 注：PDF 封面会原样打包进 EPUB，但部分阅读器不支持在封面页渲染 PDF，建议优先使用 PNG/JPG。

### 3. 阅读器里脚注不按章节展示

工具已按 EPUB 3 语义输出：正文为 `<sup role="doc-noteref">`，章尾为 `<section role="doc-endnotes" epub:type="footnotes">`。支持弹窗脚注的阅读器（Apple Books、Kobo、Thorium 等）会弹出浮层；其他阅读器跳转到章尾，可点击回跳。

### 4. 数学公式如何获得最佳效果

包含公式的章节已在 `content.opf` 中声明 `properties="mathml"`。Apple Books、Calibre、Thorium、微信读书等支持 MathML 渲染；`epub-style.css` 已为长公式配置横向滚动（`overflow-x: auto`）。

### 5. `Contents/` 文件命名有格式要求吗

没有。Typst 在编译期展开所有 `#include`，切章完全依据编译后的标题顺序，与文件名无关。建议使用补零序号（如 `01_introduction.typ`）方便本地文件管理器排序。

### 6. 如何混用编号与不编号章节

用分区环境控制：`#frontmatter`（不编号）、`#mainmatter`（1、1.1）、`#appendix`（A、A.1）。第一个正文章节会自动作为 EPUB 的 bodymatter 地标。

### 7. 编译时出现 `html export is under active development` 警告

Typst 官方 HTML 导出特性仍在快速演进，该警告由 Typst 编译器输出，不影响构建结果，可忽略。
