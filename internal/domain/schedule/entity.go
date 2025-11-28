package schedule

import "time"

type WorkSchedule struct {
	ID                 string
	CompanyID          string
	Name               string
	Type               WorkArrangement
	GracePeriodMinutes int
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time

	Times     []WorkScheduleTime
	Locations []WorkScheduleLocation
}

type WorkArrangement string

const (
	WorkArrangementWFO    WorkArrangement = "WFO"    // Work From Office
	WorkArrangementWFA    WorkArrangement = "WFA"    // Work From Anywhere
	WorkArrangementHybrid WorkArrangement = "Hybrid" // Hybrid Work Arrangement
)

var WorkArrangementValues = []string{
	string(WorkArrangementWFO),
	string(WorkArrangementWFA),
	string(WorkArrangementHybrid),
}

type WorkScheduleTime struct {
	ID                string
	WorkScheduleID    string
	DayOfWeek         int // 1=Monday, ..., 7=Sunday
	ClockInTime       time.Time
	BreakStartTime    *time.Time
	BreakEndTime      *time.Time
	ClockOutTime      time.Time
	IsNextDayCheckout bool // Indicates if checkout is on the next day
	LocationType      WorkArrangement
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type WorkScheduleLocation struct {
	ID             string
	WorkScheduleID string
	LocationName   string
	Latitude       float64
	Longitude      float64
	RadiusMeters   int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type EmployeeScheduleAssignment struct {
	ID             string
	EmployeeID     string
	WorkScheduleID string
	StartDate      time.Time
	EndDate        time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
