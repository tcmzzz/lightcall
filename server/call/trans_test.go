package call

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrans(t *testing.T) {

	cs := []struct {
		raw      string
		t        TransItem
		expected string
		hasErr   bool
	}{
		{"aaa", TransItem{"prefix", []string{"1#"}}, "1#aaa", false},
		{"aaa", TransItem{"suffix", []string{"#1"}}, "aaa#1", false},
		{"010-333", TransItem{"replace", []string{`(\w+)-`, "$1"}}, "010333", false},
		{"020-333", TransItem{"replace", []string{`(\w+)-`, "$1"}}, "020333", false},
		{"0113-333", TransItem{"replace", []string{`(\w+)-`, "$1"}}, "0113333", false},
		{"0131-333", TransItem{"replace", []string{`(\w+)-`, "$1"}}, "0131333", false},
		{"0131-333", TransItem{"replace", []string{`(\w+)-`, ""}}, "333", false},
		{"0131-333", TransItem{"replace", []string{`(\w+-`, ""}}, "0131-333", true},
	}

	for _, c := range cs {
		out, err := c.t.applyTrans(c.raw)
		if c.hasErr {
			assert.Error(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, c.expected, out)
	}
}
