package domain

type Slot struct {
	SlotIndex  int    `json:"slot_index"`
	CourseCode string `json:"course_code"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Venue      string `json:"venue"`
	Status     string `json:"status"`
}

type Timetable struct {
	ClassID string `json:"class_id"`
	Date    string `json:"date"`
	Slots   []Slot `json:"slots"`
}

type UpdateOverrideRequest struct {
	ClassID    string `json:"class_id"`
	SlotIndex  int    `json:"slot_index"`
	CourseCode string `json:"course_code"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Venue      string `json:"venue"`
	Status     string `json:"status"`
}

type DeleteSlotRequest struct {
	ClassID   string `json:"class_id"`
	SlotIndex int    `json:"slot_index"`
}
