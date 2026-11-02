// Decoders for the engine values the indexer serves as CBOR.
//
// Since core 0.48.0 the indexer REST API serves substate values, execution results, receipts,
// events, resources and UTXOs as their CBOR encoding, base64 in a JSON string. The SDK's own
// flows hand those strings to the core untouched. These functions are for callers who read the
// values themselves: each takes the served JSON string verbatim (quotes included) and returns the
// value decoded by the core. Bare base64, or a value of another type, is a "PARSE" *Error.

package ootle

import (
	"encoding/json"
	"fmt"

	"github.com/tari-project/ootle-go/internal/cffi"
)

// DecodedEvent is an engine event decoded from its served CBOR encoding.
type DecodedEvent struct {
	// SubstateID is the emitting substate id, or "" when the event has none.
	SubstateID string
	// TemplateAddress is the address of the template that emitted the event.
	TemplateAddress string
	// Topic is the event topic (e.g. "std.vault.withdraw").
	Topic string
	// Payload is the event payload as a flat key→raw-JSON map. Values are arbitrary CBOR rendered
	// to JSON (strings, numbers, or tagged objects such as a typed address or a u128 amount), so
	// do not assume a string. Nil when the event has no payload.
	Payload map[string]json.RawMessage
}

// wireDecodedEvent is the data JSON of ootle_decode_event.
type wireDecodedEvent struct {
	SubstateID      *string                    `json:"substate_id"`
	TemplateAddress string                     `json:"template_address"`
	Topic           string                     `json:"topic"`
	Payload         map[string]json.RawMessage `json:"payload"`
}

// DecodeEvent decodes an event as the indexer serves it (the event field of a
// /transactions/events/stream message, or of a /transactions/events query).
func DecodeEvent(served json.RawMessage) (DecodedEvent, error) {
	dataJSON, cerr := cffi.DecodeEvent(string(served))
	if cerr != nil {
		return DecodedEvent{}, fromCffiError(cerr)
	}
	var wire wireDecodedEvent
	if err := json.Unmarshal([]byte(dataJSON), &wire); err != nil {
		return DecodedEvent{}, &Error{Code: "ENCODING", Message: fmt.Sprintf("unmarshal decoded event: %v", err)}
	}
	out := DecodedEvent{TemplateAddress: wire.TemplateAddress, Topic: wire.Topic, Payload: wire.Payload}
	if wire.SubstateID != nil {
		out.SubstateID = *wire.SubstateID
	}
	return out, nil
}

// DecodeSubstateValue decodes a substate value as the indexer serves it (for example
// transport.FetchedSubstate.SubstateValue) into its full engine JSON, keyed by variant
// ({"Component":{…}}, {"Vault":{…}}, …). For the typed summary the SDK uses, see DecodeSubstate.
func DecodeSubstateValue(served json.RawMessage) (json.RawMessage, error) {
	return decodeServed(served, cffi.DecodeSubstateValue)
}

// DecodeExecuteResult decodes an execution result as the indexer serves it (a dry run's result,
// or a finalized result's execution_result) into its engine JSON. The SDK's own submit flows
// return the typed FinalizedResult instead.
func DecodeExecuteResult(served json.RawMessage) (json.RawMessage, error) {
	return decodeServed(served, cffi.DecodeExecuteResult)
}

// DecodeTransactionReceipt decodes a transaction receipt as the indexer serves it into its engine
// JSON.
func DecodeTransactionReceipt(served json.RawMessage) (json.RawMessage, error) {
	return decodeServed(served, cffi.DecodeTransactionReceipt)
}

// DecodeResource decodes a resource as the indexer serves it into its engine JSON.
func DecodeResource(served json.RawMessage) (json.RawMessage, error) {
	return decodeServed(served, cffi.DecodeResource)
}

// DecodeUTXO decodes a UTXO as the indexer serves it into its engine JSON.
func DecodeUTXO(served json.RawMessage) (json.RawMessage, error) {
	return decodeServed(served, cffi.DecodeUTXO)
}

func decodeServed(served json.RawMessage, decode func(string) (string, error)) (json.RawMessage, error) {
	dataJSON, cerr := decode(string(served))
	if cerr != nil {
		return nil, fromCffiError(cerr)
	}
	return json.RawMessage(dataJSON), nil
}
