package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChange_ToHTMLString(t *testing.T) {
	testcases := []struct {
		name     string
		change   Change
		excepted string
	}{
		{
			name: "New Element",
			change: Change{
				TP:         NewElement,
				CourseName: "course_name",
				New:        []string{"new_arg_1", "new_arg_2", "new_arg_3", "new_arg_4-new_arg_4", "new_arg_5 %"},
			},
			excepted: "course_name\n🔆 <b>New:</b> new_arg_1 new_arg_5% (new_arg_3/new_arg_4-new_arg_4)",
		},
		{
			name: "Changed Element",
			change: Change{
				TP:         Changed,
				CourseName: "course_name",
				Old:        []string{"common_arg_1", "old_arg_2", "old_arg_3", "old_arg_4-old_arg_4", "old_arg_5 %"},
				New:        []string{"common_arg_1", "new_arg_2", "new_arg_3", "new_arg_4-new_arg_4", "new_arg_5 %"},
			},
			excepted: "course_name\n❇️ <b>Changes</b> in common_arg_1\n" +
				"Old: old_arg_5% (old_arg_3/old_arg_4-old_arg_4)\n" +
				"New: new_arg_5% (new_arg_3/new_arg_4-new_arg_4)",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.excepted, tc.change.ToHTMLString())
		})
	}
}
