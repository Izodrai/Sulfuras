package web

import (
	"../config/utils"
	"../tools"
	"./tmpl"
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var api_c *utils.API
var bids map[int]tools.SavedBids

func StartWebServer(b map[int]tools.SavedBids, api *utils.API) error {

	api_c = api
	bids = b

	router := httprouter.New()
	router.GET("/", Home)
	router.GET("/home", Home)
	router.GET("/symbol/:id", Symbol)
	router.GET("/stock_value/json_from_last_day/:id", JsonFromLastDay)
	router.GET("/stock_value/graph/:id/:hours", Graph)
	router.GET("/stock_value/tab/:id/:hours", Tab)

	if err := http.ListenAndServe(":8080", router); err != nil {
		return err
	}

	return nil
}

func Home(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error

	tmpl, err := template.ParseFiles(api_c.Tmpl+"home.html")
	if err != nil {
		err500(w, err)
		return
	}

	data := struct {
		ActivSymbols   []tools.Symbol
		StandbySymbols []tools.Symbol
		InactivSymbols []tools.Symbol
	}{
		ActivSymbols:   api_c.ActivSymbols_t,
		StandbySymbols: api_c.StandbySymbols_t,
		InactivSymbols: api_c.InactivSymbols_t,
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

	symbol, err = checkSymbol(w, r, ps)
	if err != nil {
		err500(w, err)
		return
	}

	tmpl, err := template.ParseFiles(api_c.Tmpl+"symbol.html")
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

func Graph(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var ok bool
	var err error
	var symbol tools.Symbol

	symbol, err = checkSymbol(w, r, ps)
	if err != nil {
		err500(w, err)
		return
	}

	var hours time.Duration

	if hours, err = time.ParseDuration(ps.ByName("hours")); err != nil {
		err500(w, err)
		return
	}

	var bids_of_s tools.SavedBids

	if bids_of_s, ok = bids[symbol.Id]; !ok {
		err500(w, errors.New("No stock values for this symbol (1)"))
		return
	}

	var from = time.Now().Add(-hours)
	var yt, mt, dt = from.Date()
	var h, m, _ = from.Clock()

	var tFrom = time.Date(yt, mt, dt, h, m, 0, 0, time.UTC)

	var data []string
	var lst_c, lst_v float64

	for _, b := range bids_of_s.SortBidsByDateAscFrom(tFrom) {

		var sma_12, sma_24 float64

		var s = "\"" + b.Bid_at.Format("2006-01-02 15:04:05") + "\"," + strconv.FormatFloat(b.Last_bid, 'f', 3, 64)

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

		s = s + "," + strconv.FormatFloat(sma_12, 'f', -1, 64)
		s = s + "," + strconv.FormatFloat(sma_24, 'f', -1, 64)
		data = append(data, s)

	}

	io.WriteString(w, tmpl.StgGraph("sma_12", "sma_24", "["+strings.Join(data, "],[")+"]"))
}

func Tab(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var ok bool
	var err error
	var symbol tools.Symbol

	symbol, err = checkSymbol(w, r, ps)
	if err != nil {
		err500(w, err)
		return
	}

	var hours time.Duration

	if hours, err = time.ParseDuration(ps.ByName("hours")); err != nil {
		err500(w, err)
		return
	}

	var bids_of_s tools.SavedBids

	if bids_of_s, ok = bids[symbol.Id]; !ok {
		err500(w, errors.New("No stock values for this symbol (1)"))
		return
	}

	var from = time.Now().Add(-hours)
	var yt, mt, dt = from.Date()
	var h, m, _ = from.Clock()

	var tFrom = time.Date(yt, mt, dt, h, m, 0, 0, time.UTC)

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

	for _, b := range bids_of_s.SortBidsByDateAscFrom(tFrom) {

		var sma_12, sma_24 float64

		var s = "\"" + b.Bid_at.Format("2006-01-02 15:04:05") + "\"," + strconv.FormatFloat(b.Last_bid, 'f', 3, 64)

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

		s = s + "," + strconv.FormatFloat(sma_12, 'f', -1, 64)
		s = s + "," + strconv.FormatFloat(sma_24, 'f', -1, 64)
		data = append(data, s)

		tdata += `
			<tr>
				<td>` + strconv.Itoa(b.Id) + `</td>
				<td>` + strconv.FormatFloat(b.Last_bid, 'f', 3, 64) + `</td>
				<td>` + b.Bid_at.Format("2006-01-02 15:04:05") + `</td>
				<td>` + strconv.FormatFloat(sma_12, 'f', 3, 64) + `</td>
				<td>` + strconv.FormatFloat(sma_24, 'f', 3, 64) + `</td>
			</tr>
		`
	}

	tdata += `
	</table>
	`

	io.WriteString(w, tmpl.StgTab("sma_12", "sma_24", tdata))
}




func JsonFromLastDay(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var ok bool
	var err error
	var symbol tools.Symbol

	symbol, err = checkSymbol(w, r, ps)
	if err != nil {
		err500(w, err)
		return
	}

	var bids_of_s tools.SavedBids

	if bids_of_s, ok = bids[symbol.Id]; !ok {
		err500(w, errors.New("No stock values for this symbol (1)"))
		return
	}

	var yt, mt, dt = time.Now().Date()
	var yesterday = time.Date(yt, mt, dt, 0, 0, 0, 0, time.UTC)

	b, err := json.MarshalIndent(bids_of_s.SortBidsByDateAscFrom(yesterday), "", "\t")
	if err != nil {
		err500(w, err)
		return
	}

	io.WriteString(w, string(b))
}

func checkSymbol(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (tools.Symbol, error) {
	var id int
	var err error
	var s_id = ps.ByName("id")

	id, err = strconv.Atoi(s_id)
	if err != nil {
		return tools.Symbol{}, err
	}

	var symbol tools.Symbol

	for _, s := range api_c.AllSymbols {
		if id == s.Id {
			symbol = s
		}
	}

	if symbol.Id == 0 {
		return tools.Symbol{}, errors.New("This symbol id do not exist")
	}

	return symbol, nil
}

func err500(w http.ResponseWriter, err error) {
	tmpl, _ := template.ParseFiles(api_c.Tmpl+"/500.html")
	tmpl.Execute(w, err.Error())
}
