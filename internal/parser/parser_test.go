package parser_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/matthewmueller/diff"
	"github.com/matthewmueller/jsx/ast"
	"github.com/matthewmueller/jsx/internal/parser"
)

func equal(t *testing.T, input, expected string) {
	t.Helper()
	t.Run(input, func(t *testing.T) {
		t.Helper()
		actual := parser.Print(input, input)
		diff.TestString(t, actual, expected)
	})
}

func equalAST(t *testing.T, input string, expected ast.Node) {
	t.Helper()
	t.Run(input, func(t *testing.T) {
		t.Helper()
		actual, err := parser.Parse(input, input)
		if err != nil {
			t.Fatal(err)
		}
		diff.Test(t, actual, expected)
	})
}

var update = flag.Bool("update", false, "update golden files")

func equalFile(t *testing.T, path string) {
	t.Helper()
	t.Run(path, func(t *testing.T) {
		t.Helper()
		testdataPath := filepath.Join("..", "testdata")
		input, err := os.ReadFile(filepath.Join(testdataPath, path+".txt"))
		if err != nil {
			t.Fatal(err)
		}
		actual := parser.Print(path, string(input))
		expect, err := os.ReadFile(filepath.Join(testdataPath, path+".parse.txt"))
		if err != nil {
			if os.IsNotExist(err) {
				if err := os.WriteFile(filepath.Join(testdataPath, path+".parse.txt"), []byte(``), 0644); err != nil {
					t.Fatal(err)
				}
				return
			}
			t.Fatal(err)
		}
		if *update {
			if err := os.WriteFile(filepath.Join(testdataPath, path+".parse.txt"), []byte(actual), 0644); err != nil {
				t.Fatal(err)
			}
			return
		}
		diff.TestString(t, actual, string(expect))
	})
}

func equalJSXFile(t *testing.T, path string) {
	t.Helper()
	t.Run(path, func(t *testing.T) {
		t.Helper()
		input, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		actual := parser.Print(path, string(input))
		diff.TestString(t, actual, string(input))
	})
}

const children = `<body>
  <Page />
  <Scripts />
</body>
`

func TestSample(t *testing.T) {
	equal(t, `hello <span>world</span>`, `hello <span>world</span>`)
	equal(t, `hello <span class="hello-world">world</span>`, `hello <span class="hello-world">world</span>`)
	equal(t, `hello <span data-class="hello-world" id = "wonderful">world</span>`, `hello <span data-class="hello-world" id="wonderful">world</span>`)
	equal(t, `hello <button onClick={() => setCount(count + 1)}>Click me</button>`, `hello <button onClick={() => setCount(count + 1)}>Click me</button>`)
	equal(t, `hello <h2>world</h2>`, `hello <h2>world</h2>`)
	equal(t, `hello <input type="text" /> world`, `hello <input type="text" /> world`)
	equal(t, `hello <Planet>mars</Planet>`, `hello <Planet>mars</Planet>`)
	equal(t, children, children)
	equal(t, `hello <>fragment</>`, `hello <>fragment</>`)
	equal(t, `hello <h2>Record<string,string></h2>`, `hello <h2>Record<string,string></h2>`)
	equal(t, `type Record<string> = {}; function() { return <h2>hello world</h2> }`, `type Record<string> = {}; function() { return <h2>hello world</h2> }`)
	equal(t, `type Record<string> = {}; function() { return (<h2>hello world</h2>) }`, `type Record<string> = {}; function() { return (<h2>hello world</h2>) }`)
	equal(t, `function() { return (<h2 {...props}>{message}</h2>) }`, `function() { return (<h2 {...props}>{message}</h2>) }`)
	equal(t, `<><React.Fragment><>hello world</></React.Fragment></>`, `<><React.Fragment><>hello world</></React.Fragment></>`)

	equal(t, `hello <span data-x={<span>tag in attribute</span>}>world</span>`, `hello <span data-x={<span>tag in attribute</span>}>world</span>`)
	equal(t, `hello <span data-x={window ? <span>content1</span> : <span>content2</span>}>world</span>`, `hello <span data-x={window ? <span>content1</span> : <span>content2</span>}>world</span>`)
	equal(t, `hello <span data-x={<>text</>}></span>`, `hello <span data-x={<>text</>}></span>`)

	equal(t, `hello <span data-x={true}>world</span>`, `hello <span data-x={true}>world</span>`)
	equal(t, `hello <span data-x={i > 0}>world</span>`, `hello <span data-x={i > 0}>world</span>`)
	equal(t, `hello <span data-x={i >= 0}>world</span>`, `hello <span data-x={i >= 0}>world</span>`)
	equal(t, `hello <span data-x={i < 0}>world</span>`, `hello <span data-x={i < 0}>world</span>`)
	equal(t, `hello <span data-x={i <= 0}>world</span>`, `hello <span data-x={i <= 0}>world</span>`)
	equal(t, `hello <span data-x={i != 0}>world</span>`, `hello <span data-x={i != 0}>world</span>`)
	equal(t, `hello <span data-x={i == 0}>world</span>`, `hello <span data-x={i == 0}>world</span>`)
	equal(t, `hello <span data-x={i == 0}>world</span>`, `hello <span data-x={i == 0}>world</span>`)
}

func TestStyle(t *testing.T) {
	equalAST(t, `export default () => <style scoped>{`+"`"+`h1 { background-color: lightblue; }`+"`"+`}</style>`,
		&ast.Script{Body: []ast.Fragment{
			&ast.Text{Value: "export default () => "},
			&ast.Element{
				Name: "style",
				Attrs: []ast.Attr{
					&ast.Field{
						Name: "scoped",
						Value: &ast.BoolValue{
							Value: true,
						},
					},
				},
				Children: []ast.Fragment{&ast.Expr{
					Fragments: []ast.Fragment{
						&ast.Text{Value: "`h1 { background-color: lightblue; }`"},
					}},
				},
			},
		}},
	)
}

func TestMultiLineExpr(t *testing.T) {
	equalAST(t, `<h1 class={
		"hello"
	}>hi</h1>`, &ast.Script{Body: []ast.Fragment{
		&ast.Element{
			Name: "h1",
			Attrs: []ast.Attr{
				&ast.Field{
					Name: "class",
					Value: &ast.Expr{
						Fragments: []ast.Fragment{
							&ast.Text{Value: "\n\t\t\"hello\"\n"},
						},
					},
				},
			},
			Children: []ast.Fragment{&ast.Text{Value: "hi"}},
		},
	}})
}

func TestInExpr(t *testing.T) {
	equal(t, `export default function { return (<H1 func={() => <h1>hello world</h1>} />) }`, `export default function { return (<H1 func={() => <h1>hello world</h1>} />) }`)
	equal(t, `export default function { return (<H2 func={() => <Header>hello world</Header>} />) }`, `export default function { return (<H2 func={() => <Header>hello world</Header>} />) }`)
	equal(t, `export default function () { return (<H2 func={<Header {...props}>hello <span>world</span></Header>}/>)}`, `export default function () { return (<H2 func={<Header {...props}>hello <span>world</span></Header>} />)}`)
	equalAST(t, `export default function { return (<H1 func={() => <h1>hello world</h1>} />) }`, &ast.Script{Body: []ast.Fragment{
		&ast.Text{Value: "export default function { return ("},
		&ast.Element{
			Name: "H1",
			Attrs: []ast.Attr{
				&ast.Field{
					Name: "func",
					Value: &ast.Expr{
						Fragments: []ast.Fragment{
							&ast.Text{
								Value: "() => ",
							},
							&ast.Element{
								Name: "h1",
								Children: []ast.Fragment{
									&ast.Text{
										Value: "hello world",
									},
								},
							},
						},
					},
				},
			},
			SelfClosing: true,
		},
		&ast.Text{
			Value: ") }",
		},
	}})
	equalAST(t, `export default function { return (<H2 func={() => <Header>hello world</Header>} />) }`, &ast.Script{Body: []ast.Fragment{
		&ast.Text{Value: "export default function { return ("},
		&ast.Element{
			Name: "H2",
			Attrs: []ast.Attr{
				&ast.Field{
					Name: "func",
					Value: &ast.Expr{
						Fragments: []ast.Fragment{
							&ast.Text{
								Value: "() => ",
							},
							&ast.Element{
								Name: "Header",
								Children: []ast.Fragment{
									&ast.Text{
										Value: "hello world",
									},
								},
							},
						},
					},
				},
			},
			SelfClosing: true,
		},
		&ast.Text{
			Value: ") }",
		},
	}})
}

func TestJSXComment(t *testing.T) {
	equal(t, `export default () => (<h2>{/* hello world */}</h2>)`, `export default () => (<h2>{/* hello world */}</h2>)`)
	equal(t, `export default () => (<h2>hello {/* hello world */} world</h2>)`, `export default () => (<h2>hello {/* hello world */} world</h2>)`)
	equal(t, `export default () => (<h2>hello {hello /* hello world */} world</h2>)`, `export default () => (<h2>hello {hello /* hello world */} world</h2>)`)
}

func TestTSXFile(t *testing.T) {
	equalFile(t, "01-hello.tsx")
	equalFile(t, "02-document.jsx")
	equalFile(t, "03-button.jsx")
	equalFile(t, "04-faq.jsx")
	equalFile(t, "05-footer.jsx")
	equalFile(t, "06-header.jsx")
	equalFile(t, "07-index.jsx")
	equalFile(t, "08-pay-edit.jsx")
	equalFile(t, "09-pay.jsx")
	equalFile(t, "10-privacy.jsx")
	equalFile(t, "11-slack-button.jsx")
	equalFile(t, "12-document.tsx")
	equalFile(t, "13-inner-fragment.tsx")
}

// These tests come from styled-jsx
func TestStyledJSXFile(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "testdata", "styled-jsx", "*.js"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		equalJSXFile(t, file)
	}
}

func TestSyntaxError(t *testing.T) {
	ast, err := parser.Parse("input.jsx", `export default () => <h1>hello world</h1`)
	if err == nil {
		t.Fatal("expected an error")
	}
	if ast != nil {
		t.Fatal("expected nil ast")
	}
	diff.TestString(t, err.Error(), "expected >, got unexpected end of input while lexing (1:40)")
}

func TestIssue1(t *testing.T) {
	equalAST(t, `<div>{h.components( [ { field: x => "(<><button>PUSH_ME</button></>)", label: "Actions"} ])}</div>`, &ast.Script{Body: []ast.Fragment{
		&ast.Element{
			Name: "div",
			Children: []ast.Fragment{
				&ast.Expr{Fragments: []ast.Fragment{
					&ast.Text{Value: `h.components( [ { field: x => "(<><button>PUSH_ME</button></>)", label: "Actions"} ])`},
				}},
			},
		},
	}})
}

func TestIssue2(t *testing.T) {
	equal(t, `<div class='test'> </div>`, `<div class='test'> </div>`)
	equalAST(t, `<div class='test'> </div>`, &ast.Script{Body: []ast.Fragment{
		&ast.Element{
			Name: "div",
			Attrs: []ast.Attr{
				&ast.Field{
					Name: "class",
					Value: &ast.StringValue{
						Value: "test",
						Raw:   "'test'",
					},
				},
			},
			Children: []ast.Fragment{&ast.Text{Value: " "}},
		},
	}})
	equal(t, `<div class="child-width-1-2\@_m"> </div>`, `<div class="child-width-1-2\@_m"> </div>`)
	equalAST(t, `<div class="child-width-1-2\@_m"> </div>`, &ast.Script{Body: []ast.Fragment{
		&ast.Element{
			Name: "div",
			Attrs: []ast.Attr{
				&ast.Field{
					Name: "class",
					Value: &ast.StringValue{
						Value: "child-width-1-2\\@_m",
						Raw:   `"child-width-1-2\@_m"`,
					},
				},
			},
			Children: []ast.Fragment{&ast.Text{Value: " "}},
		},
	}})
}

func TestTemplFile(t *testing.T) {
	equalFile(t, "14-time.templ")
}
