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
	historyfunc := __(func(r *Req) Reply {
		addr := DecodeWalletAddress(r.FormValue("address") + r.FormValue("name"))
		limit := 15
		fmt.Sscanf(r.FormValue("limit"), "%d", &limit)
		return Reply{"transactions":chain.History(addr,limit)}
	})
	http.HandleFunc("/tx/info", __(func (r *Req) Reply {
		blkid := r.FormValue("id")
		blk := chain.GetTransaction(StringToAddress(blkid))
		return Reply{"tx":blk}
	}))
	http.HandleFunc("/history", historyfunc)
	http.HandleFunc("/stat", __(func (r *Req) Reply {
		return Reply{
			"peers": -1,
			"difficulty": GetDifficulty(),
		}
	}))
	http.HandleFunc("/balance", __(func (r *Req) Reply {
		return Reply{"balance":AmountToFloat(chain.GetBalance(DecodeWalletAddress(r.FormValue("address"))))}
	}))
	http.HandleFunc("/wallet/history", historyfunc)
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
		wal.Send(chain, DecodeWalletAddress(r.FormValue("to")), amt, r.FormValue("comment"))
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
	http.HandleFunc("/mining/stop", __(func (r *Req) Reply {
		pausemining = true
		return Reply{"mining":false}
	}))
	http.HandleFunc("/mining/start", __(func (r *Req) Reply {
		pausemining = false
		return Reply{"mining":true}
	}))
	http.ListenAndServe(addr, nil)
}
