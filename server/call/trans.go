package call

import "regexp"

type TransItem struct {
	Type  string   `json:"type"`
	Param []string `json:"param"`
}

func ApplyTrans(number string, trans []TransItem) (rn string, re error) {
	rn = number
	for _, t := range trans {
		rn, re = t.applyTrans(rn)
	}
	return
}

func (t TransItem) applyTrans(number string) (string, error) {

	var re error
	switch t.Type {
	case "prefix":
		if len(t.Param) > 0 {
			return t.Param[0] + number, nil
		}
	case "suffix":
		if len(t.Param) > 0 {
			return number + t.Param[0], nil
		}
	case "replace":
		if len(t.Param) > 1 {
			r, err := regexp.Compile(t.Param[0])
			if err != nil {
				re = err
				break
			}
			return r.ReplaceAllString(number, t.Param[1]), nil
		}
	}
	return number, re
}
