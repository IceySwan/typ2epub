#import "template.typ": *

#show: elegantbook.with(
  title: [Typst to EPUB 电子书示例],
  subtitle: [高保真 EPUB 3 / 2 导出工具链与排版模版],
  author: "Icey Swan",
  date: datetime.today(),
  version: [0.1.0],
)

#frontmatter[
  = 题记

  #quote-block[
    排版不仅是文字的陈列，更是思想在介质间的无损传递。
  ]

  本文档用于演示 *Typst to EPUB* 转换流水线的完整排版能力，涵盖分章与多级目录、MathML 数学公式、分章独立脚注、语义定理环境、图像自动收集及深色模式适配。

  #outline(title: [目录], depth: 3)
]

#mainmatter[
  #include "Contents/01_introduction.typ"
  #include "Contents/02_math_and_code.typ"
  #include "Contents/03_components.typ"
  #include "Contents/04_advanced.typ"
]
#appendix[
  #include "Contents/09_appendix.typ"
]
