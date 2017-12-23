package require http
package require Tk 8.5
wm title . "RCoin Wallet"
wm geometry . 720x480
set node "http://127.0.0.1:3009"
set wallet "default"
if [info exists env(WALLETRPC)] {
	set node $env(WALLETRPC)
}
proc jsonDecode {json {indexVar {}}} {
    # Link to the caller's index variable.
    if {$indexVar ne {}} {
        upvar 1 $indexVar index
    }

    # By default, start decoding at the start of the input.
    if {![info exists index]} {
        set index 0
    }

    # Skip leading whitespace.  Return empty at end of input.
    if {![regexp -indices -start $index {[^\t\n\r ]} $json range]} {
        return
    }
    set index [lindex $range 0]

    # The first character determines the JSON element type.
    switch [string index $json $index] {
    \" {
        # JSON strings start with double quote.
        set type string

        # The value is the text between matching double quotes.
        if {![regexp -indices -start $index {\A\"((?:[^"]|\\.)*)\"}\
                $json range sub]} {
            return -code error "invalid JSON string at index $index:\
                    must end with close quote"
        }
        set value [string range $json {*}$sub]

        # Process all backslash substitutions in the value.
        set start 0
        while {[regexp -indices -start $start {\\u[[:xdigit:]]{4}|\\[^u]}\
                $value sub]} {
            set char [string index $value [expr {[lindex $sub 0] + 1}]]
            switch $char {
                u {set char [subst [string range $value {*}$sub]]}
                b {set char \b} f {set char \f} n {set char \n}
                r {set char \r} t {set char \t}
            }
            set value [string replace $value {*}$sub $char]
            set start [expr {[lindex $sub 0] + 1}]
        }
    } \{ - \[ {
        # JSON objects/arrays start with open brace/bracket.
        if {[string index $json $index] eq "\{"} {
            set type object
            set endRe {\A[\t\n\r ]*\}}
            set charName brace
        } else {
            set type array
            set endRe {\A[\t\n\r ]*\]}
            set charName bracket
        }
        set value {}
        incr index

        # Loop until close brace/bracket is encountered.
        while {![regexp -indices -start $index $endRe $json range]} {
            # Each element other than the first is preceded by comma.
            if {[llength $value]} {
                if {![regexp -indices -start $index\
                        {\A[\t\n\r ]*,} $json range]} {
                    return -code error "invalid JSON $type at index $index:\
                            element not followed by comma or close $charName"
                }
                set index [expr {[lindex $range 1] + 1}]
            }

            # For objects, get key and confirm it is followed by colon.
            if {$type eq "object"} {
                set key [jsonDecode $json index]
		set key [list string $key]
                if {![llength $key]} {
                    return -code error "invalid JSON object at index $index:\
                            must end with close brace"
                } elseif {[lindex $key 0] ne "string"} {
                    return -code error "invalid JSON object at index $index:\
                            key type is \"[lindex $key 0]\", must be string"
                } elseif {![regexp -indices -start $index {\A[\t\n\r ]*:}\
                        $json range]} {
                    return -code error "invalid JSON object at index $index:\
                            key not followed by colon"
                }
                set index [expr {[lindex $range 1] + 1}]
                lappend value [lindex $key 1]
            }

            # Get element value.
            lappend value [jsonDecode $json index]
        }
    } t - f - n {
        # JSON literals are true, false, or null.
        set type literal
        if {![regexp -indices -start $index {(?:true|false|null)\M}\
                $json range]} {
            return -code error "invalid JSON literal at index $index"
        }
        set value [string range $json {*}$range]
    } - - + - 0 - 1 - 2 - 3 - 4 - 5 - 6 - 7 - 8 - 9 - . {
        # JSON numbers are integers or real numbers.
        set type number
        if {![regexp -indices -start $index --\
                {-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][-+]?\d+)?\M} $json range]} {
            return -code error "invalid JSON number at index $index"
        }
        set value [string range $json {*}$range]
    } default {
        # JSON allows only the above-listed types.
        return -code error "invalid JSON data at index $index"
    }}

    # Continue decoding after the last character matched above.
    set index [expr {[lindex $range 1] + 1}]

    # When performing a full decode, ensure only whitespace appears at end.
    if {$indexVar eq {} && [regexp -start $index {[^\t\n\r\ ]} $json]} {
        return -code error "junk at end of JSON"
    }

    # Return the type and value.
    return $value
}

proc apicall {path} {
	if [catch {
	set t [http::geturl "${::node}${path}"]
	set r [http::data $t]
	http::cleanup $t
	} msg] {
		tk_messageBox -icon error -title "Error" -message "Can't connect to RCoin Service: $msg"
		exit
	}
	return [jsonDecode $r]
}
proc getwltinfo {} {
	array set r [apicall "/wallet/stat?name=$::wltname"]
	set ::wltaddr $r(address)
	set wltinfo "Address: $r(address) (click to copy)\nBalance: $r(balance)\n + NETWORK INFO + \n"
	array set x [apicall "/stat"]
	append wltinfo "Current mining difficulty: $x(difficulty)\nTotal blocks mined: $x(height)\nUnconfirmed transactions: $x(unconfirmed)"
	set ::wltinfo $wltinfo
	array set r [apicall "/wallet/history?name=$::wltname"]
	set wlthist2 {}
	foreach y $r(transactions) {
		array set q $y
		if {$q(from) eq $::wltaddr} {
			set ln "Sent $q(amount) RCN to $q(to)"
		} elseif {$q(from) eq [string repeat A 52]} {
			set ln "Mined one block with $q(amount) RCN reward"
		} else {
			set ln "Received $q(amount) RCN from $q(from)"
		}
		lappend wlthist2 $ln
	}
	set ::wlthist $wlthist2
	after 1000 getwltinfo
}
frame .login -width 256 -height 96 -bd 2 -relief raised
pack propagate .login 0
label .login.label -text "Wallet name:"
set wltname "default"
set wlts [string map [list .wallet "" "$env(HOME)/.rcoin/" ""] [glob -nocomplain -d "$env(HOME)/.rcoin" *.wallet]]
ttk::combobox .login.name -values $wlts -textvariable wltname
bind .login.name <Return> { .login.ok invoke }
ttk::button .login.ok -text "Choose wallet" -command {
	array set r [apicall "/wallet/stat?name=$wltname"]
	if [info exists r(error)] {
		tk_messageBox -icon error -title "Error" -message "No such wallet: $wltname"
		unset r(error)
	} else {
		set ::_cont 1
	}
}
pack .login.label -fill x
pack .login.name -fill x
pack .login.ok -expand yes
pack .login -expand yes
vwait ::_cont
destroy .login
ttk::notebook .tabs
frame .tabs.main
.tabs add .tabs.main -text "Overview"
pack .tabs -expand yes -fill both
getwltinfo
label .tabs.main.info -anchor nw -justify left -textvariable wltinfo -font {tkButtonFont -16}
bind .tabs.main.info <1> {
	clipboard append $::wltaddr
}
pack .tabs.main.info
listbox .tabs.main.myhistory -listvariable wlthist
pack .tabs.main.myhistory -expand yes -fill both
frame .tabs.send
.tabs add .tabs.send -text "Send"
label .tabs.send.lto -anchor nw -justify left -text "Send to:" -font {tkButtonFont -16}
entry .tabs.send.to -width 40 -textvariable sendto -font {tkButtonFont -16}
pack .tabs.send.lto
pack .tabs.send.to
label .tabs.send.lamt -anchor nw -justify left -text "Amount:" -font {tkButtonFont -16}
entry .tabs.send.amt -width 10 -textvariable sendamt -font {tkButtonFont -16}
pack .tabs.send.lamt
pack .tabs.send.amt
ttk::button .tabs.send.now -text "Send" -command {
	set m [apicall "/wallet/send?name=$::wltname&to=$sendto&amount=$sendamt"]
	array set u $m
	if [info exists u(error)] {
		unset u
		tk_messageBox -icon error -title Error -message "Couldn't send coins: Insufficient funds"
	} else {
		unset sendamt
		unset sendto
		unset u
		tk_messageBox -title Success -message "Sent coins to $sendto"
	}
}
pack .tabs.send.now
label .tabs.send.warning -anchor nw -justify left -text {
Warning: please make sure the "Send to" address is
correct otherwise your sent coins could be lost forever
} -fg red -font {tkButtonFont -16}
pack .tabs.send.warning
