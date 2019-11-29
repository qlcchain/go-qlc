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
	rpc "github.com/qlcchain/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
)

func addTxPerformanceByShell(parentCmd *ishell.Cmd) {
	cfgPath := util.Flag{
		Name:  "config",
		Must:  false,
		Usage: "config file path",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "performance",
		Help: "get performance time",
		Func: func(c *ishell.Context) {
			args := []util.Flag{cfgPath}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}
			cfgPathP := util.StringVar(c.Args, cfgPath)
			err := getPerformanceTime(cfgPathP)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(c)
}

func addTxPerformanceByCobra(parentCmd *cobra.Command) {
	var cfgPathP string
	var performanceTimeCmd = &cobra.Command{
		Use:   "performance",
		Short: "get performance time",
		Run: func(cmd *cobra.Command, args []string) {
			err := getPerformanceTime(cfgPathP)
			if err != nil {
				cmd.Println(err)
				return
			}
		},
	}
	performanceTimeCmd.PersistentFlags().StringVarP(&cfgPathP, "config", "c", "", "config file path")
	parentCmd.AddCommand(performanceTimeCmd)
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
		if p.T0 == 0 || p.T1 == 0 || p.Hash.IsZero() {
			//_, _ = fd1.WriteString(fmt.Sprintf("invalid data %s \n", p.String()))
		} else {
			t0 := time.Unix(0, p.T0)
			t1 := time.Unix(0, p.T1)
			start = append(start, p.T0)
			end = append(end, p.T1)
			fullConsensus = append(fullConsensus, p.T1-p.T0)
			d1 := t1.Sub(t0)
			s := fmt.Sprintf("receive block[\"%s\"] @ %s, consensus cost %s\n", p.Hash.String(), time.Unix(0, p.T0).In(loc).Format("2006-01-02 15:04:05.000000"), d1)
			_, _ = fd1.WriteString(s)
		}
	}

	if len(fullConsensus) == 0 {
		fmt.Println("no transaction has completed consensus")
		return nil
	}

	sort.Slice(start, func(i, j int) bool {
		return start[i] < start[j]
	})

	sort.Slice(end, func(i, j int) bool {
		return end[i] < end[j]
	})

	sort.Slice(fullConsensus, func(i, j int) bool {
		return fullConsensus[i] < fullConsensus[j]
	})

	avFullConsensus := average(fullConsensus)
	minFullConsensus := fullConsensus[0]
	maxFullConsensus := fullConsensus[len(fullConsensus)-1]
	d1, _ := time.ParseDuration(fmt.Sprintf("%dns", avFullConsensus))
	d2, _ := time.ParseDuration(fmt.Sprintf("%dns", maxFullConsensus))
	d3, _ := time.ParseDuration(fmt.Sprintf("%dns", minFullConsensus))
	fmt.Printf("full consensus cost time => max: %s, min: %s, average: %s\n", d2, d3, d1)

	start0 := time.Unix(0, start[0])
	start1 := time.Unix(0, start[len(end)-1])
	startDuration := start1.Sub(start0)
	fmt.Printf("transfer %d Tx from %s to %s cost %s\n", len(start), start0.In(loc).Format("2006-01-02 15:04:05.000000"),
		start1.In(loc).Format("2006-01-02 15:04:05.000000"), startDuration)

	end0 := time.Unix(0, end[0])
	end1 := time.Unix(0, end[len(end)-1])
	endDuration := end1.Sub(end0)
	fmt.Printf("consensus %d Tx from %s to %s cost %s\n", len(end), end0.In(loc).Format("2006-01-02 15:04:05.000000"),
		end1.In(loc).Format("2006-01-02 15:04:05.000000"), endDuration)

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
