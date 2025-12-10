package imageops

import (
	libvips "imagine/internal/imageops/vips"
)

// NormalizeToSRGB converts the image to sRGB color space.
// It handles existing ICC profiles or falls back to standard colorspace conversion.
// This is critical for CMYK images or images with specific profiles to render correctly in browsers.
func NormalizeToSRGB(img *libvips.Image) error {
	// If image is already sRGB, we might still want to run IccTransform if there's an embedded profile
	// that needs to be "baked in" or converted to the standard sRGB profile.
	// However, usually checking interpretation is a good first step optimization, 
	// but strictly speaking, "sRGB" interpretation doesn't guarantee the data matches the standard sRGB profile 
	// if the embedded profile says otherwise.
	//
	// Safest approach: 
	// 1. If ICC profile exists -> Transform to sRGB.
	// 2. If no ICC profile -> Force interpretation to sRGB (converting values if it was CMYK/LAB etc).

	if img.HasICCProfile() {
		// "srgb" is a magic string in libvips (provided lcms2 is linked) that refers to a standard sRGB profile.
		// Embedded: true tells it to use the profile found in the image as the input profile.
		err := img.IccTransform("srgb", &libvips.IccTransformOptions{
			Embedded: true,
			Intent:   libvips.IntentPerceptual, // Perceptual is usually best for photographic content
		})
		if err != nil {
			return err
		}
		// vips_icc_transform attaches the new profile. We might want to strip it to save space
		// since browsers assume sRGB anyway, but keeping it is "more correct".
		// For web use, stripping it is common optimization if we are sure it's sRGB.
		// img.RemoveICCProfile() 
		return nil
	}

	// Fallback for images without profiles (e.g. untagged CMYK)
	if img.Interpretation() != libvips.InterpretationSrgb {
		return img.Colourspace(libvips.InterpretationSrgb, nil)
	}

	return nil
}
