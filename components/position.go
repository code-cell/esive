package components

func (c *Position) Distance(other *Position) float32 {
	return Distance(c.X, c.Y, other.X, other.Y)
}
