# 架构说明

Typst 负责求值与语义输出，Go 负责 DOM 变换与打包，两者通过语义 HTML 协议（`data-t2e-*` 属性）解耦：Typst 模板只负责打标记，转换器只按标记装配 EPUB。

## 构建流水线

```text
typst query <t2e-metadata>          ──► BookMeta JSON
typst compile --features html       ──► build/raw_export.html
        │
        ▼
net/html 解析为 DOM
  extractFootnotes       提取全局尾注，按引用顺序存入 FootnoteStore
  removeTypstTOC         删除静态目录（data-t2e-role="toc"）
  splitChapters          按分区与 <h2>（= 一级标题）切章
  assignFootnotes        尾注回填到各自章节末尾
  collectImages          图片按顺序编号复制到 OEBPS/images/ 并重写 src
  rewriteCrossReferences 跨章节 #id 链接补上目标文件名
        │
        ▼
生成 OEBPS：各章 XHTML、nav.xhtml、content.opf、container.xml
        │
        ▼
ZIP 打包：mimetype 首位 STORED（零额外字段）→ 其余 Deflate
```

## HTML IR 协议

模板层在导出 HTML 时注入以下语义属性，转换器据此处理：

| 属性 | 值 | 含义 |
|---|---|---|
| `data-t2e-matter` | `front` / `main` / `appendix` | 分部；第一个 `main` 章节作为 bodymatter 地标 |
| `data-t2e-role` | `toc` | 静态目录标记，会被移除，交由 EPUB 原生导航 |
| `data-t2e-kind` | `theorem` / `proof` / `poem` / `overview` / `quote` / `blockquote` | 语义容器，映射到 CSS 类 |
| `data-t2e-split` | `chapter` | 显式切章节点（优先级高于 `<h2>`） |

元数据通过 `typst query` 读取 `<t2e-metadata>` 标签的 JSON（title/author/version/date/lang），不做源码正则推断。

## 章节与导航

- **切章**：`<h2>`（Typst 一级标题 `=`）或 `data-t2e-split="chapter"`。
- **目录**：章节内 `<h3>`–`<h5>` 构建层级 TOC，输出 EPUB 3 `nav.xhtml`。
- **脚注**：正文中的 `sup[role="doc-noteref"]` 决定脚注归属章节，严格按引用出现顺序回填。
- **封面**：`-c` 参数 → `assets/cover.png` / `cover.jpg` / `cover.pdf` → 都没有时生成白底黑字简洁 SVG 封面。

## 关键设计

- **图片处理**：所有媒体均为本地资源，图片按出现顺序命名为 `img_0001.png`…，天然无重名冲突；`src` 统一重写为 `../images/` 相对路径。
- **结构化 Manifest**：Manifest/Spine 以结构体组装后渲染为 XML，避免字符串拼接出错。
- **EPUB 3.3 合规**：`mimetype` 归档首位、STORED、零额外字段；ZIP 先写临时文件，成功后原子 rename。
- **可测试性**：`dom.go` 全部为无状态纯函数，可输入 HTML 片段独立单测；构建流程有端到端测试。

## 源码结构

| 文件 | 职责 |
|---|---|
| `main.go` | CLI 参数解析与入口 |
| `builder.go` | 流水线编排、临时目录与产物生命周期 |
| `dom.go` | 纯函数 DOM 变换（分章、脚注、引用） |
| `navigation.go` | `nav.xhtml` 生成 |
| `media.go` | 封面、图片收集与媒体类型推断 |
| `packager.go` | `content.opf`、`container.xml`、ZIP 打包 |
| `types.go` | 数据模型与文件名清理 |
