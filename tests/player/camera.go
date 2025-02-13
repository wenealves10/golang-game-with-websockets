package main

import "math"

type Camera struct {
	X, Y float64
}

func NewCamera(X, Y float64) *Camera {
	return &Camera{
		X: X,
		Y: Y,
	}
}

func (c *Camera) FollowTarget(targetX, targetY, screenWidth, screenHeight float64) {
	c.X = -targetX + screenWidth/2.0
	// c.Y = -targetY + screenHeight/2.0
}

// retrict camera to the world
func (c *Camera) Constrain(tilemapWidthPixels, tilemapHeightPixels, screenWidth, screenHeight float64) {
	c.X = math.Min(c.X, 0.0)
	c.X = math.Max(c.X, screenWidth-tilemapWidthPixels)
}
