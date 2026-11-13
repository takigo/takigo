# Known issues

Open points that were looked at and deliberately left as they are, with
the reason and what a fix would take. Bugs that are simply not fixed yet
belong in the issue tracker, not here.

## ttk.Treeview reports item errors by logging, not by returning them

Noted 2026-11-13.

`Treeview.Insert(parentID, index, opts...) string` returns only the new
item's ID. Where Tk raises an error, the Treeview item API stays silent or
logs:

- `Insert` with an explicit `ItemID` that already exists inserts nothing,
  logs a warning ("Item … already exists") to the App's logger and returns
  `""`. Tk raises "Item … already exists".
- `Insert` with an unknown parent inserts under the root. Tk raises
  "Item … not found".
- `Move` with an unknown item, or one that would make an item its own
  ancestor, does nothing. Tk raises an error for both.

The hazard is that `""` is also the root's ID, so a caller that passes the
result of a refused `Insert` on as a parent gets the children at the top
level:

```go
dir := tv.Insert("", -1, ttk.ItemID("src")) // "src" exists: returns ""
tv.Insert(dir, -1, ttk.ItemText("main.go")) // lands under the root
```

It only arises when a program supplies its own IDs and supplies one twice;
generated IDs never collide. A program that does supply IDs can guard with
`tv.Exists(id)`.

**Why it stays.** A real error needs `Insert` to return `(string, error)`,
which breaks every caller. Adding `InsertE` beside it was rejected: it
would be the only `E`-suffixed method in the library, it would fix one
case while unknown parents and `Move` stayed silent, and it would be
redundant once the API is changed properly.

**What a fix takes.** One pass over the Treeview item API at a breaking
(minor) release: `Insert` returns `(string, error)`, `Move` and the other
item calls return `error` where `ttkTreeview.c` raises one, as `Configure`
and `Notebook.TabConfigure` already do. Demos, the tutorial and
`docs/options.md` change with it.
