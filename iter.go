package goat

// XX  Simplify to closure functions rather than channels?
//      X Callers currently use 'range' -- rewrite would be required.
type canvasIterator func(width int, height int) chan Index

func upDownMinor(width int, height int) chan Index {
	c := make(chan Index)
	go func() {
		for w := 0; w < width; w++ {
			for h := 0; h < height; h++ {
				c <- Index{w, h}
			}
		}
		close(c)
	}()
	return c
}

func leftRightMinor(width int, height int) chan Index {
	c := make(chan Index)
	go func() {
		for h := 0; h < height; h++ {
			for w := 0; w < width; w++ {
				c <- Index{w, h}
			}
		}
		close(c)
	}()
	return c
}

func diagDown(width int, height int) chan Index {
	c := make(chan Index)
	go func() {
		minSum := -height + 1
		maxSum := width

		for sum := minSum; sum <= maxSum; sum++ {
			for w := 0; w < width; w++ {
				for h := 0; h < height; h++ {
					if w-h == sum {
						c <- Index{w, h}
					}
				}
			}
		}
		close(c)
	}()
	return c
}

func diagUp(width int, height int) chan Index {
	c := make(chan Index)
	go func() {
		maxSum := width + height - 2

		for sum := 0; sum <= maxSum; sum++ {
			for w := 0; w < width; w++ {
				for h := 0; h < height; h++ {
					if h+w == sum {
						c <- Index{w, h}
					}
				}
			}
		}
		close(c)
	}()
	return c
}
