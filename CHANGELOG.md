# Changelog

## 2026.08.20 - v0.1.0 首版（个人自用精简版）

个人自用定位的首个正式版本。工具链完成一轮"去过度设计"重构：保留核心转换能力，删除多用户、可复现构建等非必要设施。

### 核心能力

- Typst 模板输出语义 HTML（`data-t2e-matter` / `data-t2e-kind` / `data-t2e-role` / `data-t2e-split`），Go 转换器经 `typst query` 读取元数据、`typst compile` 导出 HTML 后处理
- DOM 流水线：全局脚注提取并按章节回填、静态目录移除、按一级标题切章、图片收集、跨章节引用重写
- EPUB 3 输出：`nav.xhtml` 目录（含 Landmarks 地标）、`content.opf`（MathML/SVG 属性声明）、`mimetype` 首位 STORED 零额外字段合规打包、ZIP 原子替换

### 本版取舍（个人自用）

- 移除 EPUB 2 NCX 双导航、可复现构建（`SOURCE_DATE_EPOCH` / SHA-256 校验）、装饰性 SVG 封面、Manifest / AST 预校验等防御设施，代码精简至约 1300 行
- 封面三级回退：`-c` 参数 → `assets/cover.png` / `cover.jpg` / `cover.pdf` → 自动生成白底黑字简洁封面
- 图片按出现顺序编号（`img_0001.png`）收集；按本地资源假设处理，不做远程 / 路径沙箱校验
- Manifest / Spine 强类型化，XML 输出统一转义

### 工程

- 单测覆盖 DOM 变换核心逻辑（分章、脚注、引用、TOC）；端到端测试真实调用 typst 构建并校验 EPUB 结构
- 文档：`README.md`、`docs/architecture.md`、`docs/guide.md`、`docs/faq.md`
