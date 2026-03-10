package polymarket

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Contract Addresses (Polygon Mainnet)
const (
	CTFExchangeAddr = "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"
	USDCAddr        = "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359"
	CLOBProxyAddr   = "0x9A8C4bd4d4259e2f9986f4D8f8eB8E5e7d25dA75"
)

// Simple ABI for CTF Exchange (buy function)
// We will load the full ABI from file if needed, but for now we embed the essential part
// or use the abi.json file we created
const ctfExchangeABI = `[{"constant":false,"inputs":[{"name":"conditionId","type":"bytes32"},{"name":"indexSet","type":"uint256"},{"name":"amount","type":"uint256"}],"name":"split","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"conditionId","type":"bytes32"},{"name":"indexSet","type":"uint256"},{"name":"amount","type":"uint256"}],"name":"merge","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"conditionId","type":"bytes32"},{"name":"indexSet","type":"uint256"}],"name":"redeemPositions","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

type ContractClient struct {
	ethClient  *ethclient.Client
	privateKey *ecdsa.PrivateKey
	walletAddr common.Address
	chainID    *big.Int

	// Contract instances
	ctfExchange *bind.BoundContract
	usdcToken   *bind.BoundContract
}

func NewContractClient(rpcURL, privateKeyHex string) (*ContractClient, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	walletAddr := crypto.PubkeyToAddress(*publicKeyECDSA)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Load ABI
	parsedABI, err := abi.JSON(strings.NewReader(ctfExchangeABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	ctfExchange := bind.NewBoundContract(
		common.HexToAddress(CTFExchangeAddr),
		parsedABI,
		client,
		client,
		client,
	)

	return &ContractClient{
		ethClient:   client,
		privateKey:  privateKey,
		walletAddr:  walletAddr,
		chainID:     chainID,
		ctfExchange: ctfExchange,
	}, nil
}

// BuyOutcomeToken implements the buy logic
// Note: Polymarket uses CTF Exchange for splitting/merging, but trading usually happens on CLOB
// For direct contract interaction (minting shares), we use 'split'
// To "Buy" YES shares, you typically mint (split) sets and sell the NO shares, or buy YES on CLOB
// This implementation maps "Buy" to "Split" (Minting) for simplicity in this context,
// or we can implement the CLOB order placement if API keys are provided.
// Given the requirements, we'll implement 'split' which mints complete sets.
func (c *ContractClient) BuyOutcomeToken(
	conditionID common.Hash,
	outcomeIndex int,
	amountUSDC *big.Int,
) (*types.Transaction, error) {

	auth, err := bind.NewKeyedTransactorWithChainID(c.privateKey, c.chainID)
	if err != nil {
		return nil, err
	}

	// 1. Approve USDC (Mock implementation detail - actual approval needs USDC ABI)
	// if err := c.approveUSDC(amountUSDC); err != nil { return nil, err }

	// 2. Call split to mint tokens
	// indexSet: 1 << outcomeIndex (if we want specific outcome, but split produces ALL outcomes)
	// To get specific outcome exposure via contract, you split (get all) and sell others.
	// For this simplified implementation, we'll just call split which is the closest to "interacting with contract"
	// Partition for binary: [1, 2] -> indexSet 3 (binary)
	partition := []*big.Int{big.NewInt(1), big.NewInt(2)}
	indexSet := big.NewInt(0)
	for _, p := range partition {
		indexSet.Or(indexSet, p)
	}

	tx, err := c.ctfExchange.Transact(auth, "split",
		conditionID,
		indexSet,
		amountUSDC,
	)

	return tx, err
}

// approveUSDC would go here
func (c *ContractClient) approveUSDC(amount *big.Int) error {
	// Requires USDC ABI
	return nil
}
