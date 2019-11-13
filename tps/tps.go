package tps

import (
	"errors"
	"time"
)

//IsEnabled exported
func IsEnabled() bool {
	if tpsmap != nil {
		return true
	}
	return false
}

//Init exported
func Init(chcap int, mapcln time.Duration) error {

	if chcap < 3 {
		return errors.New("A minimum of 3 calls/chcap required to calculate TPS")
	} else if mapcln < 5 {
		return errors.New("TPS data clean up cycles must be 5 seconds or longer")
	} else {
		n = chcap
		mcl = mapcln
		tpsmap = make(TPSmap)
		cleanUp()
		return nil
	}
}

//TPS exported
type TPS struct {
	ch chan time.Time
	ts time.Time
	op *time.Time
}

//TPSmap exported
type TPSmap map[int64]*TPS

var tpsmap TPSmap
var n int             //requests qty to measure TPS... this is the channel capacity
var mcl time.Duration //tps map clean up cycle length in seconds

//Transaction exported
func Transaction(id int64, tpsMax float32) *time.Time {

	now := time.Now()
	if _, ok := tpsmap[id]; ok {
		if len(tpsmap[id].ch) == cap(tpsmap[id].ch) {
			tpsmap[id].setOp(<-tpsmap[id].ch, now, tpsMax)
			//fmt.Println(tpsmap[id].setOp(<-tpsmap[id].ch, now, tpsMax))
		}
	} else {
		tps := new(TPS)
		tps.ch = make(chan time.Time, n)
		tpsmap[id] = tps
	}
	tpsmap[id].ch <- now
	tpsmap[id].ts = now
	return tpsmap[id].op
}

//return actual current tps and
//sets tps.op to block transactions if max tps limit is exceeded
func (tps *TPS) setOp(tch time.Time, now time.Time, tpsMax float32) float32 {
	t := float32(now.UnixNano()-tch.UnixNano()) / float32(time.Second)
	if tr := float32(n) / tpsMax; t < tr {
		t2 := tps.ts.Add(time.Second * time.Duration(tr-t))
		tps.op = &t2
	} else {
		tps.op = nil
	}
	return float32(n) / t
}

func cleanUp() {
	go func(m TPSmap) {
		for {
			for k, v := range m {
				if v.ts.Before(time.Now().Add(-1*mcl*time.Second)) &&
					(v.op == nil || v.op.Before(time.Now())) {
					delete(m, k)
				}
				//fmt.Println(k, v.ts, v.op)
			}
			time.Sleep(mcl * time.Second)
			//fmt.Println("map len ", len(m))
		}
	}(tpsmap)
}
