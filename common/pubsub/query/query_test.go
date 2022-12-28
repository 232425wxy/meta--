package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// 小数的整数部分必须不能为0
// 比较不了负数
// 只有字符串的两端需要加上单引号

func TestName(t *testing.T) {
	res := numRegex.FindString("-567")
	t.Log(res)
}

func TestMatches(t *testing.T) {
	testCases := []struct {
		query  string
		events map[string][]string
		match  bool
	}{
		{query: "block.validator.proto = 'node1'", events: map[string][]string{"block.validator.proto": {"node1", "node2"}}, match: true},
		{query: "block.validator.proto = 'node3'", events: map[string][]string{"block.validator.proto": {"node1", "node2"}}, match: false},
		{query: "block.height > 5", events: map[string][]string{"block.validator.proto": {"node1", "node2"}}, match: false},
		{query: "block.height > 5", events: map[string][]string{"block.height": {"3", "8"}}, match: true},
		{query: "block.height > 5", events: map[string][]string{"block.height": {"3", "4"}}, match: false},
		{query: "vote > 5", events: map[string][]string{"vote.rate": {"3", "7"}}, match: false},
		{query: "vote.rate > 1.667", events: map[string][]string{"vote.rate": {"0.667", "0.666"}}, match: false},
		{query: "vote.rate >= 1.667", events: map[string][]string{"vote.rate": {"1.667", "0.666"}}, match: true},
		{query: "vote.rate >= 1", events: map[string][]string{"vote.rate": {"0", "1"}}, match: true},
		{query: "block.height > 5 AND block.txs.num > 3000", events: map[string][]string{"block.height": {"6"}, "block.txs.num": {"3001"}}, match: true},
		{query: "block.height > 5 AND block.txs.num > 3001", events: map[string][]string{"block.height": {"6"}, "block.txs.num": {"3001"}}, match: false},
		{query: "block.date > DATE 2021-11-09", events: map[string][]string{"block.date": {"2022-11-09"}}, match: true},
		{query: "block.date > DATE 2021-11-09", events: map[string][]string{"block.date": {"2020-11-09"}}, match: false},
		{query: "block.time > TIME 2022-12-02T23:58:10+08:00", events: map[string][]string{"block.time": {time.Now().Format(TimeLayout)}}, match: true},
		{query: "block.time > TIME 2022-12-02T23:58:10+08:00", events: map[string][]string{"block.time": {"2022-12-02T23:58:10+08:00"}}, match: false},
		{query: "block.time = TIME 2022-12-03T00:00:40+08:00", events: map[string][]string{"block.time": {"2022-12-03T00:00:40+08:00"}}, match: true},
		{query: "node1 EXISTS", events: map[string][]string{"node1.power": {"23"}}, match: true},
		{query: "node1 EXISTS", events: map[string][]string{"node1": {}}, match: true},
		{query: "node1 exists", events: map[string][]string{"node1": {}}, match: true},
		{query: "node1 Exists", events: map[string][]string{"node1": {}}, match: true},
		{query: "node1 EXISTS", events: map[string][]string{"node2": {}}, match: false},
		{query: "node1 EXISTS and node1.power > 23", events: map[string][]string{"node2": {}, "node1.power": {"24"}}, match: true},
		{query: "node1 EXISTS AND node1.power > 23", events: map[string][]string{"node2": {}, "node1.power": {"24"}}, match: true},
		{query: "block.validator1 EXISTS AND block.height = 12", events: map[string][]string{"block.validator1": {"10.0"}, "block.height": {"12"}}, match: true},
		{query: "block.validators CONTAINS 'validator1'", events: map[string][]string{"block.validators": {"validator2", "validator3"}}, match: false},
		{query: "block.validators CONTAINS 'validator1'", events: map[string][]string{"block.validators": {"validator2", "validator1"}}, match: true},
	}

	for _, test := range testCases {
		t.Run(test.query, func(t *testing.T) {
			query, err := New(test.query)
			if err != nil {
				t.Log(test.query, err)
				return
			}
			res, err := query.Matches(test.events)
			if err != nil {
				t.Log(test.query, err)
				return
			}
			assert.Equal(t, test.match, res)
		})
	}
}
