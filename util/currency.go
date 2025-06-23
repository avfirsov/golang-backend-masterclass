package util

const (
	USD = "USD"
	EUR = "EUR"
	RUB = "RUB"
	CAD = "CAD"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, RUB, CAD:
		return true
	default:
		return false
	}
}
