package main
import "encoding/json"
import "net/http"
import "fmt"
type R *http.Request
type W http.ResponseWriter
type Req struct {
	Http *http.Request
}
type Reply map[string]interface{}
func (r *Req) FormValue(n string) string {
	return r.Http.FormValue(n)
}
func __(a func (r *Req) Reply) (func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) { byt, _ := json.MarshalIndent(a(&Req{r}),""," "); w.Write(byt) }
}
func RPCServer(addr string) {
	http.HandleFunc("/stat", __(func (r *Req) Reply {
		return Reply{
			"difficulty": chain.GetDifficulty(),
			"unconfirmed": unconfirmed.Length(),
			"height": chain.Height(),
			"peers": peers.Length(),
		}
	}))
	http.HandleFunc("/balance", __(func (r *Req) Reply {
		return Reply{"balance":chain.GetBalance(DecodeWalletAddress(r.FormValue("address")))}
	}))
	http.HandleFunc("/wallet/send", __(func (r *Req) Reply {
		if !HasWallet(r.FormValue("name")) {
			return Reply{"error":"no_wallet"}
		}
		wal := GetWallet(r.FormValue("name"))
		var amt float64
		fmt.Sscanf(r.FormValue("amount"), "%f", &amt)
		if wal.Balance(chain) < amt {
			return Reply{"error":"no_funds"}
		}
		wal.Send(chain, DecodeWalletAddress(r.FormValue("to")), amt)
		return Reply{"success":true}
	}))
	http.HandleFunc("/wallet/create", __(func (r *Req) Reply {
		if HasWallet(r.FormValue("name")) {
			return Reply{"error":"wallet_exists"}
		}
		wal := GenerateWallet()
		PutWallet(r.FormValue("name"), wal)
		return Reply{"address":wal.Public.String()}
	}))
	http.HandleFunc("/wallet/stat", __(func (r *Req) Reply {
		if !HasWallet(r.FormValue("name")) {
			return Reply{"error":"no_wallet"}
		}
		wal := GetWallet(r.FormValue("name"))
		return Reply{"address":wal.Public.String(),"balance":wal.Balance(chain)}
	}))
	http.ListenAndServe(addr, nil)
}
