package leave

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
)

type QuotaCalculator struct {
}

func NewQuotaCalculator() *QuotaCalculator {
	return &QuotaCalculator{}
}

func (c *QuotaCalculator) CalculateQuota(ctx context.Context, employee employee.Employee, leaveType leave.LeaveType) (float64, error) {
	if leaveType.QuotaCalculationType == "fixed" {
		return float64(leaveType.QuotaRules.DefaultQuota), nil
	}

	var quota float64
	var err error

	switch leaveType.QuotaCalculationType {
	case "tenure":
		quota, err = c.calculateTenureBased(&employee, &leaveType.QuotaRules)
	}

	if err != nil {
		return 0, err
	}

	return quota, nil
}

// calculateTenureBased calculates quota based on employee tenure
func (c *QuotaCalculator) calculateTenureBased(
	emp *employee.Employee,
	rules *leave.QuotaRules,
) (float64, error) {
	if len(rules.Rules) == 0 {
		return 0, errors.New("no tenure rules defined")
	}

	tenureMonths := c.calculateTenureMonths(emp.HireDate)

	// c.logger.Debug("Calculating tenure-based quota",
	// 	zap.Int("tenure_months", tenureMonths),
	// )

	for _, rule := range rules.Rules {
		minMonths := 0
		if rule.MinMonths != nil {
			minMonths = *rule.MinMonths
		}

		maxMonths := 999999 // No limit
		if rule.MaxMonths != nil {
			maxMonths = *rule.MaxMonths
		}

		if tenureMonths >= minMonths && tenureMonths < maxMonths {
			// c.logger.Debug("Matched tenure rule",
			// 	zap.Int("min_months", minMonths),
			// 	zap.Int("max_months", maxMonths),
			// 	zap.Float64("quota", rule.Quota),
			// )
			return rule.Quota, nil
		}
	}

	// Return default quota if no rule matched
	if rules.DefaultQuota > 0 {
		return rules.DefaultQuota, nil
	}

	return 0, fmt.Errorf("no matching tenure rule for %d months", tenureMonths)
}

// calculatePositionBased calculates quota based on position
func (c *QuotaCalculator) calculatePositionBased(
	ctx context.Context,
	emp *employee.Employee,
	rules *leave.QuotaRules,
) (float64, error) {
	if len(rules.Rules) == 0 {
		return 0, errors.New("no position rules defined")
	}

	// c.logger.Debug("Calculating position-based quota",
	// 	zap.String("position_id", emp.PositionID),
	// )

	for _, rule := range rules.Rules {
		for _, positionID := range rule.PositionIDs {
			if positionID == emp.PositionID {
				// c.logger.Debug("Matched position rule",
				// 	zap.String("position_id", positionID),
				// 	zap.Float64("quota", rule.Quota),
				// )
				return rule.Quota, nil
			}
		}
	}

	// Return default quota if no rule matched
	if rules.DefaultQuota > 0 {
		return rules.DefaultQuota, nil
	}

	return 0, fmt.Errorf("no matching position rule for position %s", emp.PositionID)
}

// calculateGradeBased calculates quota based on grade
func (c *QuotaCalculator) calculateGradeBased(
	ctx context.Context,
	emp *employee.Employee,
	rules *leave.QuotaRules,
) (float64, error) {
	if emp.GradeID == "" {
		// c.logger.Warn("Employee has no grade, using default quota",
		// 	zap.String("employee_id", emp.ID),
		// )
		if rules.DefaultQuota > 0 {
			return rules.DefaultQuota, nil
		}
		return 0, errors.New("employee has no grade")
	}

	if len(rules.Rules) == 0 {
		return 0, errors.New("no grade rules defined")
	}

	// c.logger.Debug("Calculating grade-based quota",
	// 	zap.String("grade_id", emp.GradeID),
	// )

	for _, rule := range rules.Rules {
		for _, gradeID := range rule.GradeIDs {
			if gradeID == emp.GradeID {
				// c.logger.Debug("Matched grade rule",
				// 	zap.String("grade_id", gradeID),
				// 	zap.Float64("quota", rule.Quota),
				// )
				return rule.Quota, nil
			}
		}
	}

	// Return default quota if no rule matched
	if rules.DefaultQuota > 0 {
		return rules.DefaultQuota, nil
	}

	return 0, fmt.Errorf("no matching grade rule for grade %s", emp.GradeID)
}

// calculateEmploymentTypeBased calculates quota based on employment type
func (c *QuotaCalculator) calculateEmploymentTypeBased(
	emp *employee.Employee,
	rules *leave.QuotaRules,
) (float64, error) {
	if len(rules.Rules) == 0 {
		return 0, errors.New("no employment type rules defined")
	}

	// c.logger.Debug("Calculating employment type-based quota",
	// 	zap.String("employment_type", string(emp.EmploymentType)),
	// )

	for _, rule := range rules.Rules {
		if rule.EmploymentType == string(emp.EmploymentType) {
			// c.logger.Debug("Matched employment type rule",
			// 	zap.String("employment_type", rule.EmploymentType),
			// 	zap.Float64("quota", rule.Quota),
			// )
			return rule.Quota, nil
		}
	}

	// Return default quota if no rule matched
	if rules.DefaultQuota > 0 {
		return rules.DefaultQuota, nil
	}

	return 0, fmt.Errorf("no matching employment type rule for %s", emp.EmploymentType)
}

// calculateCombined handles complex rules with multiple conditions
func (c *QuotaCalculator) calculateCombined(
	ctx context.Context,
	emp *employee.Employee,
	rules *leave.QuotaRules,
) (float64, error) {
	if len(rules.Rules) == 0 {
		return 0, errors.New("no combined rules defined")
	}

	tenureMonths := c.calculateTenureMonths(emp.HireDate)

	// c.logger.Debug("Calculating combined quota",
	// 	zap.String("employee_id", emp.ID),
	// 	zap.Int("tenure_months", tenureMonths),
	// )

	// Evaluate rules in order (first match wins)
	for _, rule := range rules.Rules {
		if rule.Conditions == nil {
			continue
		}

		matched := c.matchesConditions(emp, tenureMonths, rule.Conditions)

		if matched {
			// c.logger.Debug("Matched combined rule",
			// 	zap.Int("rule_index", i),
			// 	zap.Float64("quota", rule.Quota),
			// )
			return rule.Quota, nil
		}
	}

	// Return default quota if no rule matched
	if rules.DefaultQuota > 0 {
		// c.logger.Debug("No rule matched, using default quota",
		// 	zap.Float64("default_quota", rules.DefaultQuota),
		// )
		return rules.DefaultQuota, nil
	}

	return 0, errors.New("no matching combined rule")
}

// matchesConditions checks if employee matches all conditions
func (c *QuotaCalculator) matchesConditions(
	emp *employee.Employee,
	tenureMonths int,
	conditions *leave.QuotaConditions,
) bool {
	// Check position IDs
	if len(conditions.PositionIDs) > 0 {
		matched := false
		for _, posID := range conditions.PositionIDs {
			if posID == emp.PositionID {
				matched = true
				break
			}
		}
		if !matched {
			// c.logger.Debug("Position condition not matched",
			// 	zap.String("employee_position", emp.PositionID),
			// 	zap.Strings("required_positions", conditions.PositionIDs),
			// )
			return false
		}
	}

	// Check grade IDs
	if len(conditions.GradeIDs) > 0 {
		if emp.GradeID == "" {
			// c.logger.Debug("Grade condition not matched: employee has no grade")
			return false
		}

		matched := false
		for _, gradeID := range conditions.GradeIDs {
			if gradeID == emp.GradeID {
				matched = true
				break
			}
		}
		if !matched {
			// c.logger.Debug("Grade condition not matched",
			// 	zap.String("employee_grade", emp.GradeID),
			// 	zap.Strings("required_grades", conditions.GradeIDs),
			// )
			return false
		}
	}

	// Check employment type
	if conditions.EmploymentType != "" {
		if conditions.EmploymentType != string(emp.EmploymentType) {
			// c.logger.Debug("Employment type condition not matched",
			// 	zap.String("employee_type", string(emp.EmploymentType)),
			// 	zap.String("required_type", conditions.EmploymentType),
			// )
			return false
		}
	}

	// Check minimum tenure
	if conditions.MinTenureMonths != nil {
		if tenureMonths < *conditions.MinTenureMonths {
			// c.logger.Debug("Min tenure condition not matched",
			// 	zap.Int("employee_tenure", tenureMonths),
			// 	zap.Int("required_min", *conditions.MinTenureMonths),
			// )
			return false
		}
	}

	// Check maximum tenure
	if conditions.MaxTenureMonths != nil {
		if tenureMonths >= *conditions.MaxTenureMonths {
			// c.logger.Debug("Max tenure condition not matched",
			// 	zap.Int("employee_tenure", tenureMonths),
			// 	zap.Int("required_max", *conditions.MaxTenureMonths),
			// )
			return false
		}
	}

	// c.logger.Debug("All conditions matched")
	return true
}

// calculateTenureMonths calculates tenure in months
func (c *QuotaCalculator) calculateTenureMonths(hireDate time.Time) int {
	now := time.Now()

	years := now.Year() - hireDate.Year()
	months := int(now.Month()) - int(hireDate.Month())

	totalMonths := years*12 + months

	// Adjust if day hasn't passed yet
	if now.Day() < hireDate.Day() {
		totalMonths--
	}

	// Ensure non-negative
	if totalMonths < 0 {
		totalMonths = 0
	}

	return totalMonths
}

// CalculateAccruedQuota calculates accrued quota for monthly accrual method
func (c *QuotaCalculator) CalculateAccruedQuota(
	hireDate time.Time,
	annualQuota float64,
	asOfDate time.Time,
) float64 {
	yearStart := time.Date(asOfDate.Year(), 1, 1, 0, 0, 0, 0, asOfDate.Location())

	// If hired this year, start from hire date
	if hireDate.Year() == asOfDate.Year() {
		yearStart = hireDate
	}

	// Calculate months worked
	monthsWorked := 0.0
	currentDate := yearStart

	for currentDate.Before(asOfDate) || currentDate.Equal(asOfDate) {
		if currentDate.Day() == 1 || currentDate.Equal(yearStart) {
			monthsWorked++
		}
		currentDate = currentDate.AddDate(0, 1, 0)
		if currentDate.After(asOfDate) {
			break
		}
	}

	// Pro-rated calculation
	monthlyQuota := annualQuota / 12.0
	accruedQuota := monthlyQuota * monthsWorked

	// Cap at annual quota
	if accruedQuota > annualQuota {
		accruedQuota = annualQuota
	}

	return accruedQuota
}
