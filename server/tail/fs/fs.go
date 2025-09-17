package fs

import (
	"encoding/json"

	"github.com/patrickmn/go-cache"
	"github.com/pocketbase/pocketbase/core"
)

type Handler struct {
	MasterFile string
	RecordDir  string
}

func (c *Handler) File() string { return c.MasterFile }

func processMatchedCdr(app core.App, recordDir string, aleg *CdrLine, bleg *CdrLine) error {
	return aleg.LoadWithBleg(app, recordDir, bleg)
}

func (c *Handler) Deal(app core.App, line string) error {

	cdrLine := &CdrLine{}
	if err := json.Unmarshal([]byte(line), cdrLine); err != nil {
		return err
	}

	if cdrLine.Originator == "" { // This is an aleg
		if err := processALeg(app, c.RecordDir, cdrLine); err != nil {
			return err
		}
	} else { // This is a bleg
		if err := processBLeg(app, c.RecordDir, cdrLine); err != nil {
			return err
		}
	}

	return nil
}

func processALeg(app core.App, recordDir string, cdrLine *CdrLine) error {
	// Check if a bleg is waiting for this aleg
	blegRaw, found := blegCache.Get(cdrLine.UUID)
	if !found {
		// No bleg found, cache this aleg
		alegCache.Set(cdrLine.UUID, cdrLine, cache.DefaultExpiration)
		return nil
	}

	bleg, ok := blegRaw.(*CdrLine)
	if !ok {
		return nil
	}

	// Found a match, process them
	if err := processMatchedCdr(app, recordDir, cdrLine, bleg); err != nil {
		return err
	}
	blegCache.Delete(cdrLine.UUID)
	return nil
}

func processBLeg(app core.App, recordDir string, cdrLine *CdrLine) error {
	// Check if an aleg is waiting for this bleg
	alegRaw, found := alegCache.Get(cdrLine.Originator)
	if !found {
		// No aleg found, cache this bleg
		blegCache.Set(cdrLine.Originator, cdrLine, cache.DefaultExpiration)
		return nil
	}

	aleg, ok := alegRaw.(*CdrLine)
	if !ok {
		return nil
	}

	// Found a match, process them
	if err := processMatchedCdr(app, recordDir, aleg, cdrLine); err != nil {
		return err
	}
	alegCache.Delete(cdrLine.Originator)
	return nil
}
