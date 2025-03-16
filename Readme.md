# jsx

[![Go Reference](https://pkg.go.dev/badge/github.com/matthewmueller/jsx.svg)](https://pkg.go.dev/github.com/matthewmueller/jsx)

JSX parser for `.jsx` and `.tsx` files.

This package is primarily used to rewrite JSX in JS files (think [styled-jsx](https://github.com/matthewmueller/styledjsx)). This package does not parse JS, rather it finds JSX within JS.

## Install

```sh
go get github.com/matthewmueller/jsx
```

## Usage

```go
input := `export default () => <h1>hello world</h1>`
ast, _ := jsx.Parse("input.jsx", input)
fmt.Println(ast.String())
```

## Contributors

- Matt Mueller ([@mattmueller](https://twitter.com/mattmueller))

## License

MIT
