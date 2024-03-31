package lexer_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matthewmueller/diff"
	"github.com/matthewmueller/jsx/internal/lexer"
)

func equal(t *testing.T, input, expected string) {
	t.Helper()
	t.Run(input, func(t *testing.T) {
		t.Helper()
		actual := new(bytes.Buffer)
		tokens := lexer.Lex(input)
		actuals := make([]string, len(tokens))
		for i, tok := range tokens {
			actuals[i] = tok.String()
			actual.WriteString(tok.Text)
		}
		diff.TestString(t, strings.Join(actuals, " "), expected)
		// Verify we can recover the original input if there were no errors
		diff.TestString(t, actual.String(), input)
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
		expect, err := os.ReadFile(filepath.Join(testdataPath, path+".lex.txt"))
		if err != nil {
			t.Fatal(err)
		}
		tokens := lexer.Lex(string(input))
		actual := new(bytes.Buffer)
		for _, token := range tokens {
			actual.WriteString(token.String())
			actual.WriteString("\n")
		}
		if *update {
			if err := os.WriteFile(filepath.Join(testdataPath, path+".lex.txt"), actual.Bytes(), 0644); err != nil {
				t.Fatal(err)
			}
			return
		}
		diff.TestString(t, actual.String(), string(expect))
		// Verify we can recover the original input
		generated := new(bytes.Buffer)
		for _, token := range tokens {
			generated.WriteString(token.Text)
		}
		diff.TestString(t, generated.String(), string(input))
	})
}

const children = `<body>
	<Page />
	<Scripts />
</body>
`

func TestSample(t *testing.T) {
	equal(t, `hello <span>world</span>`, `text:"hello " < identifier:"span" > text:"world" </ identifier:"span" >`)
	equal(t, `hello <span class="hello-world">world</span>`, `text:"hello " < identifier:"span" space:" " identifier:"class" = string:"\"hello-world\"" > text:"world" </ identifier:"span" >`)
	equal(t, `hello <span data-class="hello-world" id = "wonderful">world</span>`, `text:"hello " < identifier:"span" space:" " identifier:"data-class" = string:"\"hello-world\"" space:" " identifier:"id" space:" " = space:" " string:"\"wonderful\"" > text:"world" </ identifier:"span" >`)
	equal(t, `hello <button onClick={() => setCount(count + 1)}>Click me</button>`, `text:"hello " < identifier:"button" space:" " identifier:"onClick" = expr:"{() => setCount(count + 1)}" > text:"Click me" </ identifier:"button" >`)
	equal(t, `hello <h2>world</h2>`, `text:"hello " < identifier:"h2" > text:"world" </ identifier:"h2" >`)
	equal(t, `hello <input type="text" /> world`, `text:"hello " < identifier:"input" space:" " identifier:"type" = string:"\"text\"" space:" " /> text:" world"`)
	equal(t, `hello <h2>Record<string,string></h2>`, `text:"hello " < identifier:"h2" > text:"Record" < identifier:"string" text:",string>" </ identifier:"h2" >`)
	equal(t, `type Record<string> = {}; function() { return <h2>hello world</h2> }`, `text:"type Record" text:"<string> = {}; function() { return " < identifier:"h2" > text:"hello world" </ identifier:"h2" > space:" " text:"}"`)
	equal(t, `type Record<string> = {}; function() { return (<h2>hello world</h2>) }`, `text:"type Record" text:"<string> = {}; function() { return (" < identifier:"h2" > text:"hello world" </ identifier:"h2" > text:") }"`)
	equal(t, `hello <Planet>mars</Planet>`, `text:"hello " < identifier:"Planet" > text:"mars" </ identifier:"Planet" >`)
	equal(t, children, `< identifier:"body" > space:"\n\t" < identifier:"Page" space:" " /> space:"\n\t" < identifier:"Scripts" space:" " /> space:"\n" </ identifier:"body" > space:"\n"`)
	equal(t, `hello <>fragment</>`, `text:"hello " < > text:"fragment" </ >`)
}

func TestTemplateLiteral(t *testing.T) {
	t.Skip("template literals are not supported yet")
	equal(t, `export default () => (<h2 class={`+"`"+`hello`+"`"+`}>`, ``)
	equal(t, `export default () => (<h2 class={`+"`"+`${hello}`+"`"+`}>`, ``)
	equal(t, `export default () => (<h2 class={`+"`"+`${hello}${world}`+"`"+`}>`, ``)
	equal(t, `export default () => (<h2 class={`+"`"+`${hello}${world}!`+"`"+`}>`, ``)
	equal(t, `export default () => (<h2 class={`+"`"+`hello ${hello}${world}!`+"`"+`}>`, ``)
	equal(t, `export default () => (<h2 class={`+"`"+`hello ${hello} ${world}!`+"`"+`}>`, ``)
}

func TestInExpr(t *testing.T) {
	t.Skip("elements in an expr is not supported yet")
	equal(t, `export default function { return (<H1 func={() => <h1>hello world</h1>} />) }`, ``)
	equal(t, `export default function { return (<H2 func={() => <Header>hello world</Header>} />) }`, ``)
}

func TestFile(t *testing.T) {
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
}
