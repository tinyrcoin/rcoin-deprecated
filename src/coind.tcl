interp alias {} static_link {} source
static_link "libcoin.tcl"
array set loglevels {debug 5 info 4 warning 3 error 2}
set opts(-w) [file join $env(HOME) .rcoin]
array set opts {-f coind.conf -p 30009 -r 3000 -a 127.0.0.1 -l debug -b rcoin.db}
set opts(-f) [file join $opts(-w) $opts(-f)]
if {[catch {array set opts $argv}]} {
	puts {
Usage: coind ?-f coind.conf? ?-p 30009? ?-r 3000? ?-a 127.0.0.1[,etc]? ?-w /path/to/wallets? ?-b blockchain.db? ?-l loglevel?
  -f <option file name> - specify these options in a file
  -w <wallet dir> - path to store wallets
  -b <blockchain file name> - filename of blockchain
  -p <peer port> - peer-to-peer connection port
  -r <rpc port> - HTTP-RPC port
  -a <ip>[,<ip>,etc] - Allowed IP addresses for RPC
  -l <loglevel> - Logging verbosity: one of {debug info warning error severewarning fatal}
	}
exit 1
}
catch {
array set opts [read [open $opts(-f) r]]
}
set opts(-b) [file join $opts(-w) $opts(-b)]
catch { file mkdir $opts(-w) }
socket -server httpd $opts(-r)
proc reply {sock code type {extra ""}} {
	puts $sock "HTTP/1.0 $code STATUS"
	puts $sock "Content-Type: $type"
	puts $sock [join $extra "\n"]
	if {$extra ne ""} { puts $sock "" }
}
proc httpd {sock addr port} {
	global opts
	if {$addr ni [split $opts(-a)]} { close $sock; return }
	fconfigure $sock -buffering line -translation crlf
	lassign [gets $sock] method path ver
	switch -- [file root $path] {
		/stats {
			reply $sock 200 text/plain
			puts $sock "BlockchainHeight:[chain::height bc]"
			puts $sock "Peers:[llength [array names protocol::peers]]"
		}
		/balance {
			reply $sock 200 text/plain
			puts $sock "Balance:[chain::getbalance bc [resolv [file extension $path]]]"
		}
		/genwallet {
			if {[file exists [file join $opts(-w) [file extension $path]]]} {
				reply $sock 500 text/plain
				puts $sock "Error:WalletExists"
				break
			}
			reply $sock 200 text/plain
			lassign [keys::genpair] pr pu
			keys::save $pr $pu [file join $opts(-w) [file extension $path].wallet]
			set ::wallets($pu) $pr
			puts $sock "PublicKey:$pu"
			puts $sock "ShortAddress:[keys::getfp $pu]"
		}
		/peerlist {
			reply $sock 200 text/plain
			foreach c [array names protocol::peers] { puts $sock $c }
		}
	}
	close $sock
}
proc log {level msg} {
	if $::loglevels($level)<=$::loglevels($::opts(-l)) { puts stderr "[clock format [clock seconds] -format "%H:%M:%S"] $level: $msg" }
}
log info "Loading blockchain"
# do not change or your blockchain will become corrupt
set creator {308201a53082010ea003020102020101300d06092a864886f70d0101050500300e310c300a060355040313036b6579301e170d3730303130313030303030315a170d3432313231323231313731325a300e310c300a060355040313036b657930819f300d06092a864886f70d010101050003818d0030818902818100b46916773961caf09a735b3783e7eabc004b6c5a02ed2cd0b1647993e028e58f265801854fc6dfaab316d87875a96547e16573a353f201c5b3ec4a6331270ab2a6895785d4bcdd38bb2f98aeac0d302e171ce3767d22c3791d0106442dbebe90e93047d601fbd01d995ec00a2b6dda933e52b05a9e40a40052460fde54eb51b50203010001a3133011300f0603551d130101ff040530030101ff300d06092a864886f70d010105050003818100b3040aceb28add24d0d8ac453a3e5a6995853f7c1f4a2a2dd25e3954c26973880ddb3426ec2e54eda66fbb2b6430ee4e7b1117d280e18e3d959396c54deb59c18ff5e18b654372d240936af88a35c36971e459a4561d0fea2090daec6056825d2871007de1d017627df36121a85a89f35787e034847743f2c002485ae9600c32}
chain::open bc $creator $opts(-b)
protocol::init bc $opts(-p)
vwait exit
