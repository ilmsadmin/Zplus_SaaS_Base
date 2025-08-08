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

// ProductCategoryService handles product category business logic
type ProductCategoryService struct {
	repo repositories.ProductCategoryRepository
}

func NewProductCategoryService(repo repositories.ProductCategoryRepository) *ProductCategoryService {
	return &ProductCategoryService{repo: repo}
}

func (s *ProductCategoryService) CreateCategory(ctx context.Context, category *domain.ProductCategory) error {
	// Validate category
	if err := s.validateCategory(ctx, category); err != nil {
		return err
	}

	// Check if name already exists
	exists, err := s.repo.Exists(ctx, category.TenantID, category.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to check category existence: %w", err)
	}
	if exists {
		return errors.New("category with this name already exists")
	}

	// Generate slug from name if not provided
	if category.Slug == "" {
		category.Slug = generateSlug(category.Name)
	}

	return s.repo.Create(ctx, category)
}

func (s *ProductCategoryService) UpdateCategory(ctx context.Context, category *domain.ProductCategory) error {
	// Validate category
	if err := s.validateCategory(ctx, category); err != nil {
		return err
	}

	// Check if name already exists (excluding current category)
	exists, err := s.repo.Exists(ctx, category.TenantID, category.Name, &category.ID)
	if err != nil {
		return fmt.Errorf("failed to check category existence: %w", err)
	}
	if exists {
		return errors.New("category with this name already exists")
	}

	return s.repo.Update(ctx, category)
}

func (s *ProductCategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	// Check if category has children
	children, err := s.repo.GetChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check child categories: %w", err)
	}
	if len(children) > 0 {
		return errors.New("cannot delete category that has child categories")
	}

	return s.repo.Delete(ctx, id)
}

func (s *ProductCategoryService) GetCategoryTree(ctx context.Context, tenantID string) ([]*domain.ProductCategory, error) {
	return s.repo.GetTreeByTenantID(ctx, tenantID)
}

func (s *ProductCategoryService) validateCategory(ctx context.Context, category *domain.ProductCategory) error {
	if category.Name == "" {
		return errors.New("category name is required")
	}

	// Validate parent category if provided
	if category.ParentID != nil {
		parent, err := s.repo.GetByID(ctx, *category.ParentID)
		if err != nil {
			return fmt.Errorf("failed to validate parent category: %w", err)
		}
		if parent.TenantID != category.TenantID {
			return errors.New("parent category must belong to the same tenant")
		}
	}

	return nil
}

// ProductService handles product business logic
type ProductService struct {
	repo          repositories.ProductRepository
	variationRepo repositories.ProductVariationRepository
	inventoryRepo repositories.InventoryLogRepository
}

func NewProductService(
	repo repositories.ProductRepository,
	variationRepo repositories.ProductVariationRepository,
	inventoryRepo repositories.InventoryLogRepository,
) *ProductService {
	return &ProductService{
		repo:          repo,
		variationRepo: variationRepo,
		inventoryRepo: inventoryRepo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *domain.Product) error {
	// Validate product
	if err := s.validateProduct(ctx, product); err != nil {
		return err
	}

	// Check if SKU already exists
	exists, err := s.repo.Exists(ctx, product.TenantID, product.SKU, nil)
	if err != nil {
		return fmt.Errorf("failed to check product SKU existence: %w", err)
	}
	if exists {
		return errors.New("product with this SKU already exists")
	}

	// Generate slug from name if not provided
	if product.Slug == "" {
		product.Slug = generateSlug(product.Name)
	}

	// Set default values
	if product.ProductType == "" {
		product.ProductType = domain.ProductTypeSimple
	}
	if product.Status == "" {
		product.Status = domain.ProductStatusDraft
	}
	if product.StockStatus == "" {
		product.StockStatus = domain.StockStatusInStock
	}

	// Create product
	if err := s.repo.Create(ctx, product); err != nil {
		return err
	}

	// Create initial inventory log if stock quantity > 0
	if product.StockQuantity > 0 {
		inventoryLog := &domain.InventoryLog{
			TenantID:       product.TenantID,
			ProductID:      product.ID,
			Type:           domain.InventoryTypeIn,
			Quantity:       product.StockQuantity,
			QuantityBefore: 0,
			QuantityAfter:  product.StockQuantity,
			Reason:         "Initial stock",
			CostPerUnit:    product.CostPrice,
			TotalCost:      product.CostPrice * float64(product.StockQuantity),
		}
		if err := s.inventoryRepo.Create(ctx, inventoryLog); err != nil {
			// Log error but don't fail product creation
			// TODO: Add proper logging
		}
	}

	return nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, product *domain.Product) error {
	// Get current product for comparison
	currentProduct, err := s.repo.GetByID(ctx, product.ID)
	if err != nil {
		return fmt.Errorf("failed to get current product: %w", err)
	}

	// Validate product
	if err := s.validateProduct(ctx, product); err != nil {
		return err
	}

	// Check if SKU already exists (excluding current product)
	exists, err := s.repo.Exists(ctx, product.TenantID, product.SKU, &product.ID)
	if err != nil {
		return fmt.Errorf("failed to check product SKU existence: %w", err)
	}
	if exists {
		return errors.New("product with this SKU already exists")
	}

	// Handle stock quantity changes
	if product.ManageStock && product.StockQuantity != currentProduct.StockQuantity {
		difference := product.StockQuantity - currentProduct.StockQuantity
		logType := domain.InventoryTypeAdjustment
		if difference > 0 {
			logType = domain.InventoryTypeIn
		} else {
			logType = domain.InventoryTypeOut
		}

		inventoryLog := &domain.InventoryLog{
			TenantID:       product.TenantID,
			ProductID:      product.ID,
			Type:           logType,
			Quantity:       abs(difference),
			QuantityBefore: currentProduct.StockQuantity,
			QuantityAfter:  product.StockQuantity,
			Reason:         "Stock adjustment",
			CostPerUnit:    product.CostPrice,
			TotalCost:      product.CostPrice * float64(abs(difference)),
		}
		if err := s.inventoryRepo.Create(ctx, inventoryLog); err != nil {
			// Log error but don't fail product update
			// TODO: Add proper logging
		}
	}

	return s.repo.Update(ctx, product)
}

func (s *ProductService) UpdateStock(ctx context.Context, productID uuid.UUID, quantity int, reason string, userID *uuid.UUID) error {
	product, err := s.repo.GetByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if !product.ManageStock {
		return errors.New("stock management is disabled for this product")
	}

	previousQuantity := product.StockQuantity

	// Update product stock
	if err := s.repo.UpdateStock(ctx, productID, quantity); err != nil {
		return err
	}

	// Create inventory log
	difference := quantity - previousQuantity
	logType := domain.InventoryTypeAdjustment
	if difference > 0 {
		logType = domain.InventoryTypeIn
	} else if difference < 0 {
		logType = domain.InventoryTypeOut
	}

	inventoryLog := &domain.InventoryLog{
		TenantID:       product.TenantID,
		ProductID:      productID,
		Type:           logType,
		Quantity:       abs(difference),
		QuantityBefore: previousQuantity,
		QuantityAfter:  quantity,
		Reason:         reason,
		CostPerUnit:    product.CostPrice,
		TotalCost:      product.CostPrice * float64(abs(difference)),
		UserID:         userID,
	}

	return s.inventoryRepo.Create(ctx, inventoryLog)
}

func (s *ProductService) validateProduct(ctx context.Context, product *domain.Product) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.SKU == "" {
		return errors.New("product SKU is required")
	}
	if product.RegularPrice < 0 {
		return errors.New("regular price cannot be negative")
	}
	if product.SalePrice < 0 {
		return errors.New("sale price cannot be negative")
	}
	if product.CostPrice < 0 {
		return errors.New("cost price cannot be negative")
	}
	if product.StockQuantity < 0 {
		return errors.New("stock quantity cannot be negative")
	}

	return nil
}

// CartService handles shopping cart business logic
type CartService struct {
	cartRepo     repositories.CartRepository
	cartItemRepo repositories.CartItemRepository
	productRepo  repositories.ProductRepository
}

func NewCartService(
	cartRepo repositories.CartRepository,
	cartItemRepo repositories.CartItemRepository,
	productRepo repositories.ProductRepository,
) *CartService {
	return &CartService{
		cartRepo:     cartRepo,
		cartItemRepo: cartItemRepo,
		productRepo:  productRepo,
	}
}

func (s *CartService) GetOrCreateCart(ctx context.Context, tenantID string, userID *uuid.UUID, sessionID string) (*domain.Cart, error) {
	var cart *domain.Cart
	var err error

	// Try to get existing cart
	if userID != nil {
		cart, err = s.cartRepo.GetByUserID(ctx, tenantID, *userID)
	} else {
		cart, err = s.cartRepo.GetBySessionID(ctx, tenantID, sessionID)
	}

	if err != nil || cart == nil {
		// Create new cart
		cart = &domain.Cart{
			TenantID:  tenantID,
			UserID:    userID,
			SessionID: sessionID,
			Status:    domain.CartStatusActive,
			Currency:  "USD", // TODO: Make configurable
		}
		if err := s.cartRepo.Create(ctx, cart); err != nil {
			return nil, fmt.Errorf("failed to create cart: %w", err)
		}
	}

	return cart, nil
}

func (s *CartService) AddToCart(ctx context.Context, cartID, productID uuid.UUID, variationID *uuid.UUID, quantity int) error {
	// Validate product and availability
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if product.Status != domain.ProductStatusPublished {
		return errors.New("product is not available")
	}

	// Check stock availability
	availableStock := product.StockQuantity
	if product.ManageStock && availableStock < quantity {
		return fmt.Errorf("insufficient stock: only %d available", availableStock)
	}

	// Check if item already exists in cart
	existingItem, err := s.cartItemRepo.GetByCartAndProduct(ctx, cartID, productID, variationID)
	if err == nil && existingItem != nil {
		// Update quantity
		newQuantity := existingItem.Quantity + quantity
		if product.ManageStock && availableStock < newQuantity {
			return fmt.Errorf("insufficient stock: only %d available", availableStock)
		}

		existingItem.Quantity = newQuantity
		existingItem.TotalPrice = existingItem.UnitPrice * float64(newQuantity)

		if err := s.cartItemRepo.Update(ctx, existingItem); err != nil {
			return fmt.Errorf("failed to update cart item: %w", err)
		}
	} else {
		// Create new cart item
		unitPrice := product.SalePrice
		if unitPrice == 0 {
			unitPrice = product.RegularPrice
		}

		cartItem := &domain.CartItem{
			CartID:      cartID,
			ProductID:   productID,
			VariationID: variationID,
			Quantity:    quantity,
			UnitPrice:   unitPrice,
			TotalPrice:  unitPrice * float64(quantity),
			ProductData: map[string]interface{}{
				"name":          product.Name,
				"sku":           product.SKU,
				"image":         product.FeaturedImage,
				"regular_price": product.RegularPrice,
				"sale_price":    product.SalePrice,
			},
		}

		if err := s.cartItemRepo.Create(ctx, cartItem); err != nil {
			return fmt.Errorf("failed to create cart item: %w", err)
		}
	}

	// Recalculate cart totals
	return s.RecalculateCartTotals(ctx, cartID)
}

func (s *CartService) RemoveFromCart(ctx context.Context, cartItemID uuid.UUID) error {
	cartItem, err := s.cartItemRepo.GetByID(ctx, cartItemID)
	if err != nil {
		return fmt.Errorf("failed to get cart item: %w", err)
	}

	if err := s.cartItemRepo.Delete(ctx, cartItemID); err != nil {
		return fmt.Errorf("failed to delete cart item: %w", err)
	}

	// Recalculate cart totals
	return s.RecalculateCartTotals(ctx, cartItem.CartID)
}

func (s *CartService) UpdateCartItemQuantity(ctx context.Context, cartItemID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return s.RemoveFromCart(ctx, cartItemID)
	}

	cartItem, err := s.cartItemRepo.GetByID(ctx, cartItemID)
	if err != nil {
		return fmt.Errorf("failed to get cart item: %w", err)
	}

	// Check stock availability
	product, err := s.productRepo.GetByID(ctx, cartItem.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if product.ManageStock && product.StockQuantity < quantity {
		return fmt.Errorf("insufficient stock: only %d available", product.StockQuantity)
	}

	cartItem.Quantity = quantity
	cartItem.TotalPrice = cartItem.UnitPrice * float64(quantity)

	if err := s.cartItemRepo.Update(ctx, cartItem); err != nil {
		return fmt.Errorf("failed to update cart item: %w", err)
	}

	// Recalculate cart totals
	return s.RecalculateCartTotals(ctx, cartItem.CartID)
}

func (s *CartService) RecalculateCartTotals(ctx context.Context, cartID uuid.UUID) error {
	items, err := s.cartItemRepo.GetByCartID(ctx, cartID)
	if err != nil {
		return fmt.Errorf("failed to get cart items: %w", err)
	}

	var subtotal float64
	for _, item := range items {
		subtotal += item.TotalPrice
	}

	// TODO: Calculate tax, shipping, discounts
	taxTotal := 0.0      // Calculate based on tax rules
	shippingTotal := 0.0 // Calculate based on shipping rules
	discountTotal := 0.0 // Calculate based on applied discounts
	total := subtotal + taxTotal + shippingTotal - discountTotal

	totals := domain.CartTotals{
		Subtotal:      subtotal,
		TaxTotal:      taxTotal,
		ShippingTotal: shippingTotal,
		DiscountTotal: discountTotal,
		Total:         total,
	}

	return s.cartRepo.UpdateTotals(ctx, cartID, totals)
}

func (s *CartService) ClearCart(ctx context.Context, cartID uuid.UUID) error {
	if err := s.cartItemRepo.DeleteByCartID(ctx, cartID); err != nil {
		return fmt.Errorf("failed to clear cart items: %w", err)
	}

	// Reset cart totals
	totals := domain.CartTotals{}
	return s.cartRepo.UpdateTotals(ctx, cartID, totals)
}

// Helper functions
func generateSlug(name string) string {
	// TODO: Implement proper slug generation
	// For now, return a simple version
	return fmt.Sprintf("%s-%d", name, time.Now().Unix())
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
