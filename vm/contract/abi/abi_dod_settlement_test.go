package abi

import (
	"strings"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/vm/abi"
)

func TestDoDSettlementABI(t *testing.T) {
	_, err := abi.JSONToABIContract(strings.NewReader(JsonDoDSettlement))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleBillingUnitRound(t *testing.T) {
	start := time.Now().Unix() - 10

	end := DoDSettleBillingUnitRound(DoDSettleBillingUnitYear, start)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitYear, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitMonth, start)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitMonth, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitWeek, start)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitWeek, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitDay, start)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitDay, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitHour, start)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitHour, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitMinute, start)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitMinute, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitSecond, start)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitSecond, start, end) != int(end-start) {
		t.Fatal()
	}
}
