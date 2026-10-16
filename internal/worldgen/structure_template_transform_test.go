package worldgen

import "testing"

func TestTemplatePositionTransformsAgainstVanillaJAR(t *testing.T) {
	data, err := loadTemplateTransforms()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range data.Positions {
		got, err := TransformTemplatePosition(FeaturePosition{row[0], row[1], row[2]}, FeaturePosition{row[3], row[4], row[5]}, TemplateMirror(row[6]), TemplateRotation(row[7]))
		want := FeaturePosition{row[8], row[9], row[10]}
		if err != nil || got != want {
			t.Fatalf("%v: transform got %v (%v), want %v", row, got, err, want)
		}
	}
	t.Logf("matched %d original pivot/mirror/rotation vectors, including Java overflow", len(data.Positions))
}
