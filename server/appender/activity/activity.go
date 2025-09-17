package activity

import (
	"encoding/json"
	"time"

	"github.com/tcmzzz/lightcall/server/appender"

	"github.com/pkg/errors"
	"github.com/pocketbase/pocketbase/core"
)

type Event struct {
	EventType    string                 `json:"eventType"`
	TaskID       string                 `json:"taskId"`
	TaskExtID    string                 `json:"taskExtId"`
	ActivityID   string                 `json:"activityId"`
	Timestamp    time.Time              `json:"timestamp"`
	ActivityData map[string]interface{} `json:"activityData"`
}

type Handler struct {
	LogFile string
}

func (h *Handler) File() string {
	return h.LogFile
}

func (h *Handler) MustRegister(app core.App, appender *appender.Appender) error {

	app.OnRecordAfterUpdateSuccess("task").BindFunc(func(e *core.RecordEvent) error {

		task := e.Record
		if task.GetString("ext_id") == task.Id || task.GetString("ext_id") == "" {
			return e.Next()
		}

		before := e.Record.Original().GetStringSlice("activity")
		after := e.Record.GetStringSlice("activity")

		// Find new activities
		newActivities := difference(after, before)
		for _, activityID := range newActivities {
			if err := h.handleActivityCreated(e.App, activityID, e.Record, appender); err != nil {
				e.App.Logger().Error("Failed to handle new activity in task update", "err", err)
			}
		}

		return e.Next()

	})
	return nil
}

func difference(sliceA, sliceB []string) []string {
	m := make(map[string]bool)
	for _, item := range sliceB {
		m[item] = true
	}

	var diff []string
	for _, item := range sliceA {
		if _, found := m[item]; !found {
			diff = append(diff, item)
		}
	}
	return diff
}

func (h *Handler) handleActivityCreated(app core.App, activityID string, task *core.Record, activityAppender *appender.Appender) error {

	activity, err := app.FindRecordById("activity", activityID)
	if err != nil {
		return errors.Wrapf(err, "failed to find activity %s", activityID)
	}

	event := Event{
		EventType:  "activity_created",
		TaskID:     task.Id,
		TaskExtID:  task.GetString("ext_id"),
		ActivityID: activity.Id,
		Timestamp:  time.Now(),
		ActivityData: map[string]interface{}{
			"comment": activity.GetString("comment"),
			"isCall":  activity.GetBool("isCall"),
			"user":    activity.GetString("user"),
			"record":  activity.GetString("record"),
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "failed to marshal activity event")
	}

	activityAppender.Append(string(data))

	return nil
}
