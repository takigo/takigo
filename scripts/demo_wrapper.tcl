#!/usr/bin/env wish
# demo_wrapper.tcl -- Run a Tk demo standalone (without the widget launcher)
# Usage: wish demo_wrapper.tcl <demoname>
#
# Loads the fonts, icons and helper procs (addSeeDismiss, showVars, showCode…)
# straight from the widget launcher (tk/library/demos/widget), so the demo is
# laid out exactly as when started from the launcher, then sources
# tk/library/demos/<demoname>.tcl. Only positionWindow is replaced (to honour
# DEMO_GEOMETRY), plus the determinism hooks below.

package require tk
package require msgcat
namespace import ::msgcat::mc

set demoDir [file normalize [file join [file dirname [info script]] .. tk library demos]]
set tk_demoDirectory $demoDir
set demoName [lindex $argv 0]
# Demos like knightstour pass $argv on to their own procs; hide our argument.
set argv [lrange $argv 1 end]
set argc [llength $argv]

if {$demoName eq ""} {
    puts stderr "Usage: wish demo_wrapper.tcl <demoname>"
    exit 1
}
set demoFile [file join $demoDir ${demoName}.tcl]
if {![file exists $demoFile]} {
    puts stderr "Demo not found: $demoFile"
    exit 1
}

# If the caller set XFT_DPI in the environment, push it into the Tk resource
# database so TkDefaultFont matches what the Go side rendered with.
if {[info exists ::env(XFT_DPI)] && $::env(XFT_DPI) ne ""} {
    option add *TkDefaultFont.TkGlobalScaling [expr {$::env(XFT_DPI) / 72.0}] userDefault
}

# ---- Frozen timers (deterministic screenshots) -----------------------------
# Mirrors event.Loop.After on the Go side: with TAKIGO_FREEZE_TIMERS=1 every
# `after <ms>` with ms > 0 is dropped, so animations stay on their first frame.
# Cursor blinking is a C-level timer, so disable it through the option db.
if {[info exists ::env(TAKIGO_FREEZE_TIMERS)] && $::env(TAKIGO_FREEZE_TIMERS) eq "1"} {
    option add *insertOffTime 0 startupFile
    expr {srand(1)}
    rename after ::_takigo_real_after
    proc after {args} {
        set ms [lindex $args 0]
        if {[string is entier -strict $ms] && $ms > 0} {
            return after#frozen
        }
        uplevel 1 [list ::_takigo_real_after {*}$args]
    }
    # Wall-clock time is frozen too (demohelper.Now on the Go side).
    rename ::tcl::clock::seconds ::_takigo_real_clock_seconds
    proc ::tcl::clock::seconds {args} { return 1700000000 }
}

# ---- Launcher environment ---------------------------------------------------
# Evaluate only the top-level commands of the launcher that define the demo
# environment; the rest builds the launcher's own main window.
# Entries are literal prefixes of the command text.
set launcherKeep {
    {if {"defaultFont" ni [font names]}}
    {set viewData } {set refreshData } {set printData }
    {proc images } {images create}
    {image create photo ::img::delete }
    {proc addSeeDismiss } {proc showVars } {proc showCode }
    {proc evalShowCode } {proc printCode }
}
set fh [open [file join $demoDir widget]]
fconfigure $fh -encoding utf-8
set launcherSrc [read $fh]
close $fh
set cmd ""
foreach line [split $launcherSrc \n] {
    append cmd $line \n
    if {![info complete $cmd]} continue
    set c [string trim $cmd]
    set cmd ""
    foreach prefix $launcherKeep {
        if {[string equal -length [string length $prefix] $prefix $c]} {
            uplevel #0 $c
            break
        }
    }
}
unset launcherKeep launcherSrc cmd line c fh prefix

set widgetDemo 1
set font mainFont

proc positionWindow {w} {
    # Honour DEMO_GEOMETRY from env if set; this lets the screenshot
    # script align the Tcl window with the Go demo's takigo.Geometry(...).
    set geom "+300+300"
    if {[info exists ::env(DEMO_GEOMETRY)] && $::env(DEMO_GEOMETRY) ne ""} {
        set geom $::env(DEMO_GEOMETRY)
    }
    wm geometry $w $geom
}

proc showStatus {args} {}
proc setStatusCB {args} {}

# ---- Source the demo --------------------------------------------------------

# Hide the root window — demos create their own toplevels
wm withdraw .

if {[info exists ::env(TAKIGO_DUMP_TREE)] && $::env(TAKIGO_DUMP_TREE) ne ""} {
    source [file join [file dirname [info script]] tk_dump_tree.tcl]
}

source -encoding utf-8 $demoFile
