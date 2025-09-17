package call

import (
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/pocketbase/pocketbase/core"
)

type Result struct {
	OriCaller string
	OriCallee string
	Caller    string
	Callee    string
	Addr      string
}

func makeCall(app core.App, user *core.Record, taskID string) (*Result, error) {

	task, err := app.FindRecordById("task", taskID)
	if err != nil {
		return nil, err
	}

	// check userId. Admin or Owner
	own := task.GetString("own")
	isAdmin := user.GetBool("isAdmin")

	if task.GetString("own") != user.Id && !isAdmin {
		return nil, errors.Errorf("not allow to call(task owner: %s, user: %s, isAdmin: %v)", own, user.Id, isAdmin)
	}

	callee := task.GetString("callee")

	// find caller
	r, err := FindCaller(app, callee)
	if err != nil {
		return nil, err
	}
	caller := r.GetString("number")

	gw := r.ExpandedOne("outgw")
	if gw == nil || gw.Id == "" {
		return nil, errors.Errorf("number have no related outgw: %s", caller)
	}

	// Transform caller and callee
	transCallee := make([]TransItem, 0)
	transCaller := make([]TransItem, 0)

	addr := gw.GetString("addr")

	if err := gw.UnmarshalJSONField("transcaller", &transCaller); err != nil {
		app.Logger().Error("failed to parse transcaller", "err", err)
	}
	if err := gw.UnmarshalJSONField("transcallee", &transCallee); err != nil {
		app.Logger().Error("failed to parse transcallee", "err", err)
	}

	tCaller, err := ApplyTrans(caller, transCaller)
	if err != nil {
		app.Logger().Error("failed to trans transcaller", "err", err)
	}

	tCallee, err := ApplyTrans(callee, transCallee)
	if err != nil {
		app.Logger().Error("failed to trans transcallee", "err", err)
	}

	return &Result{
		OriCaller: caller,
		OriCallee: callee,
		Caller:    tCaller,
		Callee:    tCallee,
		Addr:      addr,
	}, nil
}

// TODO: system caller strategy
func FindCaller(app core.App, callee string) (*core.Record, error) {

	records, err := app.FindAllRecords("number")
	if err != nil {
		return nil, err
	}

	enabledRecords := make([]*core.Record, 0)
	for _, record := range records {
		if !record.GetBool("enable") {
			continue
		}

		errs := app.ExpandRecord(record, []string{"outgw"}, nil)
		if len(errs) > 0 {
			app.Logger().Warn("failed to expand gateway for number", "number", record.Id, "errors", errs)
			continue
		}

		gw := record.ExpandedOne("outgw")
		if gw == nil || !gw.GetBool("enable") {
			continue
		}

		enabledRecords = append(enabledRecords, record)
	}

	if len(enabledRecords) == 0 {
		return nil, errors.New("no enabled numbers with enabled gateways available")
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	ret := enabledRecords[rd.Intn(len(enabledRecords))]

	return ret, nil
}
