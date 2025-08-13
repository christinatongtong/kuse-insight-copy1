package outputs

import (
	"context"
	"fmt"
	"testing"

	"github.com/mixpanel/mixpanel-go"
)

const (
	MIXPANEL_TOKEN = "6635b781162485d2feb3ce28729b08af"
)

func TestMixpanel(t *testing.T) {
	mp := mixpanel.NewApiClient(MIXPANEL_TOKEN,
		mixpanel.ServiceAccount(0, "kuse-analyst-staging.9ced9d.mp-service-account", "tEhBf4H5dbHmysfcDGWkuTVnlZ3XCtXa"),
	)
	if err := mp.PeopleSet(context.Background(), []*mixpanel.PeopleProperties{{
		DistinctID: "310",
		Properties: map[string]any{
			"predict_is_student":       true,
			"predict_school":           "ECJTU",
			"predict_major":            "AI",
			"predict_degree_level":     "Master",
			"predict_occupation":       "Software Engineer",
			"predict_industry":         "Technology",
			"predict_primary_language": "Madarin",
			"predict_gender":           "male",
		},
		UseRequestIp: false,
	}}); err != nil {
		fmt.Printf("PeopleSetError err: %v\n", err)
	}
}
