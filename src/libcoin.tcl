package require sqlite3
package require pki
package require sha1
package require picoirc
package require http
namespace eval keys {
	proc genpair {} {
		set key [pki::rsa::generate 1024]
		set crt [pki::x509::create_cert [pki::pkcs::parse_csr [pki::pkcs::create_csr $key {CN key}]] [list {*}$key subject CN=key] 1 1 [expr [clock seconds]+(25*365*86400)] 1 {}]
		return [list [binary encode hex [pki::key $key "" 0]] [binary encode hex $crt]]
	}
	proc getfp {crt} {
		set ret $crt
		catch { set ret [sha1::sha1 -hex [binary decode hex $crt]] }
		return $ret
	}
	proc verify {sig txt key} {
		pki::verify [binary decode hex $sig] $txt [pki::x509::parse_cert [binary decode hex $key]]
	}
	proc sign {txt key} {
		binary encode hex [pki::sign $txt [pki::pkcs::parse_key [binary decode hex $key]]]
	}
	proc save {priv pub path} {
		set f [open $path w]
		puts $f $priv
		puts $f $pub
		close $f
	}
	proc load {path} {
		set f [open $path r]
		set s [read $f]
		close $f
		return $s
	}
}
set ::st [clock seconds][clock clicks]
namespace eval misc {
	proc calcfee {unclaimed size} {
		expr {1+($size/512)*($unclaimed/128)}
	}
	proc transfer {priv pub to amt} {
		if {[chain::getbalance $protocol::bc $pub] < $amt} return
		protocol::_broadcast [list trans [chain::transaction $protocol::bc $priv $pub $to $amt]]
	}
}
proc warn {msg} {
	puts stderr "Warning: $msg"
}
set NETCFGUTIL "ifconfig"
if { $tcl_platform(platform) eq "windows" } { set NETCFGUTIL "ipconfig" }
 proc getIp {{target www.google.com} {port 80}} {
     set s [socket $target $port]
     set res [fconfigure $s -sockname]
     close $s
     lindex $res 0
 }
namespace eval protocol {
	variable bc
	variable behind 0
	variable unclaimed
	variable peers
	variable irctoks
	variable meport 30009
	variable isnat 0
	variable menick "*"
	variable meip [lindex [http::data [http::geturl http://ipv4.icanhazip.com]] 0]
	variable ircbootstrap {irc://localhost/#coinstrap irc://irc.umbrellix.net/#coinstrap}
	proc init {chain {port 30009}} {
		set protocol::bc $chain
		set protocol::meport $port
		socket -server protocol::_newpeer $port
		set isnat [expr ![string match "*$protocol::meip*" [exec $::NETCFGUTIL]]]
		if {$isnat} {
			set isnat [catch {exec upnpc -r $port tcp >@stderr 2>@stderr}]
		}
		foreach {addr port} [$chain eval {SELECT * FROM peers}] {
			after 1 [list ::protocol::addpeer $addr $port]
		} 
		bootstrap
	}
	proc addpeer {ip port} {
		if {[catch {
		set new [socket $ip $port]}] && $ip eq $protocol::meip } { set new [socket 127.0.0.1 $port] }
		if { [info exists new] } {
		_newpeer $new $ip 0
		}
	}
	proc _broadcast msg {
		foreach {k v} [array get protocol::peers] {
			puts [lindex $v 0] $msg
		}
	}
	proc _peerin {sock addr port} {
		variable bc
		if [eof $sock] {
			fileevent $sock readable {}
			catch { close $sock }
			unset protocol::peers($addr:$port)
			return
		}
		if [fblocked $sock] return
		gets $sock msg
		lassign $protocol::peers($addr:$port) -> name
		lassign $msg cmd arg1 arg2
		switch -- $cmd {
			peername {
				puts $sock "me [info hostname]:$::st"
			}
			me {
				if {$arg1 eq "[info hostname]::$::st"} {
					warn "I can't connect to myself"
					return
				}
				lset protocol::peers($addr:$port) 1 $arg1
			}
			addpeer {
				if {[$bc eval {SELECT addr FROM peers WHERE addr = @arg1}] ne ""} return
				$bc eval {INSERT INTO peers VALUES(@arg1, $arg2)}
				if {[llength [array names protocol::peers]] < 256 && ![info exists protocol::peers($arg1:$arg2)]} {
					addpeer $arg1 $arg2
				}
				foreach {k v} [array get protocol::peers] {
					lassign $v sock nick
					if {$nick ne $name && $k ne "$addr:$port"} {
						puts $sock $msg
					}
				}
			}
			trans {
				if {[$bc eval {SELECT hash FROM blocks WHERE hash = @arg1}] ne {}} {
					warn "I already have this transaction."
					return
				}
				if ![chain::verifytransaction $bc $arg1] {
					warn "Dropped transaction from $name: failed verification"
					return
				}
				chain::addblock $bc {*}[chain::decoderaw $arg1]
				foreach {k v} [array get protocol::peers] {
					lassign $v sock nick
					if {$nick ne $name && $k ne "$addr:$port"} {
						puts $sock $msg
					}
				}
			}
			height {
				if {$arg1 > $protocol::behind} {set protocol::behind $arg1}
			}
			sync {
				puts $sock "height [chain::height $bc]"
				$bc eval {SELECT * FROM blocks WHERE seq > $arg1} values {
					puts $sock [list trans [chain::encodeblock [list $values(seq) $values(hash) $values(time) $values(lasthash) $values(idfrom) $values(idto) $values(amount) $values(signature)]]]
					update
				}
			}
		}
	}
	proc _newpeer {sock addr port} {
		set protocol::peers($addr:$port) [list $sock $sock]
		fconfigure $sock -buffering line -translation crlf
		puts $sock "peername"
		puts $sock "me [info hostname]:$::st"
		puts $sock "sync [chain::height $protocol::bc]"
		fileevent $sock readable [list protocol::_peerin $sock $addr $port]
	}
	proc bootstrap {} {
		set protocol::menick "[info hostname]|$protocol::isnat|[expr int(rand()*1000)]"
		foreach url $protocol::ircbootstrap {
		lappend irctoks [picoirc::connect ::protocol::_cb $protocol::menick $url]
		}
	}
	proc _addirc {token user} {
		picoirc::post $token $user ".getpeerinfo"
		after 500 {set ::_ 1}
		vwait ::_
		unset ::_
	}
	proc _cb {token state args} {
		switch $state {
			userlist {
				foreach u [lrange [lindex $args 1] end-7 end] {
					if {[string match *|0|* $u]&&$u ne $protocol::menick} { _addirc $token $u }
				}
			}
			chat {
				puts $args
				lassign [lindex $args 2] cmd opt opt2
				switch -- $cmd {
					.getpeerinfo {
						picoirc::post $token [lindex $args 1] ".addpeer $protocol::meip $protocol::meport"
					}
					.addpeer {
						variable bc
						set arg1 $opt ; set arg2 $opt2
						if {[$bc eval {SELECT addr FROM peers WHERE addr = @arg1}] ne ""} return
						$bc eval {INSERT INTO peers VALUES(@arg1, $arg2)}

						addpeer $opt $opt2
					}
				}
			}
		}
	}
}
namespace eval chain {
	proc open {name genesis {file ":memory:"}} {
		if [catch { sqlite3 $name $file -create 0 }] {
			sqlite3 $name $file
			$name eval {CREATE TABLE addresses (ID text, KEY text)}
			$name eval {CREATE TABLE peers(addr text, port int)}
			$name eval {CREATE TABLE blocks (seq int, hash text, time int, lasthash text, idfrom text, idto text, amount int, signature text)}
			addblock $name 0 0 1 0 0 0 0 0
			addblock $name 1 0 1 0 0 $genesis 1000 0
		}
	}
	proc addaddress {name key} {
		set o [keys::getfp $key]
		$name eval {INSERT INTO addresses VALUES(@o,@key)}
	}
	proc resolveaddress {name fp} {
		return [$name eval {SELECT KEY FROM addresses WHERE ID = @fp}]
	}
	proc findto {name to} {
		set x [$name eval {SELECT amount FROM blocks WHERE idto = @to}]
	}
	proc findfrom {name from} {
		set x [$name eval {SELECT amount FROM blocks WHERE idfrom = @from}]
	}
	proc height {name} {
		$name eval {SELECT seq FROM blocks ORDER BY seq DESC LIMIT 1}
	}
	proc getbalance {name id} {
		set a [findfrom $name $id]
		set b [findto $name $id]
		set amt 0
		foreach c $b { incr amt $c; }
		foreach c $a { incr amt [expr {$c * -1}]; }
		return $amt
	}
	proc addblock {name num hash time lasthash from to amount sig} {
		$name eval {INSERT INTO blocks VALUES($num,$hash,$time,$lasthash,@from,@to,$amount,$sig)}
	}
	proc transaction {name priv from to amt} {
		set blk [list [expr [$name eval {SELECT seq FROM blocks ORDER BY seq DESC LIMIT 1}]+1] 0 [clock seconds] \
			[$name eval {SELECT hash FROM blocks ORDER BY seq LIMIT 1}] $from $to $amt 0]
		set sigblk [signblock $priv $from $blk]
		addblock $name {*}$sigblk
		encodeblock $sigblk
	}
	proc getblock {name id} {
		$name eval {SELECT * FROM blocks WHERE seq = $id}
	}
	proc encodeblock {block} {
		join $block "|"
	}
	proc decoderaw {raw} {
		split $raw "|"
	}
	proc verifyraw {raw} {
		set b [decoderaw $raw]
		lassign $b seq hash time lasthash from to amount sig
		lset b 1 0
		lset b 7 0
		set k [encodeblock $b]
		if {[sha1::sha1 -hex $k] ne $hash} { return 0 }
		if {![keys::verify $sig $k $from]} { return 0 }
		return 1
	}
	proc calcdiff {name seq} {
		set lastm [$name eval \
		{SELECT seq FROM blocks WHERE length(idfrom) = 40 ORDER BY seq DESC LIMIT 1}]
		if {$lastm eq ""} {set diff 3} else {
		set diff [expr {($seq-$lastm)/32}]}
		if {$diff > 20} {set diff 20}
		return $diff
	}
	proc verifytransaction {name raw} {
		if {![verifyraw $raw]} { return 0 }
		set b [decoderaw $raw]
		set amt [lindex $b 6]
		set id [lindex $b 4]
		set lh [lindex $b 3]
		set seqn [expr {[lindex $b 0]-1}]
		set seqa [lindex $b 0]
		if {[$name eval {SELECT seq FROM blocks WHERE seq > $seqa}] ne {}} { return 0 }
		if {[$name eval {SELECT seq FROM blocks WHERE seq = $seqn}] eq {}} { return 0 }
		if {[$name eval {SELECT hash FROM blocks WHERE seq = $seqn}] ne $lh} { return 0 }
		if {[string length $id] == 40} {
			set diff [calcdiff $name [lindex $b 0]]
			if { [string range $id 0 [expr {$diff - 1}]] eq [string repeat 0 $diff]} {
				return 1
			}
			return 0
		}
		if {$amt > [getbalance $name $id]} {
			return 0
		}
		return 1
	}
	proc signblock {key pubkey block} {
		lassign $block seq hash time lasthash from to amount sig
		lset block 1 0
		lset block 7 0
		set blkraw [encodeblock $block]
		set hash [sha1::sha1 -hex $blkraw]
		set sig [keys::sign $blkraw $key]
		set from $pubkey
		return [list $seq $hash $time $lasthash $from $to $amount $sig]
	}
	proc close {name} {
		$name close
	}
}
namespace eval hash {
	variable fast 1
	variable hashers {}
	variable hasherpath [pwd]/hashaccel
	proc _takehasher {} {
		if {[llength $hash::hashers] == 0} {
			set ret [open "|$hash::hasherpath" r+]
			fconfigure $ret -buffering none
			return $ret
		}
		set ret [lindex $hash::hashers 0]
		set hash::hashers [lrange $hash::hashers 1 end]
		return $ret
	}
	proc _releasehasher {inst} {
		lappend hash::hashers $inst
	}
	proc mine {diff data {statscb ""}} {
		if {$hash::fast} {
			set inst [_takehasher]
			if {$statscb eq ""} {
			puts $inst "M[string length $data]"
			puts $inst "$diff"
			puts $inst "$data"
			flush $inst
			fileevent $inst readable [string map "%% $inst" {
				set ::_l%% [gets %%]
			}]
			vwait ::_l$inst
			set ret [split [set ::_l$inst] " "]
			unset ::_l$inst
			_releasehasher $inst
			return $ret
			} else {
			puts $inst "m[string length $data]"
			puts $inst "$diff"
			puts $inst "$data"
			flush $inst
			fileevent $inst readable [string map "%%% $statscb %% $inst" {
				set ::_q%% [gets %%]
				if { [string match "*s*" $::_q%%] } {
					%%% $::_q%%
					unset ::_q%%
				} else {
					set ::_l%% $::_q%%
					unset ::_q%%
				}
			}]
			vwait ::_l$inst
			set ret [split [set ::_l$inst] " "]
			unset ::_l$inst
			_releasehasher $inst
			return $ret
			}
		}
	}
	proc hash {seed data} {
		if {$hash::fast} {
			set inst [_takehasher]
			puts $inst "H[string length $data]"
			puts $inst "$seed"
			puts $inst "$data"
			flush $inst
			set ret [gets $inst]
			_releasehasher $inst
			return $ret
		}
	}
}
