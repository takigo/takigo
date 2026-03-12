#!/usr/bin/env wish
# demo_wrapper.tcl -- Run a Tk demo standalone (without the widget launcher)
# Usage: wish demo_wrapper.tcl <demoname>
#
# Sets up the same font/proc environment that the widget launcher provides,
# then sources tk/library/demos/<demoname>.tcl

package require tk

# ---- Font setup (mirrors widget launcher logic) ----------------------------
if {"TkDefaultFont" in [font names]} {
    font create mainFont   {*}[font configure TkDefaultFont]
    font create fixedFont  {*}[font configure TkFixedFont]
    font create boldFont   {*}[font configure TkDefaultFont] -weight bold
    font create titleFont  {*}[font configure TkDefaultFont] -weight bold
    font create statusFont {*}[font configure TkDefaultFont]
    font create varsFont   {*}[font configure TkDefaultFont]
} else {
    font create mainFont   -family Helvetica -size 12
    font create fixedFont  -family Courier   -size 10
    font create boldFont   -family Helvetica -size 12 -weight bold
    font create titleFont  -family Helvetica -size 18 -weight bold
    font create statusFont -family Helvetica -size 10
    font create varsFont   -family Helvetica -size 14
}
set font mainFont

# ---- Required globals -------------------------------------------------------
set widgetDemo 1

# Stub for msgcat (demos don't use it heavily when run standalone)
if {[catch {package require msgcat}] || !([namespace exists ::msgcat])} {
    proc mc {str args} { return $str }
} else {
    namespace import -force ::msgcat::mc
}

# ---- Helper procs (stubs matching widget launcher API) ---------------------

proc positionWindow {w} {
    wm geometry $w +300+300
}

# Minimal addSeeDismiss: just a Dismiss button (no SVG icons needed)
proc addSeeDismiss {w show {vars {}} {extra {}}} {
    ttk::frame $w
    ttk::separator $w.sep
    grid $w.sep -columnspan 3 -row 0 -sticky ew -pady 2
    ttk::button $w.dismiss -text "Dismiss" \
        -command [list destroy [winfo toplevel $w]]
    if {[llength $vars]} {
        ttk::button $w.vars -text "See Variables" \
            -command [concat [list showVars $w.dialog] $vars]
        grid x $w.vars $w.dismiss -padx 3 -pady 3
    } else {
        grid x $w.dismiss -padx 3 -pady 3
    }
    grid columnconfigure $w 0 -weight 1
    return $w
}

proc showVars {w args} {
    catch {destroy $w}
    toplevel $w
    wm title $w "Variable values"
    set b [ttk::frame $w.frame]
    grid $b -sticky news
    set f [ttk::labelframe $b.title -text "Variable values:"]
    foreach var $args {
        ttk::label $f.n$var -text "$var:" -anchor w
        ttk::label $f.v$var -textvariable $var -anchor w
        grid $f.n$var $f.v$var -padx 2 -pady 2 -sticky w
    }
    ttk::button $b.ok -text "OK" \
        -command [list destroy $w] -default active
    bind $w <Return> [list $b.ok invoke]
    bind $w <Escape> [list $b.ok invoke]
    grid $f -sticky news -padx 3
    grid $b.ok -sticky e -padx 3 -pady 4
    grid columnconfigure $b 0 -weight 1
    grid rowconfigure $b 0 -weight 1
}

proc showCode {show} {}
proc showStatus {args} {}
proc images {args} {}
proc setStatusCB {args} {}

# ---- Source the demo --------------------------------------------------------
set demoDir [file normalize [file join [file dirname [info script]] .. tk library demos]]
set tk_demoDirectory $demoDir
set demoName [lindex $argv 0]

if {$demoName eq ""} {
    puts stderr "Usage: wish demo_wrapper.tcl <demoname>"
    exit 1
}

set demoFile [file join $demoDir ${demoName}.tcl]
if {![file exists $demoFile]} {
    puts stderr "Demo not found: $demoFile"
    exit 1
}

# Hide the root window — demos create their own toplevels
wm withdraw .

source -encoding utf-8 $demoFile
