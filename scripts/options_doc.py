#!/usr/bin/env python3
"""Write docs/options.md: every widget option constructor of the classic
widgets, ttk and the canvas, with the Tk option it stands for.

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
               'Scrollbar', 'Separator', 'Sizegrip', 'Spinbox', 'Toggleswitch', 'Treeview']


def packages():
    files = glob.glob('widget/*/*.go') + glob.glob('ttk/*.go') + glob.glob('canvas/*.go') + glob.glob('geometry/*/*.go')
    dirs = sorted({os.path.dirname(f) for f in files if not f.endswith('_test.go')})
    return [d for d in dirs if '/internal' not in d]


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
    print('it is only a hint. Regenerate with `python3 scripts/options_doc.py > docs/options.md`.')
    print()
    for pkg in packages():
        rows = []
        for f in sorted(glob.glob(pkg + '/*.go')):
            if f.endswith('_test.go'):
                continue
            for m in PAT.finditer(open(f).read()):
                name, _, params, typ = m.groups()
                rows.append((typ, name, params.strip()))
        if not rows:
            continue
        print('## ' + pkg)
        print()
        print('| Option type | Go option | Takes | Tk option |')
        print('|---|---|---|---|')
        for typ, name, params in sorted(rows):
            print('| `%s` | `%s.%s` | `%s` | `%s` |' % (typ, os.path.basename(pkg), name, params.replace('|', '\\|'), tk_name(pkg, name)))
        print()


if __name__ == '__main__':
    main()
