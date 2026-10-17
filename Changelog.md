# 0.0.13 / 2026-10-17

- fix: treat `a < b` (whitespace on both sides) as less than, not a tag
- fix: skip JS strings, template literals, comments and regex so JSX-like text inside them isn't parsed (e.g. `console.warn("use a <Provider>")`)
- fix: keep brace depth when an expression contains a nested element or comment (e.g. `f={() => { return <B /> }}`)
- fix: allow `_`, `$` and `:` in tag names (e.g. `<Icon_S />`, `<_3D />`, `<svg:rect />`)
- fix: allow backslashes and newlines in attribute strings
- fix: recognize JSX right after `[`, `,`, `=`, `=>`, `?`, `:`, `&&` and `||` without a space
- fix: support comments between attributes and whitespace in closing tags

# 0.0.12 / 2026-01-19

- support inline comments

# 0.0.11 / 2025-03-23

- handle < that aren't elements

# 0.0.10 / 2025-01-19

- Added support for ':', '.' and '\_' in attribute name

# 0.0.9 / 2025-01-16

- fix: less than (<) in expressions

# 0.0.8 / 2024-09-03

- fix unquoting string attributes. closes: #2
- handle jsx elements within strings. closes: #1

# 0.0.7 / 2024-09-02

- fix when jsx immediately after {

# 0.0.6 / 2024-04-26

- remove fmt

# 0.0.5 / 2024-04-26

- pass lexing errors through

# 0.0.4 / 2024-04-04

- fix lexing inner fragments
- fix parsing components with a dot.

# 0.0.3 / 2024-04-03

- add support for jsx comments
- add support for exprs inside exprs
- better line tracking and states always return back to initial state
- improve string values

# 0.0.2 / 2024-03-24

- add visitor support

# 0.0.1 / 2024-03-24

- initial commit
