package abi

import (
	"testing"
)

func TestDoDSettleBillingType(t *testing.T) {
	n := DoDSettleBillingTypeNames()
	t.Log(n)

	bt := DoDSettleBillingTypePAYG
	if bt.String() != "PAYG" {
		t.Fatal()
	}

	DoDSettleBillingType(1000).String()

	pbt, _ := ParseDoDSettleBillingType("PAYG")
	if pbt != DoDSettleBillingTypePAYG {
		t.Fatal()
	}

	_, err := ParseDoDSettleBillingType("invalid")
	if err == nil {
		t.Fatal()
	}

	data, err := bt.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	err = bt.UnmarshalText(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleBillingUnit(t *testing.T) {
	n := DoDSettleBillingUnitNames()
	t.Log(n)

	bu := DoDSettleBillingUnitSecond
	if bu.String() != "second" {
		t.Fatal()
	}

	DoDSettleBillingUnit(1000).String()

	pbu, _ := ParseDoDSettleBillingUnit("second")
	if pbu != DoDSettleBillingUnitSecond {
		t.Fatal()
	}

	_, err := ParseDoDSettleBillingUnit("invalid")
	if err == nil {
		t.Fatal()
	}

	data, err := bu.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	err = bu.UnmarshalText(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleContractState(t *testing.T) {
	n := DoDSettleContractStateNames()
	t.Log(n)

	sr := DoDSettleContractStateRequest
	if sr.String() != "request" {
		t.Fatal()
	}

	DoDSettleContractState(1000).String()

	psr, _ := ParseDoDSettleContractState("request")
	if psr != DoDSettleContractStateRequest {
		t.Fatal()
	}

	_, err := ParseDoDSettleContractState("invalid")
	if err == nil {
		t.Fatal()
	}

	data, err := sr.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	err = sr.UnmarshalText(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleOrderState(t *testing.T) {
	n := DoDSettleOrderStateNames()
	t.Log(n)

	os := DoDSettleOrderStateSuccess
	if os.String() != "success" {
		t.Fatal()
	}

	DoDSettleOrderState(1000).String()

	pos, _ := ParseDoDSettleOrderState("success")
	if pos != DoDSettleOrderStateSuccess {
		t.Fatal()
	}

	_, err := ParseDoDSettleOrderState("invalid")
	if err == nil {
		t.Fatal()
	}

	data, err := os.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	err = os.UnmarshalText(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleOrderType(t *testing.T) {
	n := DoDSettleOrderTypeNames()
	t.Log(n)

	ot := DoDSettleOrderTypeCreate
	if ot.String() != "create" {
		t.Fatal()
	}

	DoDSettleOrderType(1000).String()

	pot, _ := ParseDoDSettleOrderType("create")
	if pot != DoDSettleOrderTypeCreate {
		t.Fatal()
	}

	_, err := ParseDoDSettleOrderType("invalid")
	if err == nil {
		t.Fatal()
	}

	data, err := ot.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	err = ot.UnmarshalText(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettlePaymentType(t *testing.T) {
	n := DoDSettlePaymentTypeNames()
	t.Log(n)

	pt := DoDSettlePaymentTypeInvoice
	if pt.String() != "invoice" {
		t.Fatal()
	}

	DoDSettlePaymentType(1000).String()

	ppt, _ := ParseDoDSettlePaymentType("invoice")
	if ppt != DoDSettlePaymentTypeInvoice {
		t.Fatal()
	}

	_, err := ParseDoDSettlePaymentType("invalid")
	if err == nil {
		t.Fatal()
	}

	data, err := pt.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	err = pt.UnmarshalText(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleResponseAction(t *testing.T) {
	n := DoDSettleResponseActionNames()
	t.Log(n)

	ac := DoDSettleResponseActionConfirm
	if ac.String() != "confirm" {
		t.Fatal()
	}

	DoDSettleResponseAction(1000).String()

	pac, _ := ParseDoDSettleResponseAction("confirm")
	if pac != DoDSettleResponseActionConfirm {
		t.Fatal()
	}

	_, err := ParseDoDSettleResponseAction("invalid")
	if err == nil {
		t.Fatal()
	}

	data, err := ac.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	err = ac.UnmarshalText(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleServiceClass(t *testing.T) {
	n := DoDSettleServiceClassNames()
	t.Log(n)

	sc := DoDSettleServiceClassGold
	if sc.String() != "gold" {
		t.Fatal()
	}

	DoDSettleServiceClass(1000).String()

	psc, _ := ParseDoDSettleServiceClass("gold")
	if psc != DoDSettleServiceClassGold {
		t.Fatal()
	}

	_, err := ParseDoDSettleServiceClass("invalid")
	if err == nil {
		t.Fatal()
	}

	data, err := sc.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	err = sc.UnmarshalText(data)
	if err != nil {
		t.Fatal(err)
	}
}
