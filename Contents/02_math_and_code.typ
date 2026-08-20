#import "../template.typ": *

= 数学公式与代码高亮 <sec-math>

EPUB 3 规范推荐使用 MathML 呈现数学公式。构建脚本会自动检测当前章节中的 `<math>` 标签并在 `content.opf` 的 `<item>` 元数据中自动附加 `properties="mathml"` 属性。

== 数学公式示例

行内公式示例：著名的质能方程 $E = m c^2$，以及欧拉恒等式 $e^(i pi) + 1 = 0$。

块级多行公式与积分运算：

$
  cal(F)(omega) = integral_(-infinity)^(+infinity) f(t) e^(-i omega t) dif t
$

高斯正态分布概率密度函数：

$
  f(x) = 1 / (sigma sqrt(2 pi)) e^(- 1/2 ((x - mu) / sigma)^2)
$

在移动端阅读器中，数学公式自带防溢出水平滚动与字体平滑缩放保护。

== 代码块展示

转换器完整保留了 Typst 代码块的高亮与等宽字体排版：

```typst
#let greeting(name) = [
  Hello, #name! Welcome to Typst EPUB exporter.
]

#greeting("Reader")
```

以及 Go 语言脚本示例：

```go
package main

import "fmt"

func main() {
    fmt.Println("Typst to EPUB pipeline in Go!")
}
```
