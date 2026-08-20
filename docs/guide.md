# 排版与组件指南

## 分区环境

模板提供三个分区环境，控制标题编号规则：

| 环境 | 标题编号 | 用途 |
|---|---|---|
| `#frontmatter[...]` | 无 | 题记、前言、序言 |
| `#mainmatter[...]` | 1 / 1.1 | 正文 |
| `#appendix[...]` | A / A.1 | 附录 |

```typst
#frontmatter[
  = 题记                        // 不编号章节
  正文内容……
]

#mainmatter[
  = 第一章 绪论 <sec-intro>      // 章节：1 第一章 绪论
  == 背景与动机 <sec-bg>         // 小节：1.1 背景与动机
  正文内容……
]

#appendix[
  = 配置手册 <sec-appendix>      // 附录：A 配置手册
  附录内容……
]
```

规则：

- 每个一级标题（`=`）生成一个独立章节文件（`chap_XX_xxx.xhtml`），并注册到 `nav.xhtml` 目录。
- 二级及以下标题（`==`、`===`）进入章节内的层级目录。
- `Contents/` 下的文件名任意（`01_intro.typ`、`ch1.typ` 均可），切章只按编译后的标题顺序。

## 定理与公式

模板基于 `@preview/ctheorems` 封装了学术环境，EPUB 导出时映射为语义卡片：

| 环境 | 用途 |
|---|---|
| `#theorem("标题")[...]` / `#lemma` / `#proposition` / `#corollary` | 定理族 |
| `#definition("标题")[...]` | 定义 |
| `#problem("标题")[...]` | 问题 |
| `#remark[...]` | 备注 |
| `#proof[...]` | 证明 |

数学公式：

- 行内：`$E = m c^2$`
- 块级：

```typst
$
  integral_(-infinity)^(+infinity) e^(-x^2) dif x = sqrt(pi)
$
```

包含公式的章节会在 `content.opf` 中自动声明 `properties="mathml"`。

## 语义容器

```typst
#overview(title: "核心速览")[ 本章要点…… ]
#poem[ 诗词内容 ]
#blockquote[ 引用内容 ]
#quote-block[ 引言 ]
```

## 特殊效果

- **涂黑文本**：`#box(fill: black)[机密内容]`（亮色与深色模式下均为实心黑块）。
- **注音**：`#rt("zhù yīn", "注音")`（EPUB 输出 `<ruby>`，PDF 输出 furiruby）。
- **图片**：

```typst
#figure(
  image("../assets/sample-diagram.svg", width: 85%),
  caption: [架构图],
)
```

以本地路径引用的图片会被按出现顺序编号为 `img_0001.png`…，复制进 EPUB 的 `OEBPS/images/` 并登记 Manifest，`src` 自动重写；新版 Typst 会把本地图片内联为 data URI，随章节直接输出，无需处理。

## 引用与脚注

```typst
= 核心理论 <sec-core>            // 定义锚点

正如 @sec-core 中所述……         // 跨章节引用
#link(<sec-core>)[查看]         // 或显式链接

正文#footnote[说明文字]          // 脚注自动归集到本章末尾
```

跨章节引用会被重写为 `chap_XX_xxx.xhtml#锚点` 形式；脚注在章尾生成 `epub:type="footnotes"` 区块，支持双向跳转。

## 表格

```typst
#table(
  columns: (1.5fr, 1fr, 2fr),
  align: (left, center, left),
  table.header([*标准*], [*版本*], [*说明*]),
  [EPUB], [3.3], [现代电子书标准],
  [MathML], [3.0], [数学公式渲染],
)
```

## 样式覆盖

视觉变量集中在 `scripts/epub-style.css` 的 `:root`，覆盖变量即可，无需改选择器：

```css
:root {
  --primary-color: #3c71b7;   /* 主题蓝 */
  --link-color:    #1f5f99;   /* 链接颜色 */
  --thm-fill:      #eef6ff;   /* 定理卡片底色 */
  --thm-stroke:    #7aa7d9;   /* 定理卡片边框色 */
}
```

深色模式在 `@media (prefers-color-scheme: dark)` 中重写同一组变量。
