# Typst to EPUB

> 因 Typst 尚未支持 EPUB 导出，该项目最初仅为满足 [个人小说](https://www.icey.one/notes/) 输出为 EPUB 的需求而编写的脚本。后来发现社区暂未有人实现类似功能，故将脚本抽象后使用 Go 重写开源。个人项目难免存在 Bug，如有问题请提 Issue。

一个 Typst → EPUB 的轻量级脚本，只需调用 Typst CLI，无其他依赖。

## 特性

- EPUB 3.3 规范输出（`mimetype` 首位 STORED、零额外字段）
- 前言 / 正文 / 附录分区，按一级标题自动拆分章节并生成目录
- 分章尾注（可双向跳转）、MathML 公式、SVG 图片
- 图片自动收集：按出现顺序编号复制进 EPUB，重写引用路径
- 跨章节引用自动重写
- 封面：`-c` 指定 → `assets/cover.png` / `cover.jpg` / `cover.pdf` → 都没有时自动生成白底黑字简洁封面

## 快速开始

> 环境要求 Go 1.25+ Typst CLI 0.15+

```bash
# 直接运行
go run ./scripts -i main.typ

# 或编译为独立二进制
go build -o build/typ2epub ./scripts
./build/typ2epub -i main.typ -o build/my_book.epub
```

### 命令行参数

| 参数 | 缩写 | 默认值 | 说明 |
|---|---|---|---|
| `--input` | `-i` | `main.typ` | Typst 入口文件 |
| `--output` | `-o` | `build/<标题>-v<版本>.epub` | 输出路径 |
| `--cover` | `-c` | 自动 | 自定义封面（PNG/JPG/SVG/PDF）；未指定时检测 `assets/cover.png`、`cover.jpg`、`cover.pdf`，都没有则生成白底黑字简洁封面 |
| `--keep-temp` | | `false` | 保留中间产物（`build/_epub_contents/`、`build/raw_export.html`） |

## 目录结构

```text
.
├── main.typ            # typ 主文件
├── template.typ        # Typst 模板
├── template/           # 模板组件（定理环境、主题变量）
├── Contents/           # typ 分章文件
├── assets/             # 静态资源
├── scripts/            # 转换脚本
│   ├── main.go         # CLI 入口
│   ├── builder.go      # 构建流程编排
│   ├── dom.go          # HTML DOM 变换（分章、脚注、引用）
│   ├── navigation.go   # nav.xhtml 生成
│   ├── media.go        # 封面与图片处理
│   ├── packager.go     # content.opf、容器文件与 ZIP 打包
│   ├── types.go        # 数据模型与文件名清理
│   ├── epub-style.css  # EPUB 样式（含深色模式）
│   └── *_test.go       # 单元测试与端到端构建测试
├── docs/               # 文档
└── build/              # 输出目录（git 忽略）
```

## 编写一本电子书

在 `main.typ` 中用三个分区环境组织内容：

```typst
#import "template.typ": *

#show: elegantbook.with(
  title: [Typst to EPUB 电子书示例],
  subtitle: [高保真 EPUB 3 / 2 导出工具链与排版模版],
  author: "Icey Swan",
  date: datetime.today(),
  version: [0.1.0],
)

#frontmatter[      // 前言：标题不带编号
  = 题记
  排版不仅是文字的陈列，更是思想在介质间的无损传递。
]

#mainmatter[       // 正文：标题自动编号 1, 1.1
  #include "Contents/01_introduction.typ"
  #include "Contents/02_math_and_code.typ"
]

#appendix[         // 附录：标题自动编号 A, A.1
  #include "Contents/09_appendix.typ"
]
```

要点：

- 每个一级标题（`=`）生成一个独立章节文件（如 `chap_01_xxx.xhtml`）。
- `Contents/` 下的文件名任意，切章只按编译后的标题顺序，与文件名无关。
- 子文件需自行导入模板：`#import "../template.typ": *`。

## 文档

- [架构说明](docs/architecture.md) — 构建流水线与 HTML IR 协议
- [排版与组件指南](docs/guide.md) — 分区、定理环境、公式、脚注等用法
- [FAQ](docs/faq.md) — 常见问题与排错

## License

[Apache-2.0](LICENSE)
