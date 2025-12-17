package git

import "testing"

func TestParseDiffStats(t *testing.T) {
	tests := map[string]Stats{
		"1 file changed, 3 insertions(+), 1 deletion(-)":        {1, 3, 1},
		"2 files changed, 10 insertions(+)":                     {2, 10, 0},
		"5 files changed, 7 deletions(-)":                       {5, 0, 7},
		"3 files changed, 12 insertions(+), 8 deletions(-)":     {3, 12, 8},
		"1 file changed, 0 insertions(+), 5 deletions(-)":       {1, 0, 5},
		"12 files changed, 240 insertions(+), 198 deletions(-)": {12, 240, 198},
	}

	for output, expectedStats := range tests {
		t.Run(output, func(t *testing.T) {
			stats, err := parseDiffStats(output)
			if err != nil {
				t.Fatal(err)
			}
			if *stats != expectedStats {
				t.Errorf("expected %v, got %v", expectedStats, stats)
			}
		})
	}
}
