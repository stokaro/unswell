package annotation

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// Packet contains only the frozen editorial instructions and annotation prompts.
// Intrinsic clues inside prose cannot be removed without changing the task.
type Packet struct {
	Version string   `json:"version"`
	SHA256  string   `json:"sha256,omitempty"`
	Rubric  string   `json:"rubric"`
	Purpose string   `json:"purpose"`
	Profile Profile  `json:"profile"`
	Units   []Prompt `json:"units"`
}

// Prompt omits acquisition metadata, origin claims, judgments, and adjudication.
type Prompt struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Role    string `json:"role"`
	Text    string `json:"text"`
	Context string `json:"context"`
}

// Packet prepares a detached view for independent review and checks annotation permission.
func (r *Round) Packet(ctx context.Context) (Packet, error) {
	if r == nil || r.data.packetSHA256 == "" {
		return Packet{}, fmt.Errorf("load a validated annotation round before preparing a packet")
	}
	for _, unit := range r.data.Units {
		if err := ctx.Err(); err != nil {
			return Packet{}, err
		}
		if !slices.Contains(unit.Rights.AllowedUses, "annotation") {
			return Packet{}, fmt.Errorf("unit %s lacks recorded annotation permission", unit.ID)
		}
	}
	return r.data.packetView(ctx)
}

func (r roundData) packetView(ctx context.Context) (Packet, error) {
	packet := Packet{Version: "unswell-annotation-packet-v1", Rubric: r.Rubric,
		Purpose: r.Purpose, Profile: r.Profile, Units: make([]Prompt, 0, len(r.Units))}
	for _, unit := range r.Units {
		if err := ctx.Err(); err != nil {
			return Packet{}, err
		}
		packet.Units = append(packet.Units, Prompt{ID: unit.ID, Kind: unit.Kind, Role: unit.Role, Text: unit.Text, Context: unit.Context})
	}
	slices.SortFunc(packet.Units, func(a, b Prompt) int { return strings.Compare(a.ID, b.ID) })
	data, err := json.Marshal(packet)
	if err != nil {
		return Packet{}, err
	}
	packet.SHA256 = fmt.Sprintf("%x", sha256.Sum256(data))
	return packet, nil
}
