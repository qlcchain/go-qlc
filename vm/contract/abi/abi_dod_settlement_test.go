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
	to := start + 1

	end := DoDSettleBillingUnitRound(DoDSettleBillingUnitYear, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitYear, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitMonth, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitMonth, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitWeek, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitWeek, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitDay, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitDay, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitHour, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitHour, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitMinute, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitMinute, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitSecond, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitSecond, start, end) != 1 {
		t.Fatal()
	}
}
