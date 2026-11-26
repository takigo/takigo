#!/usr/bin/env python3
"""Write docs/options.md: every widget option constructor of the classic
widgets, ttk, the canvas, the dialogs and the App, with the Tk option it
stands for.

The Tk name is derived from the Go name (the widget prefix of ttk options and
the Opt suffix dropped, lowercased), so it is wrong where takigo picked
another name; the table flags nothing, read it as a starting point.

    python3 scripts/options_doc.py > docs/options.md
"""
import glob
import os
import re

PAT = re.compile(r'^func ([A-Z]\w*)(\[[^\]]*\])?\(([^)]*)\) (\w+Option) \{', re.M)
TTK_WIDGETS = ['Button', 'Checkbutton', 'Combobox', 'Entry', 'Frame', 'Label', 'Labelframe',
               'Menubutton', 'Notebook', 'Panedwindow', 'Progressbar', 'Radiobutton', 'Scale',
               'Scrollbar', 'Separator', 'Sizegrip', 'Spinbox', 'Toggleswitch', 'Treeview',
               'Tag']  # treeview "tag configure" options


def packages():
    """The packages the table lists, as (prefix, header, directory) triples.

    The prefix is the package name the options are used with, the header
    the section title.
    """
    files = (glob.glob('widget/*/*.go') + glob.glob('ttk/*.go') + glob.glob('canvas/*.go')
             + glob.glob('geometry/*/*.go') + glob.glob('dialog/*.go'))
    dirs = sorted({os.path.dirname(f) for f in files if not f.endswith('_test.go')})
    out = [(os.path.basename(d), d, d) for d in dirs if '/internal' not in d]
    # The App's own options live in the root package.
    out.append(('takigo', 'takigo', '.'))
    return out


def tk_name(pkg, name):
    tk = name
    if pkg == 'ttk':
        for w in TTK_WIDGETS:
            if tk.startswith(w) and len(tk) > len(w):
                tk = tk[len(w):]
                break
    if tk.endswith('Opt'):
        tk = tk[:-3]
    return '-' + tk.lower()


def doc_comment(src, pos):
    """The // comment lines just above the declaration at pos."""
    lines = src[:pos].rstrip('\n').split('\n')
    out = []
    while lines and lines[-1].startswith('//'):
        out.append(lines.pop())
    return ' '.join(reversed(out))


def main():
    print('# Option reference')
    print()
    print("Every widget option of Tk is a Go function in the widget's package that returns")
    print("that widget's option type, to pass to `New` or to `Configure`. This table lists")
    print('them with the Tk option each stands for. The classic widget packages use the')
    print('bare Tk name where it is free and an `Opt` suffix otherwise; the `ttk` package')
    print('prefixes every option with the widget name. A distance option takes pixels or')
    print('a `screenunit.Distance`; a colour option a name or `color.RGB`; a font option')
    print('a name or `font.Attributes`.')
    print()
    print('The Tk column is derived from the Go name, so where takigo chose another name')
    print('(`canvas.ArcStyleOpt` is Tk\'s `-style`, `BitmapBackground` its `-background`)')
    print('it is only a hint. An option whose doc comment says "Not in Tk." has no Tk')
    print('counterpart and shows "none (takigo)". Regenerate with')
    print('`python3 scripts/options_doc.py > docs/options.md`.')
    print()
    for pkg, header, directory in packages():
        rows = []
        for f in sorted(glob.glob(directory + '/*.go')):
            if f.endswith('_test.go'):
                continue
            src = open(f).read()
            for m in PAT.finditer(src):
                name, _, params, typ = m.groups()
                tk = tk_name(pkg, name)
                if 'Not in Tk.' in doc_comment(src, m.start()):
                    tk = 'none (takigo)'
                rows.append((typ, name, params.strip(), tk))
        if not rows:
            continue
        print('## ' + header)
        print()
        print('| Option type | Go option | Takes | Tk option |')
        print('|---|---|---|---|')
        for typ, name, params, tk in sorted(rows):
            tk = tk if tk.startswith('none') else '`%s`' % tk
            print('| `%s` | `%s.%s` | `%s` | %s |' % (typ, pkg, name, params.replace('|', '\\|'), tk))
        print()


if __name__ == '__main__':
    main()
