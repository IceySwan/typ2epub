#import "@preview/ctheorems:1.1.3": *
#import "@preview/furiruby:0.1.0" as fr
#import "theme.typ": *

#let is-epub = sys.inputs.at("target", default: "pdf") == "epub"

// ============================================================
// Ruby & Math Operators & Numbering
// ============================================================
#let rt(top, bottom) = if is-epub {
  html.elem("ruby", [
    #bottom
    #html.elem("rp", [(])
    #html.elem("rt", top)
    #html.elem("rp", [)])
  ])
} else {
  fr.rt(top, bottom)
}

#let rm(body) = math.upright(body)
#let pd = math.partial
#let pm = math.plus.minus
#let mp = math.minus.plus

#let equation-numbering(n) = context {
  let levels = counter(heading).get()
  let chapter = levels.at(0, default: 0)
  let section = levels.at(1, default: 0)
  numbering("(1.1.1)", chapter, section, n)
}

// ============================================================
// Environment Formatting
// ============================================================
#let environment-text(body) = text(
  font: environment-font,
  style: "italic",
  body,
)
#let environment-title(body) = strong(environment-text(body))
#let environment-body(body) = {
  set par(first-line-indent: 0em)
  environment-text(body)
}

#let epub-thm-env(kind, label) = (..args, body) => {
  let title = if args.pos().len() > 0 { args.pos().at(0) } else { none }
  html.elem("div", attrs: (
    "data-t2e-kind": kind,
    class: "thm-box thm-" + kind
  ), [
    #html.elem("div", attrs: (class: "thm-header"), strong[
      #label #if title != none [ (#title)]
    ])
    #html.elem("div", attrs: (class: "thm-content"), body)
  ])
}

#let make-thm-env(kind, label, scheme, base-level: 1) = {
  if is-epub {
    epub-thm-env(kind, label)
  } else {
    thmbox(
      kind,
      label,
      base: "heading",
      base_level: base-level,
      titlefmt: environment-title,
      bodyfmt: environment-body,
      fill: scheme.fill,
      stroke: scheme.stroke,
    ).with(
      supplement: [#label],
      refnumbering: if base-level == 0 { "1" } else { "1.1" },
    )
  }
}

// ============================================================
// Theorem Family
// ============================================================
#let theorem = make-thm-env("theorem", "定理", theorem-scheme)
#let lemma = make-thm-env("theorem", "引理", theorem-scheme)
#let proposition = make-thm-env("theorem", "命题", theorem-scheme)
#let corollary = make-thm-env("theorem", "推论", theorem-scheme)

#let definition = make-thm-env("definition", "定义", definition-scheme)
#let problem = make-thm-env("problem", "问题", problem-scheme)

#let remark = if is-epub {
  epub-thm-env("remark", "注")
} else {
  thmplain(
    "remark",
    "注",
    titlefmt: environment-title,
    bodyfmt: environment-body,
  ).with(numbering: none)
}

#let base-proof = thmproof("proof", "证明")
#let proof = if is-epub {
  epub-thm-env("proof", "证明")
} else {
  (..args, body) => {
    set par(first-line-indent: 0em)
    base-proof(..args)[#environment-text(body)]
  }
}

// ============================================================
// Novel Environments
// ============================================================
#let character = make-thm-env("character", "角色", theorem-scheme, base-level: 0)
#let lore = make-thm-env("lore", "设定", definition-scheme, base-level: 0)
#let area = make-thm-env("area", "势力", area-scheme, base-level: 0)

// ============================================================
// Block Environments
// ============================================================
#let quote-block(body) = if is-epub {
  html.elem("div", attrs: ("data-t2e-kind": "quote", class: "quote-block"), body)
} else {
  set text(font: environment-font)
  body
}

#let poem(body) = if is-epub {
  html.elem("div", attrs: ("data-t2e-kind": "poem", class: "poem-box"), body)
} else {
  set par(first-line-indent: 0em)
  block(
    width: 100%,
    inset: 2em,
    stroke: (y: 0.5pt),
    body,
  )
}

#let blockquote(body) = if is-epub {
  html.elem("blockquote", attrs: ("data-t2e-kind": "blockquote", class: "custom-blockquote"), body)
} else {
  block(
    width: 100%,
    fill: fill-color,
    inset: 2em,
    stroke: (y: 0.5pt + stroke-color),
    {
      set text(font: environment-font)
      body
    },
  )
}

#let overview(title: none, body) = if is-epub {
  html.elem("div", attrs: ("data-t2e-kind": "overview", class: "thm-box thm-overview"), [
    #if title != none { html.elem("div", attrs: (class: "thm-header"), strong(title)) }
    #html.elem("div", attrs: (class: "thm-content"), body)
  ])
} else {
  set par(first-line-indent: 0em)
  environment-text(block(
    width: 100%,
    inset: 1em,
    radius: 4pt,
    fill: overview-scheme.fill,
    stroke: overview-scheme.stroke,
  )[
    #if title != none { strong(title) }
    #if title != none { linebreak() }
    #body
  ])
}
