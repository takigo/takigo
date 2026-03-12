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

proc addSeeDismiss {w show {vars {}} {extra {}}} {
    ttk::frame $w
    ttk::separator $w.sep
    grid $w.sep -columnspan 4 -row 0 -sticky ew -pady 2
    ttk::button $w.dismiss -text "Dismiss" \
        -image ::img::delete -compound left \
        -command [list destroy [winfo toplevel $w]]
    ttk::button $w.code -text "See Code" \
        -image ::img::view -compound left \
        -command [list showCode $show]
    set buttons [list x $w.code $w.dismiss]
    if {[llength $vars]} {
        ttk::button $w.vars -text "See Variables" \
            -image ::img::view -compound left \
            -command [concat [list showVars $w.dialog] $vars]
        set buttons [linsert $buttons 1 $w.vars]
    }
    if {$extra ne ""} {
        set buttons [linsert $buttons 1 [uplevel 1 $extra]]
    }
    grid {*}$buttons -padx 3 -pady 3
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

proc showStatus {args} {}
proc setStatusCB {args} {}

# ---- SVG icon images (mirrors widget launcher) ------------------------------

set viewData {
    <?xml version="1.0" encoding="UTF-8"?>
    <svg width="16" height="16" version="1.1" xmlns="http://www.w3.org/2000/svg">
     <path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5 5.5 0 0 1 11 0z" fill="#000000"/>
    </svg>
}

set refreshData {
    <?xml version="1.0" encoding="UTF-8"?>
    <svg width="16" height="16" version="1.1" xmlns="http://www.w3.org/2000/svg">
     <path d="M11 5.466V4H5a4 4 0 0 0-3.584 5.777.5.5 0 1 1-.896.446A5 5 0 0 1 5 3h6V1.534a.25.25 0 0 1 .41-.192l2.36 1.966c.12.1.12.284 0 .384l-2.36 1.966a.25.25 0 0 1-.41-.192Zm3.81.086a.5.5 0 0 1 .67.225A5 5 0 0 1 11 13H5v1.466a.25.25 0 0 1-.41.192l-2.36-1.966a.25.25 0 0 1 0-.384l2.36-1.966a.25.25 0 0 1 .41.192V12h6a4 4 0 0 0 3.585-5.777.5.5 0 0 1 .225-.67Z" fill="#000000"/>
    </svg>
}

set printData {
    <?xml version="1.0" encoding="UTF-8"?>
    <svg width="16" height="16" version="1.1" xmlns="http://www.w3.org/2000/svg">
     <path d="M2.5 8a.5.5 0 1 0 0-1 .5.5 0 0 0 0 1z" fill="#000000"/>
     <path d="M5 1a2 2 0 0 0-2 2v2H2a2 2 0 0 0-2 2v3a2 2 0 0 0 2 2h1v1a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2v-1h1a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-1V3a2 2 0 0 0-2-2H5zM4 3a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1v2H4V3zm1 5a2 2 0 0 0-2 2v1H2a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1h-1v-1a2 2 0 0 0-2-2H5zm7 2v3a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1v-3a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1z" fill="#000000"/>
    </svg>
}

proc images {arg} {
    global viewData refreshData printData
    set fgColor [ttk::style lookup . -foreground {} black]
    catch {lassign [winfo rgb . $fgColor] r g b
        set fgColor [format "#%02x%02x%02x" \
            [expr {$r >> 8}] [expr {$g >> 8}] [expr {$b >> 8}]]}

    foreach action {view refresh print} {
        upvar ${action}Data imgData
        set data $imgData
        set startIdx 0
        while {[set idx1 [string first "#000000" $data $startIdx]] >= 0} {
            set idx2 [expr {$idx1 + 6}]
            set data [string replace $data $idx1 $idx2 $fgColor]
            set startIdx [expr {$idx1 + 7}]
        }
        switch $arg {
            create    { image create photo ::img::$action \
                            -format $::tk::svgFmt -data $data }
            configure { ::img::$action configure -data $data }
        }
    }
}

if {[catch {images create}]} {
    # SVG not available — create empty placeholder images
    foreach action {view refresh print} {
        image create photo ::img::$action
    }
}

image create photo ::img::delete -data {}
catch {
    image create photo ::img::delete -format $::tk::svgFmt -data {
        <?xml version="1.0" encoding="UTF-8"?>
        <svg width="16" height="16" version="1.1" xmlns="http://www.w3.org/2000/svg">
         <path d="M2.146 2.854a.5.5 0 1 1 .708-.708L8 7.293l5.146-5.147a.5.5 0 0 1 .708.708L8.707 8l5.147 5.146a.5.5 0 0 1-.708.708L8 8.707l-5.146 5.147a.5.5 0 0 1-.708-.708L7.293 8 2.146 2.854Z" fill="#d00000"/>
        </svg>
    }
}

# ---- showCode ---------------------------------------------------------------

proc evalShowCode {w} {
    set code [$w get 1.0 end-1c]
    uplevel #0 $code
}

proc printCode {w file} {
    catch {tk print $w}
}

proc showCode {show} {
    global tk_demoDirectory
    set file [string range $show 1 end].tcl
    set top .code
    if {![winfo exists $top]} {
        toplevel $top
        if {[tk windowingsystem] eq "x11"} {wm attributes $top -type dialog}

        set t [frame $top.f]
        set text [text $t.text -font fixedFont -height 24 -wrap word \
                      -yscrollcommand [list $t.yscroll set] \
                      -setgrid 1 -highlightthickness 0 -padx 3 -pady 2 \
                      -tabstyle wordprocessor]
        ttk::scrollbar $t.yscroll -command [list $t.text yview] \
            -orient vertical

        grid $t.text $t.yscroll -sticky news
        grid rowconfigure $t 0 -weight 1
        grid columnconfigure $t 0 -weight 1

        set btns [ttk::frame $top.btns]
        ttk::separator $btns.sep
        grid $btns.sep -columnspan 4 -row 0 -sticky ew -pady 2
        ttk::button $btns.dismiss -text "Dismiss" \
            -default active -command [list destroy $top] \
            -image ::img::delete -compound left
        ttk::button $btns.print -text "Print Code" \
            -command [list printCode $text $file] \
            -image ::img::print -compound left
        ttk::button $btns.rerun -text "Rerun Demo" \
            -command [list evalShowCode $text] \
            -image ::img::refresh -compound left
        grid x $btns.rerun $btns.print $btns.dismiss -padx 3 -pady 3
        grid columnconfigure $btns 0 -weight 1

        grid $t    -sticky news
        grid $btns -sticky ew
        grid rowconfigure $top 0 -weight 1
        grid columnconfigure $top 0 -weight 1

        bind $top <Return> {
            if {[winfo class %W] ne "Text"} { .code.btns.dismiss invoke }
        }
        bind $top <Escape> [bind $top <Return>]
    } else {
        wm deiconify $top
        raise $top
    }
    wm title $top "Demo code: [file join $tk_demoDirectory $file]"
    wm iconname $top $file
    set id [open [file join $tk_demoDirectory $file]]
    fconfigure $id -encoding utf-8 -eofchar "\032 {}"
    $top.f.text delete 1.0 end
    $top.f.text insert 1.0 [read $id]
    $top.f.text mark set insert 1.0
    close $id
}

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
