# export_fileicons.tcl -- rasterize Tk 9.1's X11 [tk fileicon] images
# (::tk::icons::svgPhoto, library/fileicon.tcl) at 16 and 48px to PNG, for
# embedding in demos/demohelper/icons/ (used by the tree and image2 demos).
#
#   LD_LIBRARY_PATH=tcl/unix:tk/unix xvfb-run -a tk/unix/wish scripts/export_fileicons.tcl
set out [file join [file dirname [info script]] .. demos demohelper icons]
file mkdir $out
tk fileicon / 16 ;# loads library/fileicon.tcl
foreach name {folder executable drawing image audio calendar word
        presentation spreadsheet archive html script binary font mail text} {
    foreach size {16 48} {
	set img [::tk::icons::svgPhoto $name $size]
	$img write [file join $out fileicon-$name-$size.png] -format png
    }
}
exit
