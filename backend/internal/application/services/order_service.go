package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain/repositories"
)

// OrderService handles order business logic
type OrderService struct {
	orderRepo     repositories.OrderRepository
	orderItemRepo repositories.OrderItemRepository
	cartRepo      repositories.CartRepository
	cartItemRepo  repositories.CartItemRepository
	productRepo   repositories.ProductRepository
	inventoryRepo repositories.InventoryLogRepository
	paymentRepo   repositories.PaymentTransactionRepository
	receiptRepo   repositories.ReceiptRepository
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	orderItemRepo repositories.OrderItemRepository,
	cartRepo repositories.CartRepository,
	cartItemRepo repositories.CartItemRepository,
	productRepo repositories.ProductRepository,
	inventoryRepo repositories.InventoryLogRepository,
	paymentRepo repositories.PaymentTransactionRepository,
	receiptRepo repositories.ReceiptRepository,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		cartRepo:      cartRepo,
		cartItemRepo:  cartItemRepo,
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		paymentRepo:   paymentRepo,
		receiptRepo:   receiptRepo,
	}
}

// CreateOrderFromCart creates an order from a shopping cart
func (s *OrderService) CreateOrderFromCart(ctx context.Context, cartID uuid.UUID, customerInfo domain.CustomerInfo) (*domain.Order, error) {
	// Get cart and items
	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	if cart.Status != domain.CartStatusActive {
		return nil, errors.New("cart is not active")
	}

	cartItems, err := s.cartItemRepo.GetByCartID(ctx, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	if len(cartItems) == 0 {
		return nil, errors.New("cart is empty")
	}

	// Validate stock availability for all items
	for _, item := range cartItems {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product %s: %w", item.ProductID, err)
		}

		if product.ManageStock && product.StockQuantity < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s: only %d available", product.Name, product.StockQuantity)
		}
	}

	// Generate order number
	orderNumber, err := s.orderRepo.GenerateOrderNumber(ctx, cart.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate order number: %w", err)
	}

	// Create order
	order := &domain.Order{
		TenantID:        cart.TenantID,
		OrderNumber:     orderNumber,
		UserID:          cart.UserID,
		Status:          domain.OrderStatusPending,
		PaymentStatus:   domain.PaymentStatusPending,
		CustomerEmail:   customerInfo.Email,
		CustomerPhone:   customerInfo.Phone,
		BillingAddress:  customerInfo.BillingAddress,
		ShippingAddress: customerInfo.ShippingAddress,
		Currency:        cart.Currency,
		Subtotal:        cart.Subtotal,
		TaxTotal:        cart.TaxTotal,
		ShippingTotal:   cart.ShippingTotal,
		DiscountTotal:   cart.DiscountTotal,
		Total:           cart.Total,
		CustomerNote:    customerInfo.Note,
		DateCreated:     time.Now(),
		DateModified:    time.Now(),
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items and update inventory
	for _, cartItem := range cartItems {
		product, err := s.productRepo.GetByID(ctx, cartItem.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product: %w", err)
		}

		// Create order item
		orderItem := &domain.OrderItem{
			OrderID:     order.ID,
			ProductID:   cartItem.ProductID,
			VariationID: cartItem.VariationID,
			Quantity:    cartItem.Quantity,
			UnitPrice:   cartItem.UnitPrice,
			TotalPrice:  cartItem.TotalPrice,
			ProductName: product.Name,
			ProductSKU:  product.SKU,
			ProductData: cartItem.ProductData,
		}

		if err := s.orderItemRepo.Create(ctx, orderItem); err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		// Update product stock and create inventory log
		if product.ManageStock {
			newStock := product.StockQuantity - cartItem.Quantity
			if err := s.productRepo.UpdateStock(ctx, product.ID, newStock); err != nil {
				return nil, fmt.Errorf("failed to update product stock: %w", err)
			}

			// Create inventory log
			inventoryLog := &domain.InventoryLog{
				TenantID:       cart.TenantID,
				ProductID:      product.ID,
				VariationID:    cartItem.VariationID,
				Type:           domain.InventoryTypeSale,
				Quantity:       cartItem.Quantity,
				QuantityBefore: product.StockQuantity,
				QuantityAfter:  newStock,
				Reason:         fmt.Sprintf("Sale - Order %s", order.OrderNumber),
				ReferenceID:    &order.ID,
				ReferenceType:  "order",
				CostPerUnit:    product.CostPrice,
				TotalCost:      product.CostPrice * float64(cartItem.Quantity),
				UserID:         cart.UserID,
			}

			if err := s.inventoryRepo.Create(ctx, inventoryLog); err != nil {
				// Log error but don't fail order creation
				// TODO: Add proper logging
			}
		}
	}

	// Mark cart as converted
	if err := s.cartRepo.UpdateStatus(ctx, cartID, domain.CartStatusConverted); err != nil {
		// Log error but don't fail order creation
		// TODO: Add proper logging
	}

	return order, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, userID *uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Validate status transition
	if !s.isValidStatusTransition(order.Status, status) {
		return fmt.Errorf("invalid status transition from %s to %s", order.Status, status)
	}

	// Handle status-specific logic
	switch status {
	case domain.OrderStatusCancelled:
		if err := s.handleOrderCancellation(ctx, order, userID); err != nil {
			return fmt.Errorf("failed to handle order cancellation: %w", err)
		}
	case domain.OrderStatusDelivered:
		order.DateCompleted = &[]time.Time{time.Now()}[0]
	}

	order.Status = status
	order.DateModified = time.Now()

	return s.orderRepo.Update(ctx, order)
}

// ProcessPayment processes a payment for an order
func (s *OrderService) ProcessPayment(ctx context.Context, orderID uuid.UUID, paymentData domain.PaymentData) (*domain.PaymentTransaction, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order.PaymentStatus == domain.PaymentStatusPaid {
		return nil, errors.New("order is already paid")
	}

	// Create payment transaction
	transaction := &domain.PaymentTransaction{
		TenantID:        order.TenantID,
		OrderID:         orderID,
		TransactionID:   paymentData.TransactionID,
		PaymentGateway:  paymentData.Gateway,
		PaymentMethod:   paymentData.Method,
		Amount:          order.Total,
		Currency:        order.Currency,
		Status:          domain.TransactionStatusPending,
		Type:            domain.TransactionTypePayment,
		GatewayResponse: paymentData.GatewayResponse,
	}

	if err := s.paymentRepo.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create payment transaction: %w", err)
	}

	// TODO: Integrate with actual payment gateways
	// For now, assume payment is successful
	transaction.Status = domain.TransactionStatusCompleted
	transaction.ProcessedAt = &[]time.Time{time.Now()}[0]

	if err := s.paymentRepo.Update(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to update payment transaction: %w", err)
	}

	// Update order payment status
	order.PaymentStatus = domain.PaymentStatusPaid
	order.PaymentMethod = paymentData.Method
	order.PaymentMethodTitle = paymentData.MethodTitle
	order.TransactionID = paymentData.TransactionID
	order.DatePaid = &[]time.Time{time.Now()}[0]
	order.DateModified = time.Now()

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order payment status: %w", err)
	}

	// Generate receipt
	if err := s.generateReceipt(ctx, order); err != nil {
		// Log error but don't fail payment processing
		// TODO: Add proper logging
	}

	return transaction, nil
}

// RefundOrder processes a refund for an order
func (s *OrderService) RefundOrder(ctx context.Context, orderID uuid.UUID, amount float64, reason string, userID *uuid.UUID) (*domain.PaymentTransaction, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order.PaymentStatus != domain.PaymentStatusPaid {
		return nil, errors.New("order is not paid")
	}

	// Get refundable transactions
	refundableTransactions, err := s.paymentRepo.GetRefundableTransactions(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get refundable transactions: %w", err)
	}

	if len(refundableTransactions) == 0 {
		return nil, errors.New("no refundable transactions found")
	}

	// Calculate total refundable amount
	var totalRefundable float64
	for _, tx := range refundableTransactions {
		if tx.Type == domain.TransactionTypePayment {
			totalRefundable += tx.Amount
		}
	}

	if amount > totalRefundable {
		return nil, fmt.Errorf("refund amount exceeds refundable amount: %.2f", totalRefundable)
	}

	// Create refund transaction
	refundTransaction := &domain.PaymentTransaction{
		TenantID:            order.TenantID,
		OrderID:             orderID,
		TransactionID:       generateRefundTransactionID(),
		PaymentGateway:      refundableTransactions[0].PaymentGateway,
		PaymentMethod:       refundableTransactions[0].PaymentMethod,
		Amount:              amount,
		Currency:            order.Currency,
		Status:              domain.TransactionStatusCompleted,
		Type:                domain.TransactionTypeRefund,
		ParentTransactionID: &refundableTransactions[0].ID,
		ProcessedAt:         &[]time.Time{time.Now()}[0],
	}

	if err := s.paymentRepo.Create(ctx, refundTransaction); err != nil {
		return nil, fmt.Errorf("failed to create refund transaction: %w", err)
	}

	// Update order payment status
	if amount >= order.Total {
		order.PaymentStatus = domain.PaymentStatusRefunded
		order.Status = domain.OrderStatusRefunded
	} else {
		order.PaymentStatus = domain.PaymentStatusPartiallyRefunded
	}
	order.DateModified = time.Now()

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// Restore inventory if full refund
	if amount >= order.Total {
		if err := s.restoreInventoryFromOrder(ctx, order, userID); err != nil {
			// Log error but don't fail refund
			// TODO: Add proper logging
		}
	}

	return refundTransaction, nil
}

// Helper methods

func (s *OrderService) handleOrderCancellation(ctx context.Context, order *domain.Order, userID *uuid.UUID) error {
	// Restore inventory
	return s.restoreInventoryFromOrder(ctx, order, userID)
}

func (s *OrderService) restoreInventoryFromOrder(ctx context.Context, order *domain.Order, userID *uuid.UUID) error {
	orderItems, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("failed to get order items: %w", err)
	}

	for _, item := range orderItems {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			continue // Skip if product not found
		}

		if product.ManageStock {
			newStock := product.StockQuantity + item.Quantity
			if err := s.productRepo.UpdateStock(ctx, product.ID, newStock); err != nil {
				continue // Skip on error
			}

			// Create inventory log
			inventoryLog := &domain.InventoryLog{
				TenantID:       order.TenantID,
				ProductID:      product.ID,
				VariationID:    item.VariationID,
				Type:           domain.InventoryTypeReturn,
				Quantity:       item.Quantity,
				QuantityBefore: product.StockQuantity,
				QuantityAfter:  newStock,
				Reason:         fmt.Sprintf("Order cancellation/refund - Order %s", order.OrderNumber),
				ReferenceID:    &order.ID,
				ReferenceType:  "order_cancellation",
				CostPerUnit:    product.CostPrice,
				TotalCost:      product.CostPrice * float64(item.Quantity),
				UserID:         userID,
			}

			s.inventoryRepo.Create(ctx, inventoryLog)
		}
	}

	return nil
}

func (s *OrderService) generateReceipt(ctx context.Context, order *domain.Order) error {
	receiptNumber, err := s.receiptRepo.GenerateReceiptNumber(ctx, order.TenantID)
	if err != nil {
		return fmt.Errorf("failed to generate receipt number: %w", err)
	}

	// Get order items
	orderItems, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("failed to get order items: %w", err)
	}

	// Prepare receipt data
	receiptData := map[string]interface{}{
		"order":        order,
		"items":        orderItems,
		"generated_at": time.Now(),
	}

	receipt := &domain.Receipt{
		TenantID:       order.TenantID,
		OrderID:        order.ID,
		ReceiptNumber:  receiptNumber,
		ReceiptData:    receiptData,
		EmailRecipient: order.CustomerEmail,
	}

	return s.receiptRepo.Create(ctx, receipt)
}

func (s *OrderService) isValidStatusTransition(currentStatus, newStatus string) bool {
	validTransitions := map[string][]string{
		domain.OrderStatusPending:    {domain.OrderStatusProcessing, domain.OrderStatusCancelled},
		domain.OrderStatusProcessing: {domain.OrderStatusShipped, domain.OrderStatusCancelled},
		domain.OrderStatusShipped:    {domain.OrderStatusDelivered},
		domain.OrderStatusDelivered:  {domain.OrderStatusRefunded},
		domain.OrderStatusCancelled:  {}, // Terminal state
		domain.OrderStatusRefunded:   {}, // Terminal state
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}

	return false
}

func generateRefundTransactionID() string {
	return fmt.Sprintf("refund_%d_%s", time.Now().Unix(), uuid.New().String()[:8])
}
