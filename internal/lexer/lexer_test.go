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

const newlined = `export default () => (
	<div
		className="hello"
	>
		hello
		<span>world</span>
	</div>
)`

func TestSample(t *testing.T) {
	equal(t, `function() { return <h2>hello <span>world</span>.</h2> }`, `text:"function() { return " < identifier:"h2" > text:"hello " < identifier:"span" > text:"world" </ identifier:"span" > text:"." </ identifier:"h2" > text:" }"`)
	equal(t, `hello <span>world</span>`, `text:"hello " < identifier:"span" > text:"world" </ identifier:"span" >`)
	equal(t, `hello <span class="hello-world">world</span>`, `text:"hello " < identifier:"span" space:" " identifier:"class" = string:"\"hello-world\"" > text:"world" </ identifier:"span" >`)
	equal(t, `hello <span data-class="hello-world" id = "wonderful">world</span>`, `text:"hello " < identifier:"span" space:" " identifier:"data-class" = string:"\"hello-world\"" space:" " identifier:"id" space:" " = space:" " string:"\"wonderful\"" > text:"world" </ identifier:"span" >`)
	equal(t, `hello <button onClick={() => setCount(count + 1)}>Click me</button>`, `text:"hello " < identifier:"button" space:" " identifier:"onClick" = { expr:"() => setCount(count + 1)" } > text:"Click me" </ identifier:"button" >`)
	equal(t, `hello <h2>world</h2>`, `text:"hello " < identifier:"h2" > text:"world" </ identifier:"h2" >`)
	equal(t, `hello <input type="text" /> world`, `text:"hello " < identifier:"input" space:" " identifier:"type" = string:"\"text\"" space:" " /> text:" world"`)
	equal(t, `hello <h2>Record<string,string></h2>`, `text:"hello " < identifier:"h2" > text:"Record" < identifier:"string" text:",string>" </ identifier:"h2" >`)
	equal(t, `type Record<string> = {}; function() { return <h2>hello world</h2> }`, `text:"type Record<string> = {}; function() { return " < identifier:"h2" > text:"hello world" </ identifier:"h2" > text:" }"`)
	equal(t, `type Record<string> = {}; function() { return (<h2>hello world</h2>) }`, `text:"type Record<string> = {}; function() { return (" < identifier:"h2" > text:"hello world" </ identifier:"h2" > text:") }"`)
	equal(t, `hello <Planet>mars</Planet>`, `text:"hello " < identifier:"Planet" > text:"mars" </ identifier:"Planet" >`)
	equal(t, `<hr style={{ maxWidth: 400 }} />`, `< identifier:"hr" space:" " identifier:"style" = { expr:"{ maxWidth: 400 }" } space:" " />`)
	equal(t, children, `< identifier:"body" > text:"\n\t" < identifier:"Page" space:" " /> text:"\n\t" < identifier:"Scripts" space:" " /> text:"\n" </ identifier:"body" > text:"\n"`)
	equal(t, `hello <>fragment</>`, `text:"hello " < > text:"fragment" </ >`)
	equal(t, newlined, `text:"export default () => (\n\t" < identifier:"div" space:"\n\t\t" identifier:"className" = string:"\"hello\"" space:"\n\t" > text:"\n\t\thello\n\t\t" < identifier:"span" > text:"world" </ identifier:"span" > text:"\n\t" </ identifier:"div" > text:"\n)"`)
	equal(t, `function() { return (<h2 {...props}>{message}</h2>) }`, `text:"function() { return (" < identifier:"h2" space:" " { expr:"...props" } > { expr:"message" } </ identifier:"h2" > text:") }"`)
	equal(t, `<React.Fragment>hello</React.Fragment>`, `< identifier:"React.Fragment" > text:"hello" </ identifier:"React.Fragment" >`)
	equal(t, `<>hello</>`, `< > text:"hello" </ >`)
	equal(t, `<a><a></a></a>`, `< identifier:"a" > < identifier:"a" > </ identifier:"a" > </ identifier:"a" >`)
	equal(t, `<a><></></a>`, `< identifier:"a" > < > </ > </ identifier:"a" >`)
	equal(t, `<><></></>`, `< > < > </ > </ >`)
	equal(t, `<><React.Fragment><>hello world</></React.Fragment></>`, `< > < identifier:"React.Fragment" > < > text:"hello world" </ > </ identifier:"React.Fragment" > </ >`)
}

func TestInExpr(t *testing.T) {
	equal(t, `export default function { return (<H1 func={() => <h1>hello world</h1>} />) }`, `text:"export default function { return (" < identifier:"H1" space:" " identifier:"func" = { expr:"() => " < identifier:"h1" > text:"hello world" </ identifier:"h1" > } space:" " /> text:") }"`)
	equal(t, `export default function { return (<H2 func={() => <Header>hello <span>world</span></Header>} />) }`, `text:"export default function { return (" < identifier:"H2" space:" " identifier:"func" = { expr:"() => " < identifier:"Header" > text:"hello " < identifier:"span" > text:"world" </ identifier:"span" > </ identifier:"Header" > } space:" " /> text:") }"`)
	equal(t, `export default function (props) { return (<H2 func={() => <Header {...props}>hello <span>world</span></Header>} />) }`, `text:"export default function (props) { return (" < identifier:"H2" space:" " identifier:"func" = { expr:"() => " < identifier:"Header" space:" " { expr:"...props" } > text:"hello " < identifier:"span" > text:"world" </ identifier:"span" > </ identifier:"Header" > } space:" " /> text:") }"`)
	equal(t, `export default function (props: Record<string>) { return (<H2 func={() => <Header {...props}>hello <span>world</span></Header>} />) }`, `text:"export default function (props: Record<string>) { return (" < identifier:"H2" space:" " identifier:"func" = { expr:"() => " < identifier:"Header" space:" " { expr:"...props" } > text:"hello " < identifier:"span" > text:"world" </ identifier:"span" > </ identifier:"Header" > } space:" " /> text:") }"`)
	equal(t, `export default function () { return (<H2 func={<Header {...props}>hello <span>world</span></Header>}/>)}`, `text:"export default function () { return (" < identifier:"H2" space:" " identifier:"func" = { < identifier:"Header" space:" " { expr:"...props" } > text:"hello " < identifier:"span" > text:"world" </ identifier:"span" > </ identifier:"Header" > } /> text:")}"`)

	equal(t, `export default function () { return (<span data-x={<>text</>}></span>)}`, `text:"export default function () { return (" < identifier:"span" space:" " identifier:"data-x" = { < > text </ > } > </ identifier:"span" > text:")}"`)
	equal(t, `export default function () { return (<span data-x={<span data-x={<span>foo</span>}>text</span>}></span>)}`, `text:"export default function () { return (" < identifier:"span" space:" " identifier:"data-x" = { < identifier:"span" space:" " identifier:"data-x" = { < identifier:"span" > text:"foo" </ identifier:"span" > } > text </ identifier:"span" > } > </ identifier:"span" > text:")}"`)

}

func TestJSXComment(t *testing.T) {
	equal(t, `export default () => (<h2>{/* hello world */}</h2>)`, `text:"export default () => (" < identifier:"h2" > { comment:"/* hello world */" } </ identifier:"h2" > text:")"`)
	equal(t, `export default () => (<h2>hello {/* hello world */} world</h2>)`, `text:"export default () => (" < identifier:"h2" > text:"hello " { comment:"/* hello world */" } text:" world" </ identifier:"h2" > text:")"`)
	equal(t, `export default () => (<h2>hello {hello /* hello world */} world</h2>)`, `text:"export default () => (" < identifier:"h2" > text:"hello " { expr:"hello " comment:"/* hello world */" } text:" world" </ identifier:"h2" > text:")"`)
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
	equalFile(t, "13-inner-fragment.tsx")
}

func TestIssue1(t *testing.T) {
	equal(t, `<div>{h.components( [ { field: x => "(<><button>PUSH_ME</button></>)", label: "Actions"} ])}</div>`, `< identifier:"div" > { expr:"h.components( [ { field: x => \"(<><button>PUSH_ME</button></>)\", label: \"Actions\"} ])" } </ identifier:"div" >`)
	equal(t, `<div>{h.components( [ { field: x => '(<><button>PUSH_ME</button></>)', label: 'Actions'} ])}</div>`, `< identifier:"div" > { expr:"h.components( [ { field: x => '(<><button>PUSH_ME</button></>)', label: 'Actions'} ])" } </ identifier:"div" >`)
	equal(t, `<div>{h.components( [ { field: x => `+"`"+`(<><button>PUSH_ME</button></>)`+"`"+`, label: `+"`"+`Actions`+"`"+`} ])}</div>`, `< identifier:"div" > { expr:"h.components( [ { field: x => `+"`"+`(<><button>PUSH_ME</button></>)`+"`"+`, label: `+"`"+`Actions`+"`"+`} ])" } </ identifier:"div" >`)
}

func TestIssue2(t *testing.T) {
	equal(t, `<div class="child-width-1-2\@_m"> </div>`, `< identifier:"div" space:" " identifier:"class" = string:"\"child-width-1-2\\@_m\"" > text:" " </ identifier:"div" >`)
}

func TestAttributes(t *testing.T) {
	equal(t, `<div x:bind="model"></div>`, `< identifier:"div" space:" " identifier:"x:bind" = string:"\"model\"" > </ identifier:"div" >`)
	equal(t, `<div x_bind="model"></div>`, `< identifier:"div" space:" " identifier:"x_bind" = string:"\"model\"" > </ identifier:"div" >`)
	equal(t, `<div x.bind="model"></div>`, `< identifier:"div" space:" " identifier:"x.bind" = string:"\"model\"" > </ identifier:"div" >`)
}

func TestForLoop(t *testing.T) {
	equal(t, `<div>for i := 0; i < 25; i++ {  }</div>`, `< identifier:"div" > text:"for i := 0; i " < space:" " text:"2" text:"5; i++ " { expr:"  " } </ identifier:"div" >`)
	equal(t, `<div>for i := 0; i <25; i++ {  }</div>`, `< identifier:"div" > text:"for i := 0; i " text:"<25; i++ " { expr:"  " } </ identifier:"div" >`)
	equal(t, `<   div></div>`, `< space:"   " identifier:"div" > </ identifier:"div" >`)
	equal(t, `<div><   a></a></div>`, `< identifier:"div" > < space:"   " identifier:"a" > </ identifier:"a" > </ identifier:"div" >`)
}

func TestTempl(t *testing.T) {
	equalFile(t, "14-time.templ")
}
