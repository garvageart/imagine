package imageops

import (
	"imagine/internal/config"
	"imagine/internal/dto"
	"imagine/internal/entities"
	"imagine/internal/images"
	libvips "imagine/internal/imageops/vips"
	"imagine/internal/transform"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Initialize libvips for tests
	WarmupAllOps(config.LibvipsConfig{
		Concurrency: 1,
	})
	
	code := m.Run()
	
	libvips.Shutdown()
	os.Exit(code)
}

func TestGenerateTransform_ModernSamples(t *testing.T) {
	// Path to samples relative to this test file
	samplesDir := "../../resources/test/samples"
	
	// Ensure samples exist
	if _, err := os.Stat(samplesDir); os.IsNotExist(err) {
		t.Skipf("Samples directory not found at %s, skipping test", samplesDir)
	}

	files, err := os.ReadDir(samplesDir)
	if err != nil {
		t.Fatalf("Failed to read samples directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".cr2" || ext == ".cr3" || ext == ".arw" || ext == ".dng" || ext == ".nef" || ext == ".orf" {
				t.Skip("Skipping RAW file for speed")
			}

			path := filepath.Join(samplesDir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", path, err)
			}

			// Mock entity
			imgEnt := entities.Image{
				ImageMetadata: &dto.ImageMetadata{
					FileType: strings.TrimPrefix(filepath.Ext(file.Name()), "."),
					Checksum: "mock-checksum-" + file.Name(),
				},
			}

			// Test cases for each image
			testCases := []struct {
				name   string
				params *transform.TransformParams
			}{
				{
					name: "Resize 200x200 WebP",
					params: &transform.TransformParams{
						Width:   200,
						Height:  200,
						Format:  "webp",
						Quality: 80,
					},
				},
				{
					name: "Resize Height 300 JPG",
					params: &transform.TransformParams{
						Height:  300,
						Format:  "jpg",
						Quality: 75,
					},
				},
				{
					name: "Resize Width 150 PNG",
					params: &transform.TransformParams{
						Width:   150,
						Format:  "png",
						Quality: 90,
					},
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					result, err := GenerateTransform(tc.params, imgEnt, data)
					if err != nil {
						t.Fatalf("GenerateTransform failed for %s: %v. Hint: On Windows, run .\\scripts\\setup-libvips.ps1 to ensure all delegates (RAW, etc.) are installed.", file.Name(), err)
					}

					if result == nil {
						t.Fatal("Result is nil")
					}

					if len(result.ImageData) == 0 {
						t.Fatal("Result ImageData is empty")
					}

					expectedFormat := tc.params.Format
					// Special case: jpg/jpeg
					if expectedFormat == "jpg" || expectedFormat == "jpeg" {
						if result.Ext != "jpg" && result.Ext != "jpeg" {
							t.Errorf("Expected ext %s, got %s", expectedFormat, result.Ext)
						}
					} else if result.Ext != expectedFormat {
						t.Errorf("Expected ext %s, got %s", expectedFormat, result.Ext)
					}
					
					// Verify dimensions
					resImg, err := libvips.NewImageFromBuffer(result.ImageData, libvips.DefaultLoadOptions())
					if err != nil {
						t.Fatalf("Failed to decode result image for verification: %v", err)
					}
					defer resImg.Close()
					
					width := int64(resImg.Width())
					height := int64(resImg.Height())

					if tc.params.Width > 0 && tc.params.Height > 0 {
						if width > tc.params.Width || height > tc.params.Height {
							t.Errorf("Result image %dx%d exceeds bounds %dx%d", width, height, tc.params.Width, tc.params.Height)
						}
						// Check if at least one dimension matches (approx)
						if width != tc.params.Width && height != tc.params.Height {
							if diff(width, tc.params.Width) > 1 && diff(height, tc.params.Height) > 1 {
								t.Errorf("Result image %dx%d does not match constraint %dx%d (neither dim matches)", width, height, tc.params.Width, tc.params.Height)
							}
						}
					} else if tc.params.Width > 0 {
						if diff(width, tc.params.Width) > 1 {
							t.Errorf("Result width %d does not match requested %d", width, tc.params.Width)
						}
					} else if tc.params.Height > 0 {
						if diff(height, tc.params.Height) > 1 {
							t.Errorf("Result height %d does not match requested %d", height, tc.params.Height)
						}
					}

					t.Logf("Success: %s -> %d bytes %s (%dx%d)", tc.name, len(result.ImageData), result.Ext, width, height)
				})
			}
		})
	}
}

func TestPermanentTransforms(t *testing.T) {
	// Path to samples relative to this test file
	samplesDir := "../../resources/test/samples"

	if _, err := os.Stat(samplesDir); os.IsNotExist(err) {
		t.Skipf("Samples directory not found at %s, skipping test", samplesDir)
	}

	files, err := os.ReadDir(samplesDir)
	if err != nil {
		t.Fatalf("Failed to read samples directory: %v", err)
	}

	permTransforms := images.GetAllPermanentTransforms()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".cr2" || ext == ".cr3" || ext == ".arw" || ext == ".dng" || ext == ".nef" || ext == ".orf" {
				t.Skip("Skipping RAW file for speed")
			}

			path := filepath.Join(samplesDir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", path, err)
			}

			imgEnt := entities.Image{
				ImageMetadata: &dto.ImageMetadata{
					FileType: strings.TrimPrefix(filepath.Ext(file.Name()), "."),
					Checksum: "mock-checksum-" + file.Name(),
				},
			}

			for name, params := range permTransforms {
				t.Run(string(name), func(t *testing.T) {
					p := params
					result, err := GenerateTransform(&p, imgEnt, data)
					if err != nil {
						t.Fatalf("GenerateTransform failed for %s with %s: %v. Hint: On Windows, run .\\scripts\\setup-libvips.ps1 to ensure all delegates (RAW, etc.) are installed.", file.Name(), name, err)
					}

					if result == nil {
						t.Fatal("Result is nil")
					}

					if len(result.ImageData) == 0 {
						t.Fatal("Result ImageData is empty")
					}
					
					// Verify result with libvips
					resImg, err := libvips.NewImageFromBuffer(result.ImageData, libvips.DefaultLoadOptions())
					if err != nil {
						t.Fatalf("Failed to decode transformed image: %v", err)
					}
					defer resImg.Close()

					width := int64(resImg.Width())
					height := int64(resImg.Height())

					if p.Width > 0 && p.Height > 0 {
						if width > p.Width || height > p.Height {
							t.Errorf("Transformed image %dx%d exceeds permanent transform %s bounds %dx%d", width, height, name, p.Width, p.Height)
						}
					}

					t.Logf("Success: %s [%s] -> %d bytes (%dx%d)", file.Name(), name, len(result.ImageData), width, height)
				})
			}
		})
	}
}

func diff(a, b int64) int64 {
	if a > b {
		return a - b
	}
	return b - a
}