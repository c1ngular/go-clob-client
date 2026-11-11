package clob

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nijaru/go-clob-client/internal/polyrelay"
)

// Erc20TradingApproval is one required ERC-20 allowance that is not configured
// for a wallet. It mirrors the official SDK model name.
type Erc20TradingApproval = ERC20ApprovalRequest

// Erc1155TradingApproval is one required ERC-1155 operator approval that is not
// configured for a wallet. It mirrors the official SDK model name; the Go
// SDK's existing ERC1155ApprovalForAllRequest is the same shape.
type Erc1155TradingApproval = ERC1155ApprovalForAllRequest

// TradingApprovalScope selects the protocol family whose approvals are read or
// granted. Passing no scope covers every official target, preserving the
// behavior of callers that predate Protocol V2.
type TradingApprovalScope string

const (
	// TradingApprovalScopeAll covers every approval the SDK knows.
	TradingApprovalScopeAll TradingApprovalScope = "all"
	// TradingApprovalScopeV2 covers Protocol V2 trading and settlement only:
	// pUSD allowances and PositionManager operator approvals for ExchangeV3 and
	// the V2 Router. It deliberately excludes the CTF exchanges, collateral
	// adapters, auto-redeem operator, and perps deposit contract.
	TradingApprovalScopeV2 TradingApprovalScope = "v2"
)

// v2OnlyTradingApprovals resolves optional scopes. No scope means the full
// catalog; an explicit All dominates; an unknown scope is an error rather than
// a silent widening or narrowing of the grant.
func v2OnlyTradingApprovals(scopes []TradingApprovalScope) (bool, error) {
	if len(scopes) == 0 {
		return false, nil
	}
	all, v2 := false, false
	for _, scope := range scopes {
		switch scope {
		case "", TradingApprovalScopeAll:
			all = true
		case TradingApprovalScopeV2:
			v2 = true
		default:
			return false, fmt.Errorf("trading approvals: unknown scope %q", scope)
		}
	}
	if all {
		return false, nil
	}
	return v2, nil
}

// MissingTradingApprovals is the set of trading approvals a wallet still needs
// to grant before it can trade through the official contract set.
type MissingTradingApprovals struct {
	ERC20Approvals   []Erc20TradingApproval
	ERC1155Approvals []ERC1155ApprovalForAllRequest
}

// Empty reports whether the wallet has all required approvals.
func (m MissingTradingApprovals) Empty() bool {
	return len(m.ERC20Approvals) == 0 && len(m.ERC1155Approvals) == 0
}

// TradingApprovalsState is the current trading approval state for a wallet.
// It is a read-only snapshot: callers read it and, if missing approvals exist,
// submit them through SetupTradingApprovals or SetupTradingApprovalsGasless.
type TradingApprovalsState struct {
	Missing         *MissingTradingApprovals
	IsFullyApproved bool
}

// buildMissingTradingApprovalCalls converts a missing-approvals set into the
// ABI calls required to grant it. It is shared by the EOA and gasless setup
// paths so the two flows cannot drift.
func buildMissingTradingApprovalCalls(
	missing *MissingTradingApprovals,
) ([]polyrelay.TransactionCall, error) {
	calls := make(
		[]polyrelay.TransactionCall,
		0,
		len(missing.ERC20Approvals)+len(missing.ERC1155Approvals),
	)
	for _, approval := range missing.ERC20Approvals {
		data, err := packERC20Approval(approval)
		if err != nil {
			return nil, err
		}
		calls = append(calls, tokenCall(approval.TokenAddress, data))
	}
	for _, approval := range missing.ERC1155Approvals {
		data, err := packERC1155ApprovalForAll(approval)
		if err != nil {
			return nil, err
		}
		calls = append(calls, tokenCall(approval.TokenAddress, data))
	}
	return calls, nil
}

// GetTradingApprovalsState reads the configured wallet's current on-chain
// approval state and returns a snapshot of what is missing. It does not
// submit any transactions. The wallet defaults to the configured funder
// address; pass an explicit wallet to inspect a different owner. Pass
// TradingApprovalScopeV2 to read solely the native Protocol V2 pairs.
func (c *SignerClient) GetTradingApprovalsState(
	ctx context.Context,
	wallet string,
	scopes ...TradingApprovalScope,
) (*TradingApprovalsState, error) {
	config, err := getContractConfig(c.chainID)
	if err != nil {
		return nil, err
	}
	erc20, erc1155, err := requiredTradingApprovals(c.chainID, config, scopes...)
	if err != nil {
		return nil, err
	}
	ec, err := c.dialRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("token: dial rpc for trading approvals state: %w", err)
	}
	defer ec.Close()

	owner := c.WalletAddress()
	if wallet != "" {
		if !common.IsHexAddress(wallet) {
			return nil, fmt.Errorf("token: invalid wallet address %q", wallet)
		}
		owner = common.HexToAddress(wallet)
	}

	missing := &MissingTradingApprovals{
		ERC20Approvals:   make([]Erc20TradingApproval, 0, len(erc20)),
		ERC1155Approvals: make([]ERC1155ApprovalForAllRequest, 0, len(erc1155)),
	}
	for _, approval := range erc20 {
		allowance, err := readERC20Allowance(
			ctx,
			ec,
			approval.TokenAddress,
			owner,
			approval.SpenderAddress,
		)
		if err != nil {
			return nil, fmt.Errorf("token: read ERC20 approval for %s: %w",
				approval.SpenderAddress.Hex(), err)
		}
		if allowance.Cmp(approval.Amount) < 0 {
			missing.ERC20Approvals = append(missing.ERC20Approvals, approval)
		}
	}
	for _, approval := range erc1155 {
		approved, err := readERC1155ApprovalForAll(
			ctx,
			ec,
			approval.TokenAddress,
			owner,
			approval.OperatorAddress,
		)
		if err != nil {
			return nil, fmt.Errorf("token: read ERC1155 approval for %s: %w",
				approval.OperatorAddress.Hex(), err)
		}
		if !approved {
			missing.ERC1155Approvals = append(missing.ERC1155Approvals, approval)
		}
	}

	return &TradingApprovalsState{
		Missing:         missing,
		IsFullyApproved: missing.Empty(),
	}, nil
}
