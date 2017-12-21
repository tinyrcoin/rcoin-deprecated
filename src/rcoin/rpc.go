package main

import "net/http"
import "fmt"
type R *http.Request
type W http.ResponseWriter
func __(a func (W, *http.Request)) (func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) { a(W(w), r) }
}
func RPCServer(addr string) {
	http.HandleFunc("/stat", __(func (w W, r *http.Request) {
		fmt.Fprintf(w, "UnconfirmedTransactions:%d\nBlockchainHeight:%d\nPeers:%d\n", len(unconfirmed), chain.Height(), len(peers)) 
		fmt.Fprintf(w, "Difficulty:%d\n", chain.GetDifficulty())
	}))
	http.HandleFunc("/wallet/balance", __(func (w W, r *http.Request) {
		fmt.Fprintf(w, "%0.4f\n", chain.GetBalance(DecodeWalletAddress(r.FormValue("address"))))
	}))
	http.HandleFunc("/wallet/send", __(func (w W, r *http.Request) {
		if !HasWallet(r.FormValue("name")) {
			fmt.Fprintf(w, "error:NotExists\n")
			return
		}
		wal := GetWallet(r.FormValue("name"))
		var amt float64
		fmt.Sscanf(r.FormValue("amount"), "%f", &amt)
		if wal.Balance(chain) < amt {
			fmt.Fprintf(w, "error:NotEnoughFunds\n")
			return
		}
		wal.Send(chain, DecodeWalletAddress(r.FormValue("to")), amt)
	}))
	http.HandleFunc("/wallet/create", __(func (w W, r *http.Request) {
		if HasWallet(r.FormValue("name")) {
			fmt.Fprintf(w, "error:Exists\n")
			return
		}
		wal := GenerateWallet()
		PutWallet(r.FormValue("name"), wal)
		fmt.Fprintf(w, "address:%s\n", wal.Public.String())
	}))
	http.HandleFunc("/wallet/stat", __(func (w W, r *http.Request) {
		if !HasWallet(r.FormValue("name")) {
			fmt.Fprintf(w, "error:NotExists\n")
			return
		}
		wal := GetWallet(r.FormValue("name"))
		fmt.Fprintf(w, "address:%s\n", wal.Public.String())
	}))
	http.ListenAndServe(addr, nil)
}
