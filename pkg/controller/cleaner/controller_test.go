package cleaner

import (
	"testing"

	batchv1 "k8s.io/api/batch/v1"
)

func TestFilterJobFromCondition(t *testing.T) {
	testCases := map[string]struct {
		conditions      []batchv1.JobConditionType
		targetCondition batchv1.JobConditionType
		expect          bool
	}{
		"All conditions are Complete and target condition is Complete": {
			[]batchv1.JobConditionType{batchv1.JobComplete, batchv1.JobComplete},
			batchv1.JobComplete,
			true,
		},
		"Conditions Failed -> Complete and target condition is Complete": {
			[]batchv1.JobConditionType{batchv1.JobFailed, batchv1.JobComplete},
			batchv1.JobComplete,
			true,
		},
		"All conditions are Failed and target condition is Complete": {
			[]batchv1.JobConditionType{batchv1.JobFailed, batchv1.JobFailed},
			batchv1.JobComplete,
			false,
		},
		"All conditions are Complete and target condition is Failed": {
			[]batchv1.JobConditionType{batchv1.JobComplete, batchv1.JobComplete},
			batchv1.JobFailed,
			false,
		},
		"Conditions Failed -> Complete and target condition is Failed": {
			[]batchv1.JobConditionType{batchv1.JobFailed, batchv1.JobComplete},
			batchv1.JobFailed,
			false,
		},
		"All conditions are Failed and target condition is Failed": {
			[]batchv1.JobConditionType{batchv1.JobFailed, batchv1.JobFailed},
			batchv1.JobFailed,
			true,
		},
	}

	for name, tc := range testCases {
		job := createJobFromCondition(tc.conditions)
		actual := isJobCondition(job, tc.targetCondition)
		if actual != tc.expect {
			t.Errorf("test name: '%s', expect: %+v, actual: %+v\n", name, tc.expect, actual)
		}
	}
}

func createJobFromCondition(conditionTypes []batchv1.JobConditionType) *batchv1.Job {
	var conditions []batchv1.JobCondition
	for _, v := range conditionTypes {
		conditions = append(conditions, batchv1.JobCondition{Type: v})
	}
	return &batchv1.Job{
		Status: batchv1.JobStatus{
			Conditions: conditions,
		},
	}
}
