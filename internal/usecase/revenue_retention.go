package usecase

type RevenueRetentionInput struct {
	StartingMRR    float64
	ExpansionMRR   float64
	ContractionMRR float64
	ChurnedMRR     float64
}

type RevenueRetentionMetrics struct {
	NRR float64
	GRR float64
}

type RevenueRetentionSummary struct {
	StartingMRR    float64 `json:"starting_mrr"`
	ExpansionMRR   float64 `json:"expansion_mrr"`
	ContractionMRR float64 `json:"contraction_mrr"`
	ChurnedMRR     float64 `json:"churned_mrr"`
	NRR            float64 `json:"nrr"`
	GRR            float64 `json:"grr"`
}

func CalculateRevenueRetentionMetrics(input RevenueRetentionInput) RevenueRetentionMetrics {
	if input.StartingMRR <= 0 {
		return RevenueRetentionMetrics{}
	}

	nrr := (input.StartingMRR + input.ExpansionMRR - input.ContractionMRR - input.ChurnedMRR) / input.StartingMRR * 100
	grr := (input.StartingMRR - input.ContractionMRR - input.ChurnedMRR) / input.StartingMRR * 100

	if nrr < 0 {
		nrr = 0
	}
	if grr < 0 {
		grr = 0
	}

	return RevenueRetentionMetrics{
		NRR: nrr,
		GRR: grr,
	}
}

func BuildRevenueRetentionSummary(input RevenueRetentionInput) RevenueRetentionSummary {
	metrics := CalculateRevenueRetentionMetrics(input)
	return RevenueRetentionSummary{
		StartingMRR:    input.StartingMRR,
		ExpansionMRR:   input.ExpansionMRR,
		ContractionMRR: input.ContractionMRR,
		ChurnedMRR:     input.ChurnedMRR,
		NRR:            metrics.NRR,
		GRR:            metrics.GRR,
	}
}
