package jsx_test

import (
	"fmt"

	"github.com/matthewmueller/jsx"
)

func ExampleParse() {
	input := `export default () => <h1>hello world</h1>`
	ast, err := jsx.Parse("input.jsx", input)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ast.String())
	// Output:
	// export default () => <h1>hello world</h1>
}
