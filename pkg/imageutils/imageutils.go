package imageutils

// RGB represents an RGB color with 8-bit channels.
type RGB struct {
	R, G, B int
}

// TileImage holds information about a tile image, including its path, average color, and usage status.
type TileImage struct {
	Path       string
	RGB        RGB
	UsageCount int
}

// squaredEuclideanDistance calculates the squared distance between two RGB colors.
// We use squared distance as it's faster than taking the square root and is sufficient for comparison.
func SquaredEuclideanDistance(c1, c2 RGB) int {
	dr := c1.R - c2.R
	dg := c1.G - c2.G
	db := c1.B - c2.B
	return dr*dr + dg*dg + db*db
}
