package cst

const (
	AppName           = "telegram-energy"
	BaseName          = "telegram"
	DateTimeFormatter = "2006-01-02 15:04:05"
	TimeFormatter     = "02 15:04:05"
	PerCountEnergy    = 32000
	LicenseApi        = "http://127.0.0.1/api/license"
)

const (
	OrderStatus = iota
	OrderStatusSuccess
	OrderStatusRunning
	OrderStatusReceived
	OrderStatusApiSuccess
	OrderStatusApiFailure
	OrderStatusFailure
	OrderStatusNotSufficientFunds
	OrderStatusCancel
)
