package richtext

// DefaultStylesheet is the user-agent stylesheet: the look of plain HTML,
// in the spirit of a modern documentation page. Author styles (<style>
// elements, Stylesheet and style attributes) override it.
const DefaultStylesheet = `
h1 { font-size: 2em; font-weight: bold; margin: 0.67em 0 0.4em; }
h2 { font-size: 1.5em; font-weight: bold; margin: 0.83em 0 0.4em; }
h3 { font-size: 1.25em; font-weight: bold; margin: 1em 0 0.4em; }
h4 { font-size: 1em; font-weight: bold; margin: 1em 0 0.4em; }
h5 { font-size: 0.875em; font-weight: bold; margin: 1em 0 0.4em; }
h6 { font-size: 0.85em; font-weight: bold; margin: 1em 0 0.4em; color: #59636e; }
p { margin: 0 0 1em; }
ul, ol { margin: 0 0 1em; padding-left: 2em; }
ul { list-style-type: disc; }
ol { list-style-type: decimal; }
ul ul, ol ul { list-style-type: circle; margin-bottom: 0; }
ul ul ul, ol ul ul, ul ol ul, ol ol ul { list-style-type: square; }
ol ol, ol ul, ul ol { margin-bottom: 0; }
li { margin: 0.15em 0; }
dl { margin: 0 0 1em; }
dt { font-weight: bold; margin-top: 0.5em; }
dd { margin: 0 0 0.5em 2em; }
blockquote { margin: 0 0 1em; padding: 0 1em; border-left: 4px solid #d1d9e0; color: #59636e; }
pre { font-family: monospace; white-space: pre; background: #f6f8fa; padding: 0.8em 1em;
      margin: 0 0 1em; border-radius: 6px; }
code, kbd, samp, tt, var { font-family: monospace; }
code, kbd { background: #eff1f3; }
pre code { background: transparent; }
a { color: #0969da; text-decoration: underline; }
a:hover { color: #0550ae; }
b, strong { font-weight: bold; }
i, em, cite, dfn, var, address { font-style: italic; }
u, ins { text-decoration: underline; }
s, strike, del { text-decoration: line-through; }
small { font-size: 0.85em; }
big { font-size: 1.2em; }
sub { vertical-align: sub; font-size: 0.75em; }
sup { vertical-align: super; font-size: 0.75em; }
mark { background: #fff8c5; }
hr { border-top: 2px solid #d1d9e0; margin: 1.2em 0; }
table { margin: 0 0 1em; }
th, td { padding: 0.25em 0.8em; border-bottom: 1px solid #d1d9e0; }
th { font-weight: bold; background: #f6f8fa; text-align: left; }
caption { font-style: italic; text-align: center; }
figure { margin: 0 0 1em 2em; }
figcaption { font-style: italic; color: #59636e; }
center { text-align: center; }
`
