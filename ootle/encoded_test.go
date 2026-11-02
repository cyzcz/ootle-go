package ootle

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// servedVault is a public fungible vault of 1 000 000 resource_0101… as the indexer serves it.
const servedVault = `"ggKBgoIAg9iDWCABAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBARoAD0JAAAA="`

func TestDecodeEvent(t *testing.T) {
	ev, err := DecodeEvent(json.RawMessage(servedDepositEvent))
	if err != nil {
		t.Fatalf("DecodeEvent: %v", err)
	}
	if ev.Topic != "std.deposit" || ev.SubstateID != "component_"+strings.Repeat("41", 32) {
		t.Errorf("DecodeEvent = %+v", ev)
	}
	if string(ev.Payload["amount"]) != `"1000000"` {
		t.Errorf("Payload = %v", ev.Payload)
	}
}

func TestDecodeSubstateValue(t *testing.T) {
	got, err := DecodeSubstateValue(json.RawMessage(servedVault))
	if err != nil {
		t.Fatalf("DecodeSubstateValue: %v", err)
	}
	var v map[string]json.RawMessage
	if err := json.Unmarshal(got, &v); err != nil || v["Vault"] == nil {
		t.Fatalf("DecodeSubstateValue = %s, want a Vault variant", got)
	}
	if !strings.Contains(string(got), "resource_"+strings.Repeat("01", 32)) {
		t.Errorf("DecodeSubstateValue = %s, want it to name resource_0101…", got)
	}
}

// TestDecodersRejectUnservedInput asserts bare base64 (no JSON quotes) and a served value of
// another type are both PARSE errors.
func TestDecodersRejectUnservedInput(t *testing.T) {
	cases := map[string]func() error{
		"bare base64": func() error {
			_, err := DecodeSubstateValue(json.RawMessage(strings.Trim(servedVault, `"`)))
			return err
		},
		"event as a substate value": func() error {
			_, err := DecodeSubstateValue(json.RawMessage(servedDepositEvent))
			return err
		},
		"vault as an event": func() error {
			_, err := DecodeEvent(json.RawMessage(servedVault))
			return err
		},
	}
	for name, run := range cases {
		var oerr *Error
		if err := run(); !errors.As(err, &oerr) || oerr.Code != "PARSE" {
			t.Errorf("%s: err = %v, want a PARSE *Error", name, err)
		}
	}
}
