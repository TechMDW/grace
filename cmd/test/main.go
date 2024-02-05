package main

import (
	"context"

	grace "github.com/TechMDW/grace"
)

func main() {
	grace.AddX(5)
	grace.Add()

	grace.Done()
	grace.DoneX(5)
	grace.Add()

	grace.Grace(context.Background())
}
