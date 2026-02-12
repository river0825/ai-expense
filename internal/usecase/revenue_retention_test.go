package usecase

import "testing"

func TestCalculateRevenueRetentionMetrics(t *testing.T) {
	tests := []struct {
		name  string
		input RevenueRetentionInput
		want  RevenueRetentionMetrics
	}{
		{
			name: "baseline with expansion and contraction",
			input: RevenueRetentionInput{
				StartingMRR:    1000,
				ExpansionMRR:   200,
				ContractionMRR: 100,
				ChurnedMRR:     50,
			},
			want: RevenueRetentionMetrics{
				NRR: 105,
				GRR: 85,
			},
		},
		{
			name: "nrr can exceed one hundred",
			input: RevenueRetentionInput{
				StartingMRR:    1000,
				ExpansionMRR:   300,
				ContractionMRR: 50,
				ChurnedMRR:     50,
			},
			want: RevenueRetentionMetrics{
				NRR: 120,
				GRR: 90,
			},
		},
		{
			name: "refunds are reflected as contraction mrr",
			input: RevenueRetentionInput{
				StartingMRR:    1000,
				ExpansionMRR:   100,
				ContractionMRR: 150,
				ChurnedMRR:     0,
			},
			want: RevenueRetentionMetrics{
				NRR: 95,
				GRR: 85,
			},
		},
		{
			name: "chargebacks are reflected as churned mrr",
			input: RevenueRetentionInput{
				StartingMRR:    800,
				ExpansionMRR:   0,
				ContractionMRR: 0,
				ChurnedMRR:     200,
			},
			want: RevenueRetentionMetrics{
				NRR: 75,
				GRR: 75,
			},
		},
		{
			name: "plan upgrades and downgrades net into expansion and contraction",
			input: RevenueRetentionInput{
				StartingMRR:    500,
				ExpansionMRR:   120,
				ContractionMRR: 70,
				ChurnedMRR:     30,
			},
			want: RevenueRetentionMetrics{
				NRR: 104,
				GRR: 80,
			},
		},
		{
			name: "cancel at period end is excluded until it actually churns",
			input: RevenueRetentionInput{
				StartingMRR:    700,
				ExpansionMRR:   0,
				ContractionMRR: 0,
				ChurnedMRR:     0,
			},
			want: RevenueRetentionMetrics{
				NRR: 100,
				GRR: 100,
			},
		},
		{
			name: "late arriving events are counted when they appear in the period rollup",
			input: RevenueRetentionInput{
				StartingMRR:    200,
				ExpansionMRR:   30,
				ContractionMRR: 10,
				ChurnedMRR:     10,
			},
			want: RevenueRetentionMetrics{
				NRR: 105,
				GRR: 90,
			},
		},
		{
			name: "empty cohorts return zeros",
			input: RevenueRetentionInput{
				StartingMRR:    0,
				ExpansionMRR:   100,
				ContractionMRR: 10,
				ChurnedMRR:     10,
			},
			want: RevenueRetentionMetrics{},
		},
		{
			name: "negative result floors at zero",
			input: RevenueRetentionInput{
				StartingMRR:    100,
				ExpansionMRR:   0,
				ContractionMRR: 60,
				ChurnedMRR:     60,
			},
			want: RevenueRetentionMetrics{
				NRR: 0,
				GRR: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateRevenueRetentionMetrics(tt.input)
			if got.NRR != tt.want.NRR {
				t.Fatalf("NRR mismatch: got %v want %v", got.NRR, tt.want.NRR)
			}
			if got.GRR != tt.want.GRR {
				t.Fatalf("GRR mismatch: got %v want %v", got.GRR, tt.want.GRR)
			}
		})
	}
}

func TestBuildRevenueRetentionSummary(t *testing.T) {
	input := RevenueRetentionInput{
		StartingMRR:    1000,
		ExpansionMRR:   200,
		ContractionMRR: 100,
		ChurnedMRR:     50,
	}

	summary := BuildRevenueRetentionSummary(input)

	if summary.StartingMRR != 1000 {
		t.Fatalf("StartingMRR mismatch: got %v", summary.StartingMRR)
	}
	if summary.ExpansionMRR != 200 {
		t.Fatalf("ExpansionMRR mismatch: got %v", summary.ExpansionMRR)
	}
	if summary.ContractionMRR != 100 {
		t.Fatalf("ContractionMRR mismatch: got %v", summary.ContractionMRR)
	}
	if summary.ChurnedMRR != 50 {
		t.Fatalf("ChurnedMRR mismatch: got %v", summary.ChurnedMRR)
	}
	if summary.NRR != 105 {
		t.Fatalf("NRR mismatch: got %v", summary.NRR)
	}
	if summary.GRR != 85 {
		t.Fatalf("GRR mismatch: got %v", summary.GRR)
	}
}
