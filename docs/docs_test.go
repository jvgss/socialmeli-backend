package docs

import "testing"

func TestDocsPackageLoads(t *testing.T) {
	// só referenciar algo do package já executa init()
	// e conta statements
	if SwaggerInfo == nil {
		t.Fatalf("SwaggerInfo should not be nil")
	}
}
