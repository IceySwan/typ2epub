// ============================================================
// Fonts
// ============================================================
#let cjk-font-stack(stack) = (
  (name: stack.at(0), covers: "latin-in-cjk"),
  stack.at(1),
)

#let heading-font = cjk-font-stack(("Source Sans 3", "Noto Sans CJK SC"))
#let body-font = cjk-font-stack(("Source Serif 4", "Noto Serif CJK SC"))
#let environment-font = cjk-font-stack(("Source Serif 4", "LXGW Wenkai"))
#let math-font = "New Computer Modern Math"
#let code-font = "Maple Mono NF"

// ============================================================
// Brand & Base Colors
// ============================================================
#let structure-color = rgb("3c71b7")
#let link-color = rgb("1f5f99")
#let stroke-color = luma(200)
#let fill-color = luma(250)

// ============================================================
// Environment Color Schemes
// ============================================================
#let theorem-scheme = (fill: rgb("eef6ff"), stroke: rgb("7aa7d9"))
#let definition-scheme = (fill: rgb("f2fbf4"), stroke: rgb("73ad7b"))
#let problem-scheme = (fill: rgb("fff8e8"), stroke: rgb("c89b3c"))
#let overview-scheme = (fill: rgb("f7f7fb"), stroke: rgb("a8a8b8"))
#let area-scheme = (fill: rgb("fbfcef"), stroke: rgb("bfa004"))
