#import "template/components.typ": *

// ============================================================
// Helpers
// ============================================================
#let to-plain-text(it) = {
  if it == none { "" }
  else if type(it) == str { it }
  else if type(it) == datetime { it.display("[year]-[month]-[day]") }
  else if type(it) == content {
    if it.has("text") { it.text }
    else if it.has("children") { it.children.map(to-plain-text).join("") }
    else if it.has("body") { to-plain-text(it.body) }
    else if it.has("child") { to-plain-text(it.child) }
    else { "" }
  } else { str(it) }
}

// ============================================================
// Section Environments (frontmatter / mainmatter / appendix)
// ============================================================
#let frontmatter(body) = {
  set page(numbering: "i")
  set heading(numbering: none)
  counter(page).update(1)
  if is-epub {
    html.elem("section", attrs: ("data-t2e-matter": "front"), body)
  } else {
    body
  }
}

#let mainmatter(body) = {
  set page(numbering: "1")
  set heading(numbering: "1.1")
  counter(page).update(1)
  counter(heading).update(0)
  if is-epub {
    html.elem("section", attrs: ("data-t2e-matter": "main"), body)
  } else {
    body
  }
}

#let appendix(body) = {
  set page(numbering: "1")
  set heading(numbering: "A.1")
  counter(heading).update(0)
  if is-epub {
    html.elem("section", attrs: ("data-t2e-matter": "appendix"), body)
  } else {
    body
  }
}

// ============================================================
// Main Template: elegantbook
// ============================================================
#let elegantbook(
  title: none,
  subtitle: none,
  author: none,
  date: none,
  version: none,
  lang: "zh-CN",
  body,
) = {
  show: thmrules.with(qed-symbol: $square$)
  set document(
    title: title,
    author: author,
    date: date,
  )
  set page(
    paper: "a4",
    margin: (x: 2.54cm, y: 2.0cm),
    numbering: none,
  )
  set text(
    lang: if lang.starts-with("zh") { "zh" } else { lang },
    region: if lang == "zh-CN" { "CN" } else { none },
    font: body-font,
    fallback: true,
    cjk-latin-spacing: auto,
    size: 10.5pt,
  )
  set par(
    justify: true,
    leading: 0.75em,
    first-line-indent: 2em,
    linebreaks: "optimized",
  )
  set heading(numbering: "1.1")
  set math.equation(numbering: equation-numbering, supplement: none)

  // Global show rules
  show heading: set text(font: heading-font, fill: structure-color)
  show math.equation: set text(font: math-font)
  show raw: set text(font: code-font)
  show link: set text(fill: link-color)

  // Outline handling for EPUB
  show outline: it => if is-epub {
    html.elem("nav", attrs: ("data-t2e-role": "toc"), it)
  } else {
    it
  }

  // Align & Rect mapping for EPUB
  show align: it => if is-epub {
    let pos = if it.alignment == right { "right" } else if it.alignment == center { "center" } else { "left" }
    html.elem("div", attrs: (class: "text-align-" + pos), it.body)
  } else {
    it
  }
  show rect: it => if is-epub {
    html.elem("div", attrs: (class: "custom-rect"), it.body)
  } else {
    it
  }
  show box: it => if is-epub {
    if it.fill == black {
      html.elem("span", attrs: (class: "redacted"), it.body)
    } else {
      it
    }
  } else {
    it
  }

  // Heading spacing & equation counter resets
  show heading.where(level: 1): it => {
    pagebreak()
    counter(math.equation).update(0)
    set block(above: 1em, below: 1em)
    it
  }
  show heading.where(level: 2): it => {
    counter(math.equation).update(0)
    set block(above: 0.8em, below: 0.8em)
    it
  }

  show math.equation.where(block: true): it => layout(size => {
    let natural = measure(it)
    if natural.width > size.width {
      let factor = (size.width / natural.width) * 100%
      scale(x: factor, y: factor, origin: left, reflow: true, it)
    } else {
      it
    }
  })

  // Export structured metadata for Go / toolchain query
  [#metadata((
    title: to-plain-text(title),
    subtitle: to-plain-text(subtitle),
    author: to-plain-text(author),
    date: to-plain-text(date),
    version: to-plain-text(version),
    lang: to-plain-text(lang),
  )) <t2e-metadata>]

  // Title page (PDF only or tagged for EPUB)
  if not is-epub {
    align(center)[
      #if title != none {
        text(font: heading-font, size: 22pt, weight: "bold", fill: structure-color, title)
      }
      #if subtitle != none {
        v(0.8em)
        text(font: heading-font, size: 12pt, subtitle)
      }
      #if author != none {
        v(1.2em)
        author
      }
      #if date != none {
        v(0.4em)
        date.display("[year]年[month]月[day]日")
      }
      #if version != none {
        v(0.4em)
        version
      }
    ]
  }

  body
}
