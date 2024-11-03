package main

import (
	"context"
	"flag"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/tiff"

	"image/color"

	"github.com/google/subcommands"
)

func load(filename string) (image.Image, error) {
	f, oerr := os.Open(filename)
	if oerr != nil {
		return nil, oerr
	}
	defer f.Close()

	log.Printf("Processing %s\n", filename)
	rawExtensions := []string{".cr2", ".nef", ".raf", ".arw", ".dng"}
	for _, ext := range rawExtensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			// Decode using golibraw
			img, err := ImportRaw(filename)
			if err != nil {
				return nil, err
			}

			return img, nil
		}
	}

	p, _, derr := image.Decode(f)
	if derr != nil {
		return nil, derr
	}

	return p, nil
}

func samplePalette(picture image.Image, sampleArea image.Rectangle) *Palette {
	var palette Palette
	for x := sampleArea.Min.X; x < sampleArea.Max.X; x++ {
		for y := sampleArea.Min.Y; y < sampleArea.Max.Y; y++ {
			palette.Add(color.RGBA64Model.Convert(picture.At(x, y)).(color.RGBA64))
		}
	}
	return &palette
}

type convertCmd struct {
	inputDir              string
	outputDir             string
	sampleFraction        float64
	lowlights, highlights float64
	scurve                bool
	outputFormat          string
}

func (*convertCmd) Name() string {
	return "convert"
}

func (*convertCmd) Synopsis() string {
	return "Invert input image, normalize colors and output a file"
}

func (*convertCmd) Usage() string {
	return ""
}

func (c *convertCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.inputDir, "input", "", "Input directory containing TIFF files")
	f.StringVar(&c.outputDir, "output", "", "Output directory for converted JPEG files")
	f.Float64Var(&c.sampleFraction,
		"sample_fraction", 0.8,
		"Sample palette from a fraction crop of the center, 0 < fraction < 1 (default 0.8)")
	f.Float64Var(&c.lowlights,
		"lowlights", 0.01,
		"Shadows start here, lower values save more shadows")
	f.Float64Var(&c.highlights,
		"highlights", 0.99,
		"Highlights start here, lower values saves more highlights")
	f.BoolVar(&c.scurve,
		"s-curve", false,
		"Use sigmoid funciton instead of linear mapping")
	f.StringVar(&c.outputFormat, "output-format", "tiff", "Output file format TIFF default, available options TIFF | JPEG ")
}

func (c *convertCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if c.inputDir == "" || c.outputDir == "" {
		log.Println("Please specify both input and output directories.")
		return subcommands.ExitUsageError
	}

	if strings.ToLower(c.outputFormat) != "tiff" && strings.ToLower(c.outputFormat) != "jpeg" {
		log.Println("Invalid format. Please use TIFF or JPEG.")
		return subcommands.ExitUsageError
	}

	if err := os.MkdirAll(c.outputDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	err := filepath.Walk(c.inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			outputFile := filepath.Join(c.outputDir, strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))+"."+strings.ToLower(c.outputFormat))

			picture, load_err := load(path)

			if load_err != nil {
				log.Fatalf("Could not load input file `%s`: %v",
					path,
					load_err)
			}

			sampleArea := sampleBounds(c.sampleFraction, picture)
			palette := samplePalette(picture, sampleArea)

			t := Transformation{
				Range{Low: palette.Red.Percentile(c.lowlights), High: palette.Red.Percentile(c.highlights)},
				Range{Low: palette.Green.Percentile(c.lowlights), High: palette.Green.Percentile(c.highlights)},
				Range{Low: palette.Blue.Percentile(c.lowlights), High: palette.Blue.Percentile(c.highlights)},
				c.lowlights - c.highlights,
			}

			mapping := t.Linear()
			if c.scurve {
				mapping = t.Sigmoid()
			}

			copy := mapping.Apply(picture)

			of, ferr := os.Create(outputFile)
			if ferr != nil {
				log.Fatal(ferr)
			}
			defer of.Close()

			if strings.ToLower(c.outputFormat) == "jpeg" {
				jpeg.Encode(of, copy, &jpeg.Options{Quality: 95})
			} else {
				tiff.Encode(of, copy, &tiff.Options{Compression: tiff.Deflate, Predictor: true})
			}

			log.Printf("Successfully processed and saved: %s\n", outputFile)

		}
		return nil
	})

	if err != nil {
		log.Printf("Error processing directory: %v\n", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess

}

// sampleBounds gets a bounding box for a center fraction of the image, based
// on parameter fraction
func sampleBounds(fraction float64, picture image.Image) image.Rectangle {
	bounds := picture.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	border := (1 - fraction) / 2

	sampleArea := image.Rectangle{
		image.Point{bounds.Min.X + int(float64(width)*border),
			bounds.Min.Y + int(float64(height)*border)},
		image.Point{bounds.Max.X - int(float64(width)*border),
			bounds.Max.Y - int(float64(height)*border)}}

	return sampleArea
}

func main() {
	subcommands.Register(&convertCmd{}, "")

	flag.Parse()

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
