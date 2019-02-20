/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/rpc"
)

func init() {
	cfgPath := Flag{
		Name:  "config",
		Must:  false,
		Usage: "config file path",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "performance",
		Help: "get performance time",
		Func: func(c *ishell.Context) {
			args := []Flag{cfgPath}
			if HelpText(c, args) {
				return
			}
			if err := CheckArgs(c, args); err != nil {
				Warn(err)
				return
			}
			cfgPathP := StringVar(c.Args, cfgPath)
			err := getPerformanceTime(cfgPathP)
			if err != nil {
				Warn(err)
			}
		},
	}
	shell.AddCmd(c)
}

func getPerformanceTime(cfgPathP string) error {
	if cfgPathP == "" {
		cfgPathP = config.DefaultDataDir()
	}

	path1 := filepath.Join(cfgPathP, "performanceTime.json")
	_ = os.Remove(path1)

	fd1, err := os.OpenFile(path1, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	path2 := filepath.Join(cfgPathP, "performanceTime_orig.json")
	_ = os.Remove(path2)

	fd2, err := os.OpenFile(path2, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	var consensus []int64
	var start []int64
	var end []int64
	var fullConsensus []int64

	loc, _ := time.LoadLocation("Asia/Shanghai")

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()
	var pts []*types.PerformanceTime
	err = client.Call(&pts, "ledger_performance")
	if err != nil {
		return err
	}
	if len(pts) == 0 {
		return nil
	}

	for _, p := range pts {
		_, _ = fd2.WriteString(fmt.Sprintf("%s\n", p.String()))
		if p.T0 == 0 || p.T1 == 0 || p.T2 == 0 || p.T3 == 0 || p.Hash.IsZero() {
			//_, _ = fd1.WriteString(fmt.Sprintf("invalid data %s \n", p.String()))
		} else {
			t0 := time.Unix(0, p.T0)
			t1 := time.Unix(0, p.T1)
			start = append(start, p.T0)
			end = append(end, p.T1)
			fullConsensus = append(fullConsensus, p.T1-p.T0)
			d1 := t1.Sub(t0)
			t2 := time.Unix(0, p.T2)
			t3 := time.Unix(0, p.T3)
			d2 := t3.Sub(t2)
			consensus = append(consensus, d2.Nanoseconds())
			s := fmt.Sprintf("receive block[\"%s\"] @ %s, 1st consensus cost %s, full consensus cost %s\n", p.Hash.String(), time.Unix(0, p.T0).In(loc).Format("2006-01-02 15:04:05.000000"), d2, d1)
			_, _ = fd1.WriteString(s)
		}
	}

	sort.Slice(consensus, func(i, j int) bool {
		return consensus[i] < consensus[j]
	})
	sort.Slice(start, func(i, j int) bool {
		return start[i] < start[j]
	})
	sort.Slice(end, func(i, j int) bool {
		return end[i] < end[j]
	})
	sort.Slice(fullConsensus, func(i, j int) bool {
		return fullConsensus[i] < fullConsensus[j]
	})

	//fmt.Println("performance time size: ", len(consensus))

	i := average(consensus)
	duration, _ := time.ParseDuration(fmt.Sprintf("%dns", i))
	max := int64(float64(i) * 1.05)
	min := int64(float64(i) * 0.95)
	maxDuration, _ := time.ParseDuration(fmt.Sprintf("%dns", max))
	minDuration, _ := time.ParseDuration(fmt.Sprintf("%dns", min))

	consensus2 := filter(consensus, func(i int64) bool {
		return i <= max || i >= min
	})

	avFullConsensus := average(fullConsensus)
	minFullConsensus := fullConsensus[0]
	maxFullConsensus := fullConsensus[len(fullConsensus)-1]
	d1, _ := time.ParseDuration(fmt.Sprintf("%dns", avFullConsensus))
	d2, _ := time.ParseDuration(fmt.Sprintf("%dns", maxFullConsensus))
	d3, _ := time.ParseDuration(fmt.Sprintf("%dns", minFullConsensus))
	fmt.Printf("full consensus cost time => max: %s, min: %s, average: %s\n", d2, d3, d1)

	av := average(consensus2)

	t := fmt.Sprintf("%dns", av)
	duration2, _ := time.ParseDuration(t)

	fmt.Printf("average consensus[%d]: %s, filter average by [%s %s] consensus[%d]: %s, capacity %f\n", len(consensus), duration, minDuration, maxDuration, len(consensus2),
		duration2, float64(1000000000)/float64(duration.Nanoseconds()))

	//fmt.Println("===> start ", start)
	//fmt.Println("===> end ", end)

	start0 := time.Unix(0, start[0])
	start1 := time.Unix(0, start[len(end)-1])

	startDuration := start1.Sub(start0)

	fmt.Printf("transfer %d Tx from %s to %s cost %s\n", len(start), start0.In(loc).Format("2006-01-02 15:04:05.000000"),
		start1.In(loc).Format("2006-01-02 15:04:05.000000"), startDuration)

	end0 := time.Unix(0, end[0])
	end1 := time.Unix(0, end[len(end)-1])

	fmt.Printf("consensus %d Tx from %s to %s cost %s\n", len(start), end0.In(loc).Format("2006-01-02 15:04:05.000000"),
		end1.In(loc).Format("2006-01-02 15:04:05.000000"), end1.Sub(end0))

	d := end1.Sub(start0)
	i2 := float64(0)
	if d.Seconds() > 0 {
		i2 = d.Seconds() / float64(len(start))
	} else {
		i2 = 0
	}

	fmt.Printf("all %d Tx from %s to %s cost %s, one Tx cost %f second(s) \n", len(start), start0.In(loc).Format("2006-01-02 15:04:05.000000"), end1.In(loc).Format("2006-01-02 15:04:05.000000"), d, i2)
	return nil
}

func average(slice []int64) int64 {
	sum := int64(0)

	for _, v := range slice {
		sum += v
	}

	return sum / int64(len(slice))
}

func filter(slice []int64, fn func(int64) bool) []int64 {
	b := slice[:0]
	for _, x := range slice {
		if fn(x) {
			b = append(b, x)
		}
	}

	return b
}
