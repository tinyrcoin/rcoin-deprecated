package main

import "net/http"
import "fmt"
type R *http.Request
type W http.ResponseWriter
func __(a func (W, *http.Request)) (func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) { a(W(w), r) }
}
func RPCServer(addr string) {
	http.HandleFunc("/wallet/balance", __(func (w W, r *http.Request) {
		fmt.Fprintf(w, "%d\n", chain.GetBalance(DecodeAddress(r.FormValue("address"))))
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
