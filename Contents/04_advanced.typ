#import "../template.typ": *

= 媒体资源与交叉引用 <sec-advanced>

本章演示转换器对多媒体资源（如 SVG / PNG 图片）的自动探测收集，以及跨章节锚点链接重写功能。

== 图像资源自动打包

构建脚本会自动扫描正文中的 `<img>` 标签，解析本地相对路径并将其自动复制到 EPUB 的 `OEBPS/images/` 目录，同时在 `content.opf` 的 `<manifest>` 中登记合规的 MIME 类型。

#figure(
  image("../assets/sample-diagram.svg", width: 85%),
  caption: [转换流水线架构示意图],
)

== 跨章节双向交叉引用

转换脚本在处理 HTML AST 时，会收集所有章节中出现的元素 `id` 属性，并自动重写 `#id` 相对链接为跨文件的 `chap_XX_xxx.xhtml#id` 形式：

- 点击返回 @sec-intro 阅读项目简介与设计目标；
- 点击查看 @sec-math 回顾数学公式与 MathML 支持；
- 点击跳转至 @sec-components 浏览各类定理环境展示。

== 章节结语

至此，您已完整浏览了 Typst 到 EPUB 转换工具链的主要特性。本工程已完全模块化，欢迎根据需要替换 `Contents/` 下的内容构建属于您自己的电子书！
