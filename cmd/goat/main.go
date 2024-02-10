package main

// Import ...
import (
	"flag"
	"log"
	"os"

	"github.com/blampe/goat"
)

// Function init ...
func init() {
	log.SetFlags(/*log.Ldate |*/ log.Ltime | log.Lshortfile)
}

func main() {
	var (
		inputFilename,
		outputFilename string
		svgConfig goat.SVGConfig
	)

	goat.InitFromFlags(&svgConfig)

	flag.StringVar(&inputFilename, "i", "", "Input filename (default stdin)")
	flag.StringVar(&outputFilename, "o", "", "Output filename (default stdout for SVG)")
	
	flag.Parse()

	input := os.Stdin
	if inputFilename != "" {
		if _, err := os.Stat(inputFilename); os.IsNotExist(err) {
			log.Fatalf("input file not found: %s", inputFilename)
		}
		var err error
		input, err = os.Open(inputFilename)
		defer input.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	output := os.Stdout
	if outputFilename != "" {
		var err error
		output, err = os.Create(outputFilename)
		defer output.Close()          // XX  Move outside 'if' -- close os.Stdout as well?
		if err != nil {
			log.Fatal(err)
		}
	}
	goat.BuildAndWriteSVG(input, output, svgConfig)
}
