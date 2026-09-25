# export_launcher_icons.tcl -- rasterize the widget launcher's SVG icons
# (::img::view/delete/refresh/print, created by tk/library/demos/widget) to
# PNG with Tk itself, for embedding in demos/demohelper/icons/.
#
# Run through the wrapper so the icons come from the launcher's own code:
#   LD_LIBRARY_PATH=tcl/unix:tk/unix tk/unix/wish scripts/demo_wrapper.tcl \
#       $PWD/scripts/export_launcher_icons
set out [file join [file dirname [info script]] .. demos demohelper icons]
file mkdir $out
foreach name {view delete refresh print} {
    ::img::$name write [file join $out $name.png] -format png
    puts "$name [image width ::img::$name]x[image height ::img::$name]"
}
exit
