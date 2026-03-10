package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"moneyvault/internal/encryption"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
)

type NotificationService struct {
	notifRepo       *repositories.NotificationRepository
	budgetRepo      *repositories.BudgetRepository
	holdingRepo     *repositories.HoldingRepository
	analyticsRepo   *repositories.AnalyticsRepository
	categoryRepo    *repositories.CategoryRepository
	transactionRepo *repositories.TransactionRepository
	enc             *encryption.Service
	pushService     *PushService
}

func NewNotificationService(
	notifRepo *repositories.NotificationRepository,
	budgetRepo *repositories.BudgetRepository,
	holdingRepo *repositories.HoldingRepository,
	analyticsRepo *repositories.AnalyticsRepository,
	categoryRepo *repositories.CategoryRepository,
	transactionRepo *repositories.TransactionRepository,
	enc *encryption.Service,
) *NotificationService {
	return &NotificationService{
		notifRepo:       notifRepo,
		budgetRepo:      budgetRepo,
		holdingRepo:     holdingRepo,
		analyticsRepo:   analyticsRepo,
		categoryRepo:    categoryRepo,
		transactionRepo: transactionRepo,
		enc:             enc,
	}
}

// SetPushService sets the push service for sending browser push notifications.
func (s *NotificationService) SetPushService(ps *PushService) {
	s.pushService = ps
}

// Notification CRUD
func (s *NotificationService) List(userID uuid.UUID, unreadOnly bool, limit int) ([]models.Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.notifRepo.List(userID, unreadOnly, limit)
}

func (s *NotificationService) UnreadCount(userID uuid.UUID) (int, error) {
	return s.notifRepo.UnreadCount(userID)
}

func (s *NotificationService) MarkRead(userID, id uuid.UUID) error {
	return s.notifRepo.MarkRead(userID, id)
}

func (s *NotificationService) MarkAllRead(userID uuid.UUID) error {
	return s.notifRepo.MarkAllRead(userID)
}

func (s *NotificationService) Delete(userID, id uuid.UUID) error {
	return s.notifRepo.Delete(userID, id)
}

func (s *NotificationService) ClearAll(userID uuid.UUID) error {
	return s.notifRepo.ClearAll(userID)
}

func (s *NotificationService) CreateNotification(userID uuid.UUID, nType models.NotificationType, title, message, link string) error {
	n := &models.Notification{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    nType,
		Title:   title,
		Message: message,
		Link:    link,
	}
	err := s.notifRepo.Create(n)

	// Also send push notification if push service is configured
	if s.pushService != nil {
		go s.pushService.SendToUser(userID, title, message, link)
	}

	return err
}

// Alert Rules
func (s *NotificationService) CreateRule(userID uuid.UUID, req models.CreateAlertRuleRequest) (*models.AlertRule, error) {
	rule := &models.AlertRule{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      models.AlertRuleType(req.Type),
		Condition: req.Condition,
		IsActive:  true,
	}
	if err := s.notifRepo.CreateRule(rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *NotificationService) ListRules(userID uuid.UUID) ([]models.AlertRule, error) {
	return s.notifRepo.ListRules(userID)
}

func (s *NotificationService) ToggleRule(userID, id uuid.UUID) error {
	return s.notifRepo.ToggleRule(userID, id)
}

func (s *NotificationService) DeleteRule(userID, id uuid.UUID) error {
	return s.notifRepo.DeleteRule(userID, id)
}

// EvaluateAlerts checks all active rules and creates notifications
func (s *NotificationService) EvaluateAlerts() {
	rules, err := s.notifRepo.GetActiveRules()
	if err != nil {
		log.Printf("Alert evaluation: failed to get rules: %v", err)
		return
	}

	for _, rule := range rules {
		if rule.LastTriggered != nil && time.Since(*rule.LastTriggered) < 1*time.Hour {
			continue
		}

		triggered := false
		var title, message, link string

		switch rule.Type {
		case models.AlertBudgetOverspend:
			triggered, title, message, link = s.evalBudgetAlert(rule)
		case models.AlertPriceDrop, models.AlertPriceRise:
			triggered, title, message, link = s.evalPriceAlert(rule)
		case models.AlertNetWorthMilestone:
			triggered, title, message, link = s.evalNetWorthAlert(rule)
		}

		if triggered {
			nType := models.NotifInfo
			switch rule.Type {
			case models.AlertBudgetOverspend:
				nType = models.NotifBudgetAlert
			case models.AlertPriceDrop, models.AlertPriceRise:
				nType = models.NotifPriceAlert
			case models.AlertNetWorthMilestone:
				nType = models.NotifMilestone
			}
			_ = s.CreateNotification(rule.UserID, nType, title, message, link)
			_ = s.notifRepo.UpdateLastTriggered(rule.ID)
		}
	}
}

// calcSpentForCategory fetches encrypted transactions and sums amounts in application code.
func (s *NotificationService) calcSpentForCategory(userID uuid.UUID, categoryID *uuid.UUID, start, end time.Time) (float64, error) {
	txs, err := s.transactionRepo.ListExpensesByCategory(userID, categoryID, start, end)
	if err != nil {
		return 0, err
	}
	var total float64
	for _, tx := range txs {
		plain, err := s.enc.DecryptField(userID, tx.Amount)
		if err != nil {
			continue
		}
		amt, err := strconv.ParseFloat(plain, 64)
		if err != nil {
			continue
		}
		total += amt
	}
	return total, nil
}

func (s *NotificationService) evalBudgetAlert(rule models.AlertRule) (bool, string, string, string) {
	var cond models.BudgetAlertCondition
	if err := json.Unmarshal(rule.Condition, &cond); err != nil {
		return false, "", "", ""
	}
	if cond.Threshold <= 0 {
		cond.Threshold = 100
	}

	budgets, err := s.budgetRepo.List(rule.UserID)
	if err != nil {
		return false, "", "", ""
	}

	categories, _ := s.categoryRepo.ListByUser(rule.UserID)
	catMap := make(map[string]string)
	for _, c := range categories {
		catMap[c.ID.String()] = c.Name
	}

	now := time.Now()
	for _, b := range budgets {
		if b.CategoryID == nil {
			continue
		}

		periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		periodEnd := periodStart.AddDate(0, 1, 0)

		spent, err := s.calcSpentForCategory(rule.UserID, b.CategoryID, periodStart, periodEnd)
		if err != nil {
			continue
		}

		if b.Amount > 0 {
			pct := (spent / b.Amount) * 100
			if pct >= cond.Threshold {
				catName := catMap[b.CategoryID.String()]
				if catName == "" {
					catName = "Unknown"
				}
				return true,
					fmt.Sprintf("Budget alert: %s", catName),
					fmt.Sprintf("You've spent %.0f%% of your %s budget ($%.2f / $%.2f)", pct, catName, spent, b.Amount),
					"/budgets"
			}
		}
	}
	return false, "", "", ""
}

func (s *NotificationService) evalPriceAlert(rule models.AlertRule) (bool, string, string, string) {
	var cond models.PriceAlertCondition
	if err := json.Unmarshal(rule.Condition, &cond); err != nil {
		return false, "", "", ""
	}

	price, err := s.holdingRepo.GetPrice(cond.Symbol)
	if err != nil {
		return false, "", "", ""
	}

	if rule.Type == models.AlertPriceRise && price.Price >= cond.Price {
		return true,
			fmt.Sprintf("%s price alert", cond.Symbol),
			fmt.Sprintf("%s is now $%.2f (target: $%.2f)", cond.Symbol, price.Price, cond.Price),
			"/investments"
	}
	if rule.Type == models.AlertPriceDrop && price.Price <= cond.Price {
		return true,
			fmt.Sprintf("%s price alert", cond.Symbol),
			fmt.Sprintf("%s dropped to $%.2f (threshold: $%.2f)", cond.Symbol, price.Price, cond.Price),
			"/investments"
	}
	return false, "", "", ""
}

func (s *NotificationService) evalNetWorthAlert(rule models.AlertRule) (bool, string, string, string) {
	var cond models.NetWorthMilestoneCondition
	if err := json.Unmarshal(rule.Condition, &cond); err != nil {
		return false, "", "", ""
	}

	latest, err := s.analyticsRepo.GetLatestNetWorth(rule.UserID)
	if err != nil {
		return false, "", "", ""
	}

	if latest.TotalValue >= cond.Amount {
		return true,
			"Net worth milestone reached!",
			fmt.Sprintf("Your net worth has reached $%.2f (milestone: $%.2f)", latest.TotalValue, cond.Amount),
			"/reports"
	}
	return false, "", "", ""
}
