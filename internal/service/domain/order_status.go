package domain

type OrderStatus string

const (
	OrderStatusReceived         OrderStatus = "Received"
	OrderStatusInDiagnosis      OrderStatus = "In diagnosis"
	OrderStatusAwaitingApproval OrderStatus = "Awaiting approval"
	OrderStatusInExecution      OrderStatus = "In execution"
	OrderStatusCompleted        OrderStatus = "Completed"
	OrderStatusDelivered        OrderStatus = "Delivered"
)
