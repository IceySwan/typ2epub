#import "../template.typ": *

= 附录：规范参考与配置手册 <sec-appendix>

本附录整理了 *Typst to EPUB* 转换器涉及的核心技术标准、CLI 构建参数速查表以及样式系统 CSS 变量，供深度定制与排版参考。

== 国际数字出版标准合规性清单

转换器输出产物严格遵从 W3C / IDPF EPUB 3.3 与 EPUB 2.0.1 规范标准：

#table(
  columns: (1.2fr, 1fr, 2fr),
  align: (left, center, left),
  table.header([*规范维度*], [*适用标准*], [*实现机制说明*]),
  [文档包定义], [EPUB 3.3 / OPF], [声明 `unique-identifier`、`dcterms:modified` 及 RFC 4122 v5 确定性 UUID],
  [导航地标 (Landmarks)], [EPUB 3 / Nav], [在 `nav.xhtml` 生成语义化 `<nav epub:type="landmarks">`（封面/目录/正文）],
  [向下兼容导航 (NCX)], [EPUB 2 / NCX], [生成带有严格单调递增 `playOrder` 的 `toc.ncx` 树状目录],
  [归档包结构], [ZIP / OCF], [`mimetype` 首位无压缩（`STORED`）存储，其余文件字典序 `Deflate` 压缩],
  [数学公式支持], [MathML 3.0], [自动检测 `<math>` 标签并在 Manifest 标记 `properties="mathml"`],
  [分章独立尾注], [EPUB 3 Footnotes], [将全局注脚按章节分离并赋予 `role="doc-endnotes"` 与双向回跳锚点#footnote[这是附录中的独立尾注测试，用于验证附录分部的注脚回跳功能。]],
)

== CLI 构建参数速查

在终端中执行转换脚本时，可通过以下参数控制构建流程：

```bash
# 基础构建（输出至 build/<title>-v<version>.epub）
go run ./scripts -i main.typ

# 指定输出路径与自定义封面
go run ./scripts -i main.typ -o build/my_manual.epub -c assets/cover.png

# 调试模式：构建完成后保留临时 HTML AST 与 OEBPS 目录
go run ./scripts -i main.typ --keep-temp
```

== 样式表 CSS 变量速查

所有视觉样式与主题配色均通过 `scripts/epub-style.css` 中的 CSS 变量统一调度。如需自定义主题色或适配阅读器深色模式，可直接覆盖以下变量：

```css
:root {
  /* 基础品牌色与文字 */
  --primary-color: #3c71b7;   /* 主题蓝：标题与重点强调 */
  --link-color:    #1f5f99;   /* 链接文字颜色 */
  --text-color:    #2c3e50;   /* 正文字体主色 */
  --bg-color:      #ffffff;   /* 页面主背景色 */

  /* 定理与语义卡片背景与边框 */
  --thm-fill:      #eef6ff;   /* 定理卡片浅蓝底色 */
  --thm-stroke:    #7aa7d9;   /* 定理卡片左侧边框色 */
  --def-fill:      #f2fbf4;   /* 定义卡片浅绿底色 */
  --def-stroke:    #73ad7b;   /* 定义卡片边框色 */
  --prob-fill:     #fff8e8;   /* 问题卡片暖黄底色 */
  --prob-stroke:   #c89b3c;   /* 问题卡片边框色 */
}
```

== 附录结语

您可以通过扩展 `Contents/` 下的文件，轻松将本模板应用于学术论文合集、小说文库或技术开发手册的电子书出版。
