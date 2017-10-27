package decision

import (
	"../../config/utils"
	"../../tools"

	"../../log"
	//"math"
	//"time"
)

func DecisionMaker(api_c *utils.API, ch_req_to_exec chan tools.Request, bids map[int]tools.SavedBids, trades map[int]map[string]tools.SavedTrades, ch_symbol chan tools.Symbol) {

	/*for _, s := range api_c.Symbols {
		trades[s.Id] = make(map[string]tools.SavedTrades)

		var st tools.SavedTrades
		st.OpenTrades = make(map[int]tools.Trade)
		st.ClosedTrades = make(map[int]tools.Trade)

		trades[s.Id]["trades_6_12"] = st
		trades[s.Id]["trades_12_24"] = st
		trades[s.Id]["trades_24_48"] = st
	}*/

	for symbol := range ch_symbol {

		log.Info(symbol, ", last insert : ", bids[symbol.Id].LastDate)


		/*var tNow = time.Now()

		log.Info(tNow)

		var y, mo, d = tNow.Date()
		var h, m, _ = tNow.Clock()

		if mod := math.Mod(float64(m), 5); mod != 0 {
			for math.Mod(float64(m), 5) != 0 {
				tNow = tNow.Add(-1 * time.Minute)
				_, m, _ = tNow.Clock()
			}
		}

		var tLast = time.Date(y,mo,d,h,m,0,0,time.UTC)

		log.Info(tLast)

		if val, ok := bids[symbol.Id].ByDate[tLast]; ok {
			log.Info(val)
		} else {
			log.Info("not exist...")
		}*/


		/*var sma6, sma12, sma24, sma48 = 0,0,0,0
		var diff_6_12, diff_12_24, diff_24_48 = 0,0,0

		if v, ok := b.Calculations["sma_6"]; !ok {
			continue
		} else {
			sma6 = v
		}

		if v, ok := b.Calculations["sma_12"]; !ok {
			continue
		} else {
			sma12 = v
		}

		if v, ok := b.Calculations["sma_24"]; !ok {
			continue
		} else {
			sma24 = v
		}

		if v, ok := b.Calculations["sma_48"]; !ok {
			continue
		} else {
			sma48 = v
		}

		diff_6_12 = sma6 - sma12
		diff_12_24 = sma12 - sma24
		diff_24_48 = sma24 - sma48*/

	}
}
