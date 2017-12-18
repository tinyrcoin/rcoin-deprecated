# "static link" a tcl script
# combines all static_link'd scripts and adds the proper interpreter preamble
if { [llength $argv] != 2 } {
	puts stderr "Usage: staticlink.tcl in out"
	exit 1
}
lassign $argv inf outf
set in [open $inf r]
set out [open $outf w]
set data [read $in]
close $in
puts $out "#![info nameofexecutable]"
set data [regsub -all {static_link "(.+?)"} $data "\n#include \\1\n"]
set outdata {}
foreach ln [split $data "\n"] {
	if { [string match "#include*" $ln] } {
		lappend outdata [read [open [file join [file dirname $inf] [lindex $ln 1]] r]]
	} else {
		lappend outdata $ln
	}
}
puts $out [join $outdata "\n"]
close $out
