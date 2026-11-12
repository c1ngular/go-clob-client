package gamma

import (
	"fmt"
	"strings"
)

// OutcomeAsset pairs an outcome's wire index and label with its quoted price
// and identifiers. Price's empty value and nil IDs mean the service did not
// supply that value; no token ID is substituted for a native position ID.
type OutcomeAsset struct {
	Index      int
	Label      string
	Price      Decimal
	TokenID    *string
	PositionID *string
}

// OutcomeDetails returns all outcomes in wire order, including legacy markets
// with more than two outcomes. Short or absent price/identifier arrays leave the
// corresponding value absent, matching discovery of markets not yet tradable.
// The returned identifiers do not alias the market's arrays.
func (m Market) OutcomeDetails() []OutcomeAsset {
	if m.Outcomes == nil {
		return nil
	}
	assets := make([]OutcomeAsset, len(m.Outcomes))
	for i, label := range m.Outcomes {
		assets[i] = OutcomeAsset{Index: i, Label: label}
		if i < len(m.OutcomePrices) {
			assets[i].Price = Decimal(m.OutcomePrices[i])
		}
		if i < len(m.CLOBTokenIDs) {
			assets[i].TokenID = new(m.CLOBTokenIDs[i])
		}
		if i < len(m.PositionIDs) {
			assets[i].PositionID = new(m.PositionIDs[i])
		}
	}
	return assets
}

// TradingID returns the venue identifier for one outcome, selected by the
// market's protocol version: the position ID for a v2 market or the CLOB token
// ID for a v1 market. The V2 migration requires this choice even when both
// arrays are present; an absent or unknown version, an out-of-range outcome,
// and a missing or non-decimal identifier are rejected rather than guessed.
func (m Market) TradingID(index int) (string, error) {
	if index < 0 || index >= len(m.Outcomes) {
		return "", fmt.Errorf(
			"gamma market %s: outcome index %d out of range",
			m.ID,
			index,
		)
	}
	if m.Version == nil {
		return "", fmt.Errorf("gamma market %s: missing protocol version", m.ID)
	}

	var ids []string
	switch *m.Version {
	case ProtocolVersionV1:
		ids = m.CLOBTokenIDs
	case ProtocolVersionV2:
		ids = m.PositionIDs
	case "":
		return "", fmt.Errorf("gamma market %s: missing protocol version", m.ID)
	default:
		return "", fmt.Errorf(
			"gamma market %s: unsupported protocol version %q",
			m.ID,
			*m.Version,
		)
	}

	if index >= len(ids) {
		return "", fmt.Errorf(
			"gamma market %s: outcome %d has no %s identifier",
			m.ID,
			index,
			*m.Version,
		)
	}
	id := strings.TrimSpace(ids[index])
	if id == "" {
		return "", fmt.Errorf(
			"gamma market %s: outcome %d has an empty %s identifier",
			m.ID,
			index,
			*m.Version,
		)
	}
	for _, digit := range id {
		if digit < '0' || digit > '9' {
			return "", fmt.Errorf(
				"gamma market %s: outcome %d identifier %q is not a decimal integer",
				m.ID,
				index,
				id,
			)
		}
	}
	return id, nil
}
