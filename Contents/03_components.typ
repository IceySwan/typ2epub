#import "../template.typ": *

= 定理环境与语义组件 <sec-components>

本模版内置了丰富的语义化定理环境与块级组件，在 EPUB 与 PDF 双端均拥有统一的美观视觉风格。

== 经典定理环境

#definition("半群 (Semigroup)")[
  若代数系统 $(S, *)$ 满足二元运算 $*$ 在集合 $S$ 上是封闭的，且对任意 $a, b, c in S$ 均满足结合律：
  $
    (a * b) * c = a * (b * c)
  $
  则称 $(S, *)$ 为一个半群。
]

#theorem("拉格朗日中值定理")[
  若函数 $f(x)$ 满足在闭区间 $[a, b]$ 上连续，在开区间 $(a, b)$ 内可导，则在 $(a, b)$ 内至少存在一点 $xi$，使得：
  $
    f'(xi) = (f(b) - f(a)) / (b - a)
  $
]

#proof[
  构造辅助函数 $g(x) = f(x) - f(a) - frac(f(b) - f(a), b - a)(x - a)$，易证 $g(a) = g(b) = 0$。根据罗尔中值定理，必存在 $xi in (a, b)$ 使得 $g'(xi) = 0$，代入即得结论。
]

== 块级语义容器

#overview(title: "设计概览")[
  通过在 `template/components.typ` 中使用 `html.elem` 针对 EPUB 进行定制化映射，所有块级组件在导出 HTML 时均保留了精准的语义 Class，从而可以通过 CSS 完成细腻的圆角边框、柔和背景与深色模式自适应。
]

#blockquote[
  “书籍是人类进步的阶梯。” —— 这是引用块（Blockquote）排版效果。
]

#poem[
  海上生明月，天涯共此时。\
  情人怨遥夜，竟夕起相思。\
  灭烛怜光满，披衣觉露滋。\
  不堪盈手赠，还寝梦佳期。
]

== 特殊排版特性

1. *黑条机密文本（Redacted Text）*：通过 `#box(fill: black)[机密内容]` 实现涂黑隐藏，如：绝密计划代号为 #box(fill: black)[PROJECT-NEXUS-2026]，在暗黑模式与亮色模式下均保持完全遮盖，读者划词选中或点击方可阅读。
2. *Ruby 注音*：支持中日文注音标注，例如：#rt("ふりがな", "振仮名") 与 #rt("zhu yin", "注音")。
