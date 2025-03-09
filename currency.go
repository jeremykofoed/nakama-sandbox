package main

type CurrencyType string
const (
	Gold CurrencyType = "gold"
	Gems CurrencyType = "gems"
	Experience CurrencyType = "experience"
)

type Currency struct {
	Type CurrencyType `json:"type"`
	Amount int64 `json:"amount"`
}

type RewardType interface {}

// Information on rewards.
type RewardInfo struct {
	Type RewardType `json:"type"`
	Amount int64 `json:"amount"`
}