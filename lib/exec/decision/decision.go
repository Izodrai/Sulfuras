package decision

import (
	"../../tools"
	"../../config/utils"

	"time"
	"math"
	"../../log"
)

func DecisionMaker(api_c *utils.API, ch_req_to_exec chan tools.Request, bids map[int]map[int]tools.Bid){

	var last_m int

	for {
		tNow := time.Now()

		var h,m,_ = tNow.Clock()

		if mod := math.Mod(float64(m-1), 5); mod == 0 && last_m != m {

			log.Info(h,m)
			log.Info("############")
			log.SkipLines(1)
			//last_h = h
			last_m = m
		}

		time.Sleep(500 * time.Millisecond)
	}
}