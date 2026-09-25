# tk_dump_tree.tcl -- Tk counterpart of internal/treedump (Go).
#
# Sourced by demo_wrapper.tcl when TAKIGO_DUMP_TREE=<file> is set. Every
# 250ms it snapshots all mapped toplevels and their descendants and rewrites
# <file> (atomically) whenever the JSON changes. Schema: see
# internal/treedump/treedump.go.

namespace eval ::takigo_dump {
    variable file $::env(TAKIGO_DUMP_TREE)
    variable last ""
    variable sample "The quick brown fox jumps over the lazy dog 0123456789"
    variable opts {-text -font -relief -borderwidth -padx -pady -width -height
        -image -compound -style -anchor -justify -wraplength -orient
        -highlightthickness -underline -state}
    variable fonts {TkDefaultFont TkTextFont TkFixedFont TkMenuFont
        TkHeadingFont TkCaptionFont TkSmallCaptionFont TkIconFont TkTooltipFont}
}

proc ::takigo_dump::str {s} {
    set map {\\ \\\\ \" \\\" \n \\n \r \\r \t \\t}
    set s [string map $map $s]
    set out ""
    foreach ch [split $s ""] {
        scan $ch %c code
        if {$code < 0x20} {
            append out [format "\\u%04x" $code]
        } else {
            append out $ch
        }
    }
    return "\"$out\""
}

proc ::takigo_dump::bool {b} { expr {$b ? "true" : "false"} }

proc ::takigo_dump::isTop {w} { expr {[winfo toplevel $w] eq $w} }

proc ::takigo_dump::allWindows {w} {
    set res [list $w]
    foreach c [winfo children $w] {
        lappend res {*}[allWindows $c]
    }
    return $res
}

proc ::takigo_dump::widget {top w index parent} {
    variable opts
    set x [expr {[winfo rootx $w] - [winfo rootx $top]}]
    set y [expr {[winfo rooty $w] - [winfo rooty $top]}]
    if {$w eq $top} { set x 0; set y 0 }
    set o {}
    foreach opt $opts {
        if {![catch {$w cget $opt} v]} {
            lappend o "[str [string range $opt 1 end]]: [str $v]"
        }
    }
    set fields [list \
        "\"path\": [str $w]" \
        "\"class\": [str [winfo class $w]]" \
        "\"toplevel\": [str $top]" \
        "\"parent\": [str $parent]" \
        "\"index\": $index" \
        "\"x\": $x" "\"y\": $y" \
        "\"w\": [winfo width $w]" "\"h\": [winfo height $w]" \
        "\"reqw\": [winfo reqwidth $w]" "\"reqh\": [winfo reqheight $w]" \
        "\"manager\": [str [winfo manager $w]]" \
        "\"mapped\": [bool [winfo ismapped $w]]" \
        "\"opts\": {[join $o {, }]}"]
    set res [list "{[join $fields {, }]}"]
    set i 0
    foreach c [winfo children $w] {
        if {[isTop $c]} continue
        lappend res {*}[widget $top $c $i $w]
        incr i
    }
    return $res
}

proc ::takigo_dump::snapshot {} {
    variable fonts
    variable sample
    set tops {}
    foreach w [allWindows .] {
        if {[isTop $w] && [winfo ismapped $w]} { lappend tops $w }
    }
    set tops [lsort $tops]
    set tl {}
    set ws {}
    foreach top $tops {
        set title ""
        catch {set title [wm title $top]}
        lappend tl "{\"path\": [str $top], \"title\": [str $title], \"w\": [winfo width $top], \"h\": [winfo height $top]}"
        lappend ws {*}[widget $top $top 0 ""]
    }
    set fl {}
    foreach f $fonts {
        if {$f ni [font names]} continue
        array set a [font actual $f]
        array set m [font metrics $f]
        lappend fl "[str $f]: {\"family\": [str $a(-family)], \"size\": $a(-size), \"weight\": [str $a(-weight)], \"slant\": [str $a(-slant)], \"ascent\": $m(-ascent), \"descent\": $m(-descent), \"linespace\": $m(-linespace), \"fixed\": [bool $m(-fixed)], \"sample\": [font measure $f $sample]}"
    }
    return "{\n \"side\": \"tcl\",\n \"toplevels\": \[\n  [join $tl ",\n  "]\n \],\n \"widgets\": \[\n  [join $ws ",\n  "]\n \],\n \"fonts\": {\n  [join $fl ",\n  "]\n }\n}\n"
}

proc ::takigo_dump::tick {} {
    variable file
    variable last
    if {[catch {snapshot} json]} {
        puts stderr "tk_dump_tree: $json"
    } elseif {$json ne $last} {
        set tmp "$file.tmp[pid]"
        set fh [open $tmp w]
        fconfigure $fh -encoding utf-8
        puts -nonewline $fh $json
        close $fh
        file rename -force $tmp $file
        set last $json
    }
    set afterCmd after
    if {[llength [info commands ::_takigo_real_after]]} {
        set afterCmd ::_takigo_real_after
    }
    $afterCmd 250 ::takigo_dump::tick
}

::takigo_dump::tick
