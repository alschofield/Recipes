package search

import "testing"

func TestCalibratedOverallWithJudge_BaselineGuardrail(t *testing.T) {
	tests := []struct {
		name                 string
		deterministicOverall float64
		judgeOverall         float64
		judgeConfidence      float64
		minConfidence        float64
		want                 float64
	}{
		{
			name:                 "high confidence blends with deterministic baseline",
			deterministicOverall: 0.40,
			judgeOverall:         0.80,
			judgeConfidence:      0.90,
			minConfidence:        0.65,
			want:                 0.50,
		},
		{
			name:                 "equal threshold still blends",
			deterministicOverall: 0.52,
			judgeOverall:         0.36,
			judgeConfidence:      0.65,
			minConfidence:        0.65,
			want:                 0.48,
		},
		{
			name:                 "low confidence keeps deterministic baseline",
			deterministicOverall: 0.61,
			judgeOverall:         0.95,
			judgeConfidence:      0.50,
			minConfidence:        0.65,
			want:                 0.61,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calibratedOverallWithJudge(tc.deterministicOverall, tc.judgeOverall, tc.judgeConfidence, tc.minConfidence)
			if !roughlyEqual(got, tc.want) {
				t.Fatalf("got %.4f want %.4f", got, tc.want)
			}
		})
	}
}

func TestCalibratedOverallWithJudge_ManualSpotChecks(t *testing.T) {
	spotChecks := []struct {
		name                 string
		deterministicOverall float64
		judgeOverall         float64
		judgeConfidence      float64
		minConfidence        float64
		wantMin              float64
		wantMax              float64
	}{
		{
			name:                 "judge nudges weak deterministic score upward",
			deterministicOverall: 0.35,
			judgeOverall:         0.92,
			judgeConfidence:      0.88,
			minConfidence:        0.65,
			wantMin:              0.49,
			wantMax:              0.50,
		},
		{
			name:                 "judge cannot fully override deterministic baseline",
			deterministicOverall: 0.82,
			judgeOverall:         0.20,
			judgeConfidence:      0.90,
			minConfidence:        0.65,
			wantMin:              0.66,
			wantMax:              0.67,
		},
	}

	for _, sc := range spotChecks {
		t.Run(sc.name, func(t *testing.T) {
			got := calibratedOverallWithJudge(sc.deterministicOverall, sc.judgeOverall, sc.judgeConfidence, sc.minConfidence)
			if got < sc.wantMin || got > sc.wantMax {
				t.Fatalf("got %.4f outside expected range [%.4f, %.4f]", got, sc.wantMin, sc.wantMax)
			}
		})
	}
}

func roughlyEqual(a, b float64) bool {
	const eps = 0.0001
	if a > b {
		return a-b <= eps
	}
	return b-a <= eps
}
