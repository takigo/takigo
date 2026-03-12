#!/usr/bin/env bash
# demo_map.sh -- Mapping of Go demo names to Tcl demo names
#
# Source this file or run it directly.
#
# Usage (direct):
#   demo_map.sh                    -- print all comparable demos (go_name tcl_name)
#   demo_map.sh <go_demo_name>     -- print the Tcl name for a Go demo
#
# Each line: GO_NAME TCL_NAME
# Lines starting with # are comments.
# GO_NAME="-" means no Tcl counterpart (Go-only demo).
# TCL_NAME="-" means no Go counterpart (Tcl-only).

# ---------------------------------------------------------------------------
# DEMO_MAP associative array: go_name -> tcl_name
# "-" means no counterpart on that side
# ---------------------------------------------------------------------------
declare -A DEMO_MAP

# 1:1 same name
DEMO_MAP[anilabel]=anilabel
DEMO_MAP[aniwave]=aniwave
DEMO_MAP[arrow]=arrow
DEMO_MAP[bind]=bind
DEMO_MAP[bitmap]=bitmap
DEMO_MAP[button]=button
DEMO_MAP[check]=check
DEMO_MAP[clrpick]=clrpick
DEMO_MAP[colors]=colors
DEMO_MAP[combo]=combo
DEMO_MAP[cscroll]=cscroll
DEMO_MAP[ctext]=ctext
DEMO_MAP[dialog1]=dialog1
DEMO_MAP[dialog2]=dialog2
DEMO_MAP[entry1]=entry1
DEMO_MAP[entry2]=entry2
DEMO_MAP[entry3]=entry3
DEMO_MAP[filebox]=filebox
DEMO_MAP[floor]=floor
DEMO_MAP[fontchoose]=fontchoose
DEMO_MAP[form]=form
DEMO_MAP[goldberg]=goldberg
DEMO_MAP[hscale]=hscale
DEMO_MAP[icon]=icon
DEMO_MAP[image1]=image1
DEMO_MAP[image2]=image2
DEMO_MAP[items]=items
DEMO_MAP[knightstour]=knightstour
DEMO_MAP[label]=label
DEMO_MAP[labelframe]=labelframe
DEMO_MAP[mclist]=mclist
DEMO_MAP[menu]=menu
DEMO_MAP[menubu]=menubu
DEMO_MAP[msgbox]=msgbox
DEMO_MAP[paned1]=paned1
DEMO_MAP[paned2]=paned2
DEMO_MAP[pendulum]=pendulum
DEMO_MAP[plot]=plot
DEMO_MAP[print]=print
DEMO_MAP[puzzle]=puzzle
DEMO_MAP[radio]=radio
DEMO_MAP[ruler]=ruler
DEMO_MAP[sayings]=sayings
DEMO_MAP[search]=search
DEMO_MAP[spin]=spin
DEMO_MAP[states]=states
DEMO_MAP[style]=style
DEMO_MAP[text]=text
DEMO_MAP[textpeer]=textpeer
DEMO_MAP[toolbar]=toolbar
DEMO_MAP[tree]=tree
DEMO_MAP[ttkbut]=ttkbut
DEMO_MAP[ttkmenu]=ttkmenu
DEMO_MAP[ttknote]=ttknote
DEMO_MAP[ttkpane]=ttkpane
DEMO_MAP[ttkprogress]=ttkprogress
DEMO_MAP[ttkscale]=ttkscale
DEMO_MAP[ttkspin]=ttkspin
DEMO_MAP[twind]=twind
DEMO_MAP[unicodeout]=unicodeout
DEMO_MAP[vscale]=vscale

# Different names
DEMO_MAP[windowicons]=icon   # Go "windowicons" -> Tcl "icon"

# Go-only (no Tcl counterpart) — listed for completeness, skipped in batch
DEMO_MAP[images]=-
DEMO_MAP[msgwidget]=-
DEMO_MAP[square]=-
DEMO_MAP[systray]=-
DEMO_MAP[widget_demo]=-
DEMO_MAP[demohelper]=-

# ---------------------------------------------------------------------------
# CLI interface
# ---------------------------------------------------------------------------
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -eq 0 ]]; then
        # Print all comparable demos (go_name tcl_name)
        for go_name in $(echo "${!DEMO_MAP[@]}" | tr ' ' '\n' | sort); do
            tcl_name="${DEMO_MAP[$go_name]}"
            [[ "$tcl_name" == "-" ]] && continue
            echo "$go_name $tcl_name"
        done
    else
        go_name="$1"
        if [[ -v DEMO_MAP[$go_name] ]]; then
            echo "${DEMO_MAP[$go_name]}"
        else
            echo "-"
            exit 1
        fi
    fi
fi
