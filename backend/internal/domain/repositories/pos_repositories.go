package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// ProductCategoryRepository interface for product category operations
type ProductCategoryRepository interface {
	Create(ctx context.Context, category *domain.ProductCategory) error
	Update(ctx context.Context, category *domain.ProductCategory) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ProductCategory, error)
	GetByTenantID(ctx context.Context, tenantID string, filter *domain.ProductCategoryFilter) ([]*domain.ProductCategory, int64, error)
	GetBySlug(ctx context.Context, tenantID string, slug string) (*domain.ProductCategory, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.ProductCategory, error)
	GetTreeByTenantID(ctx context.Context, tenantID string) ([]*domain.ProductCategory, error)
	Exists(ctx context.Context, tenantID string, name string, excludeID *uuid.UUID) (bool, error)
	UpdateSortOrder(ctx context.Context, categoryID uuid.UUID, sortOrder int) error
}

// ProductRepository interface for product operations
type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetByTenantID(ctx context.Context, tenantID string, filter *domain.ProductFilter) ([]*domain.Product, int64, error)
	GetBySKU(ctx context.Context, tenantID string, sku string) (*domain.Product, error)
	GetBySlug(ctx context.Context, tenantID string, slug string) (*domain.Product, error)
	GetByCategoryID(ctx context.Context, categoryID uuid.UUID, filter *domain.ProductFilter) ([]*domain.Product, int64, error)
	GetFeatured(ctx context.Context, tenantID string, limit int) ([]*domain.Product, error)
	GetLowStock(ctx context.Context, tenantID string) ([]*domain.Product, error)
	GetOutOfStock(ctx context.Context, tenantID string) ([]*domain.Product, error)
	UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error
	BulkUpdateStock(ctx context.Context, updates []domain.StockUpdate) error
	Search(ctx context.Context, tenantID string, query string, filter *domain.ProductFilter) ([]*domain.Product, int64, error)
	Exists(ctx context.Context, tenantID string, sku string, excludeID *uuid.UUID) (bool, error)
}

// ProductVariationRepository interface for product variation operations
type ProductVariationRepository interface {
	Create(ctx context.Context, variation *domain.ProductVariation) error
	Update(ctx context.Context, variation *domain.ProductVariation) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ProductVariation, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*domain.ProductVariation, error)
	GetBySKU(ctx context.Context, tenantID string, sku string) (*domain.ProductVariation, error)
	UpdateStock(ctx context.Context, variationID uuid.UUID, quantity int) error
	BulkCreate(ctx context.Context, variations []*domain.ProductVariation) error
	BulkUpdate(ctx context.Context, variations []*domain.ProductVariation) error
	BulkDelete(ctx context.Context, variationIDs []uuid.UUID) error
}

// InventoryLogRepository interface for inventory log operations
type InventoryLogRepository interface {
	Create(ctx context.Context, log *domain.InventoryLog) error
	GetByProductID(ctx context.Context, productID uuid.UUID, filter *domain.InventoryLogFilter) ([]*domain.InventoryLog, int64, error)
	GetByTenantID(ctx context.Context, tenantID string, filter *domain.InventoryLogFilter) ([]*domain.InventoryLog, int64, error)
	GetByType(ctx context.Context, tenantID string, logType string, filter *domain.InventoryLogFilter) ([]*domain.InventoryLog, int64, error)
	GetByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time) ([]*domain.InventoryLog, error)
	GetSummaryByProduct(ctx context.Context, productID uuid.UUID, startDate, endDate time.Time) (*domain.InventorySummary, error)
	GetMovementsByReference(ctx context.Context, referenceID uuid.UUID, referenceType string) ([]*domain.InventoryLog, error)
}

// CartRepository interface for cart operations
type CartRepository interface {
	Create(ctx context.Context, cart *domain.Cart) error
	Update(ctx context.Context, cart *domain.Cart) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Cart, error)
	GetByUserID(ctx context.Context, tenantID string, userID uuid.UUID) (*domain.Cart, error)
	GetBySessionID(ctx context.Context, tenantID string, sessionID string) (*domain.Cart, error)
	GetActiveByTenantID(ctx context.Context, tenantID string, filter *domain.CartFilter) ([]*domain.Cart, int64, error)
	GetAbandonedCarts(ctx context.Context, tenantID string, before time.Time) ([]*domain.Cart, error)
	UpdateStatus(ctx context.Context, cartID uuid.UUID, status string) error
	UpdateTotals(ctx context.Context, cartID uuid.UUID, totals domain.CartTotals) error
	CleanupExpiredCarts(ctx context.Context, tenantID string, before time.Time) (int64, error)
}

// CartItemRepository interface for cart item operations
type CartItemRepository interface {
	Create(ctx context.Context, item *domain.CartItem) error
	Update(ctx context.Context, item *domain.CartItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CartItem, error)
	GetByCartID(ctx context.Context, cartID uuid.UUID) ([]*domain.CartItem, error)
	GetByCartAndProduct(ctx context.Context, cartID, productID uuid.UUID, variationID *uuid.UUID) (*domain.CartItem, error)
	UpdateQuantity(ctx context.Context, itemID uuid.UUID, quantity int) error
	DeleteByCartID(ctx context.Context, cartID uuid.UUID) error
	BulkCreate(ctx context.Context, items []*domain.CartItem) error
	BulkUpdate(ctx context.Context, items []*domain.CartItem) error
}

// OrderRepository interface for order operations
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	Update(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetByOrderNumber(ctx context.Context, tenantID string, orderNumber string) (*domain.Order, error)
	GetByTenantID(ctx context.Context, tenantID string, filter *domain.OrderFilter) ([]*domain.Order, int64, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, filter *domain.OrderFilter) ([]*domain.Order, int64, error)
	GetByStatus(ctx context.Context, tenantID string, status string, filter *domain.OrderFilter) ([]*domain.Order, int64, error)
	GetByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time) ([]*domain.Order, error)
	UpdateStatus(ctx context.Context, orderID uuid.UUID, status string) error
	UpdatePaymentStatus(ctx context.Context, orderID uuid.UUID, paymentStatus string) error
	GetRecentOrders(ctx context.Context, tenantID string, limit int) ([]*domain.Order, error)
	GetOrderStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (*domain.OrderStats, error)
	GenerateOrderNumber(ctx context.Context, tenantID string) (string, error)
}

// OrderItemRepository interface for order item operations
type OrderItemRepository interface {
	Create(ctx context.Context, item *domain.OrderItem) error
	Update(ctx context.Context, item *domain.OrderItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderItem, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderItem, error)
	GetByProductID(ctx context.Context, productID uuid.UUID, filter *domain.OrderItemFilter) ([]*domain.OrderItem, int64, error)
	BulkCreate(ctx context.Context, items []*domain.OrderItem) error
	BulkUpdate(ctx context.Context, items []*domain.OrderItem) error
	DeleteByOrderID(ctx context.Context, orderID uuid.UUID) error
	GetTopSellingProducts(ctx context.Context, tenantID string, startDate, endDate time.Time, limit int) ([]*domain.ProductSales, error)
}

// PaymentTransactionRepository interface for payment transaction operations
type PaymentTransactionRepository interface {
	Create(ctx context.Context, transaction *domain.PaymentTransaction) error
	Update(ctx context.Context, transaction *domain.PaymentTransaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PaymentTransaction, error)
	GetByTransactionID(ctx context.Context, transactionID string) (*domain.PaymentTransaction, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.PaymentTransaction, error)
	GetByTenantID(ctx context.Context, tenantID string, filter *domain.PaymentTransactionFilter) ([]*domain.PaymentTransaction, int64, error)
	GetByStatus(ctx context.Context, tenantID string, status string) ([]*domain.PaymentTransaction, error)
	GetByGateway(ctx context.Context, tenantID string, gateway string, filter *domain.PaymentTransactionFilter) ([]*domain.PaymentTransaction, int64, error)
	GetByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time) ([]*domain.PaymentTransaction, error)
	UpdateStatus(ctx context.Context, transactionID uuid.UUID, status string) error
	GetTotalByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time) (float64, error)
	GetRefundableTransactions(ctx context.Context, orderID uuid.UUID) ([]*domain.PaymentTransaction, error)
}

// ReceiptRepository interface for receipt operations
type ReceiptRepository interface {
	Create(ctx context.Context, receipt *domain.Receipt) error
	Update(ctx context.Context, receipt *domain.Receipt) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Receipt, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.Receipt, error)
	GetByReceiptNumber(ctx context.Context, tenantID string, receiptNumber string) (*domain.Receipt, error)
	GetByTenantID(ctx context.Context, tenantID string, filter *domain.ReceiptFilter) ([]*domain.Receipt, int64, error)
	UpdateEmailStatus(ctx context.Context, receiptID uuid.UUID, sent bool, sentAt *time.Time, recipient string) error
	UpdatePrintStatus(ctx context.Context, receiptID uuid.UUID, printed bool, printCount int, lastPrintedAt *time.Time) error
	GetUnsentReceipts(ctx context.Context, tenantID string) ([]*domain.Receipt, error)
	GenerateReceiptNumber(ctx context.Context, tenantID string) (string, error)
}

// SalesReportRepository interface for sales report operations
type SalesReportRepository interface {
	Create(ctx context.Context, report *domain.SalesReport) error
	Update(ctx context.Context, report *domain.SalesReport) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SalesReport, error)
	GetByTenantID(ctx context.Context, tenantID string, filter *domain.SalesReportFilter) ([]*domain.SalesReport, int64, error)
	GetByTypeAndPeriod(ctx context.Context, tenantID string, reportType string, startDate, endDate time.Time) (*domain.SalesReport, error)
	GetLatestByType(ctx context.Context, tenantID string, reportType string) (*domain.SalesReport, error)
	DeleteExpiredReports(ctx context.Context, tenantID string, before time.Time) (int64, error)
	GetAvailableReports(ctx context.Context, tenantID string) ([]*domain.SalesReportSummary, error)
}

// DiscountRepository interface for discount operations
type DiscountRepository interface {
	Create(ctx context.Context, discount *domain.Discount) error
	Update(ctx context.Context, discount *domain.Discount) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Discount, error)
	GetByCode(ctx context.Context, tenantID string, code string) (*domain.Discount, error)
	GetByTenantID(ctx context.Context, tenantID string, filter *domain.DiscountFilter) ([]*domain.Discount, int64, error)
	GetActiveDiscounts(ctx context.Context, tenantID string) ([]*domain.Discount, error)
	GetApplicableDiscounts(ctx context.Context, tenantID string, productIDs []uuid.UUID, categoryIDs []uuid.UUID, amount float64) ([]*domain.Discount, error)
	UpdateUsageCount(ctx context.Context, discountID uuid.UUID, increment int) error
	ValidateDiscount(ctx context.Context, tenantID string, code string, productIDs []uuid.UUID, amount float64, userID *uuid.UUID) (*domain.DiscountValidation, error)
	GetUsageByCustomer(ctx context.Context, discountID, userID uuid.UUID) (int, error)
	Exists(ctx context.Context, tenantID string, code string, excludeID *uuid.UUID) (bool, error)
}

// WishlistRepository interface for wishlist operations
type WishlistRepository interface {
	Create(ctx context.Context, wishlist *domain.Wishlist) error
	Update(ctx context.Context, wishlist *domain.Wishlist) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Wishlist, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Wishlist, error)
	GetByShareToken(ctx context.Context, token string) (*domain.Wishlist, error)
	GetPublicWishlists(ctx context.Context, tenantID string, filter *domain.WishlistFilter) ([]*domain.Wishlist, int64, error)
	GenerateShareToken(ctx context.Context, wishlistID uuid.UUID) (string, error)
}

// WishlistItemRepository interface for wishlist item operations
type WishlistItemRepository interface {
	Create(ctx context.Context, item *domain.WishlistItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WishlistItem, error)
	GetByWishlistID(ctx context.Context, wishlistID uuid.UUID) ([]*domain.WishlistItem, error)
	GetByWishlistAndProduct(ctx context.Context, wishlistID, productID uuid.UUID, variationID *uuid.UUID) (*domain.WishlistItem, error)
	DeleteByWishlistID(ctx context.Context, wishlistID uuid.UUID) error
	BulkCreate(ctx context.Context, items []*domain.WishlistItem) error
	BulkDelete(ctx context.Context, itemIDs []uuid.UUID) error
	Exists(ctx context.Context, wishlistID, productID uuid.UUID, variationID *uuid.UUID) (bool, error)
}

// Combined POS repository interface
type POSRepositories struct {
	ProductCategory    ProductCategoryRepository
	Product            ProductRepository
	ProductVariation   ProductVariationRepository
	InventoryLog       InventoryLogRepository
	Cart               CartRepository
	CartItem           CartItemRepository
	Order              OrderRepository
	OrderItem          OrderItemRepository
	PaymentTransaction PaymentTransactionRepository
	Receipt            ReceiptRepository
	SalesReport        SalesReportRepository
	Discount           DiscountRepository
	Wishlist           WishlistRepository
	WishlistItem       WishlistItemRepository
}
