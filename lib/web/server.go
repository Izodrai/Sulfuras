package web

import (
	"net/http"
	"../tools"
	"html/template"
	"../config/utils"
	"github.com/julienschmidt/httprouter"
	"strconv"
	"errors"
	"encoding/json"
	"io"
	"time"
	"./tmpl"
	"strings"
	//"../log"
)

var api_c *utils.API
var bids map[int]map[int]tools.Bid

func StartWebServer(b map[int]map[int]tools.Bid, api *utils.API){

	api_c = api
	bids = b

	router := httprouter.New()
	router.GET("/home", Home)
	router.GET("/symbol/:id", Symbol)
	router.GET("/stock_value/from_last_day/:id", SvFromLastDay)

	router.GET("/test/:id", Test)
	//router.GET("/test2/:id", Test2)

	http.ListenAndServe(":8080", router)

}

func Home(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error

	tmpl, err := template.ParseFiles("lib/web/tmpl/home.html")
	if err != nil {
		err500(w, err)
		return
	}

	data := struct {
		ActivSymbols []tools.Symbol
		StandbySymbols []tools.Symbol
		InactivSymbols []tools.Symbol
	}{
		ActivSymbols: api_c.ActivSymbols,
		StandbySymbols: api_c.StandbySymbols,
		InactivSymbols: api_c.InactivSymbols,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		err500(w, err)
		return
	}
}

func Symbol(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var err error
	var symbol tools.Symbol

	symbol, err = checkSymbol(w,r,ps)
	if err != nil {
		err500(w, err)
		return
	}

	tmpl, err := template.ParseFiles("lib/web/tmpl/symbol.html")
	if err != nil {
		err500(w, err)
		return
	}

	data := struct {
		Symbol tools.Symbol
	}{
		Symbol: symbol,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		err500(w, err)
		return
	}
}

func Test(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var ok bool
	var err error
	var symbol tools.Symbol

	symbol, err = checkSymbol(w,r,ps)
	if err != nil {
		err500(w, err)
		return
	}

	var bids_of_s map[int]tools.Bid

	if bids_of_s, ok = bids[symbol.Id]; !ok {
		err500(w, errors.New("No stock values for this symbol (1)"))
		return
	}

	var last_bids []tools.Bid

	var yt,mt,dt = time.Now().Date()
	var yesterday = time.Date(yt,mt,dt,0,0,0,0, time.UTC)

	var id_max int

	for _,b := range bids_of_s {
		if b.Id > id_max {
			id_max = b.Id
		}
	}

	for i := 0; i <= id_max; i++ {
		if b, ok := bids_of_s[i]; ok {
			if b.Bid_at.After(yesterday) {
				last_bids = append(last_bids, b)
			}
		}
	}

	var data []string
	var tdata string
	var lst_c, lst_v float64

	tdata += `
	<table style="border-style:solid">
	<tr>
		<td style="border-style:solid;border-width:0px 0px 2px 0px">Id</td>
		<td style="border-style:solid;border-width:0px 0px 2px 0px">At</td>
		<td style="border-style:solid;border-width:0px 0px 2px 0px">Value</td>
		<td style="border-style:solid;border-width:0px 0px 2px 0px">Sma12</td>
		<td style="border-style:solid;border-width:0px 0px 2px 0px">Sma24</td>
	</tr>
	`

	for _,b := range last_bids {

		var sma_12, sma_24 float64

		var s = "\""+b.Bid_at.Format("2006-01-02 15:04:05") + "\"," + strconv.FormatFloat(b.Last_bid, 'f', -1, 64)

		if val, ok := b.Calculations["sma_12"]; ok {
			sma_12 = val
			lst_c = val
		} else {
			sma_12 = lst_c
		}
		if val, ok := b.Calculations["sma_24"]; ok {
			sma_24 = val
			lst_v = val
		} else {
			sma_24 = lst_v
		}

		s = s + ","+ strconv.FormatFloat(sma_12, 'f', -1, 64)
		s = s + ","+ strconv.FormatFloat(sma_24, 'f', -1, 64)
		data = append(data, s)

		tdata += `
			<tr>
				<td>`+strconv.Itoa(b.Id)+`</td>
				<td>`+strconv.FormatFloat(b.Last_bid, 'f', -1, 64)+`</td>
				<td>`+b.Bid_at.Format("2006-01-02 15:04:05")+`</td>
				<td>`+strconv.FormatFloat(sma_12, 'f', -1, 64)+`</td>
				<td>`+strconv.FormatFloat(sma_24, 'f', -1, 64)+`</td>
			</tr>
		`
	}

	tdata += `
	</table>
	`

	io.WriteString(w, tmpl.Stg("sma_12","sma_24","["+strings.Join(data,"],[")+"]", tdata))
}

/*
func Test2(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var ok bool
	var err error
	var symbol tools.Symbol

	symbol, err = checkSymbol(w,r,ps)
	if err != nil {
		err500(w, err)
		return
	}

	var bids_of_s map[int]tools.Bid

	if bids_of_s, ok = bids[symbol.Id]; !ok {
		err500(w, errors.New("No stock values for this symbol (1)"))
		return
	}

	var last_bids []tools.Bid

	var yt,mt,dt = time.Now().Date()
	var yesterday = time.Date(yt,mt,dt,0,0,0,0, time.UTC)

	var id_max int

	for _,b := range bids_of_s {
		if b.Id > id_max {
			id_max = b.Id
		}
	}

	for i := 0; i <= id_max; i++ {
		if b, ok := bids_of_s[i]; ok {
			if b.Bid_at.After(yesterday) {
				last_bids = append(last_bids, b)
			}
		}
	}

	var data []string
	for _,b := range last_bids {
		data = append(data, "\""+b.Bid_at.Format("2006-01-02 15:04:05") + "\"," + strconv.FormatFloat(b.Last_bid, 'f', -1, 64))
	}
	//"[0, 0, 0], [1, 10, 5], [2, 23, 15], [3, 17, 9], [4, 18, 10], [5, 9, 5]"


	tmpl, err := template.ParseFiles("lib/web/tmpl/stock_value_graph.html")
	if err != nil {
		err500(w, err)
		return
	}

	d := struct {
		Values string
	}{
		Values: "["+strings.Join(data,"],[")+"]",
	}

	err = tmpl.Execute(w, d)
	if err != nil {
		err500(w, err)
		return
	}

}

type test_s struct{
	Dates string
	Values float64
}*/


func SvFromLastDay(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var ok bool
	var err error
	var symbol tools.Symbol

	symbol, err = checkSymbol(w,r,ps)
	if err != nil {
		err500(w, err)
		return
	}

	var bids_of_s map[int]tools.Bid

	if bids_of_s, ok = bids[symbol.Id]; !ok {
		err500(w, errors.New("No stock values for this symbol (1)"))
		return
	}

	var last_bids []tools.Bid

	var yt,mt,dt = time.Now().Date()
	var yesterday = time.Date(yt,mt,dt,0,0,0,0, time.UTC)

	var id_max int

	for _,b := range bids_of_s {
		if b.Id > id_max {
			id_max = b.Id
		}
	}

	for i := 0; i <= id_max; i++ {
		if b, ok := bids_of_s[i]; ok {
			if b.Bid_at.After(yesterday) {
				last_bids = append(last_bids, b)
			}
		}
	}

	b, err := json.MarshalIndent(last_bids, "", "\t")
	if err != nil {
		err500(w, err)
		return
	}

	io.WriteString(w, string(b))
}


func checkSymbol(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (tools.Symbol,error) {
	var id int
	var err error
	var s_id = ps.ByName("id")

	id, err = strconv.Atoi(s_id)
	if err != nil {
		return tools.Symbol{}, err
	}

	var symbol tools.Symbol

	for _,s := range api_c.AllSymbols {
		if id == s.Id {
			symbol = s
		}
	}

	if symbol.Id == 0 {
		return tools.Symbol{}, errors.New("This symbol id do not exist")
	}

	return symbol, nil
}

func err500(w http.ResponseWriter, err error){
	tmpl,_ := template.ParseFiles("lib/web/tmpl/500.html")
	tmpl.Execute(w, err.Error())
}