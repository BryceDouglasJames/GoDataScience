package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	"github.com/gofiber/fiber/v2"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// PageData holds the data to render the HTML page
type PageData struct {
	PlotImage string
}

func main() {
	// Create the Fiber app
	app := fiber.New()

	// Serve the plot image file
	app.Static("/images", "images")

	// Serve the templates
	app.Static("/templates", "templates")

	fmt.Println(app.Stack())

	// Serve the HTML page
	app.Get("/", func(c *fiber.Ctx) error {
		// Create the plot
		p := plot.New()
		p.Title.Text = "Plotting in Go with gonum/plot"
		p.X.Label.Text = "X"
		p.Y.Label.Text = "Y"

		pts := plotter.XYs{{X: 1, Y: 1}, {X: 2, Y: 2}, {X: 3, Y: 3}, {X: 4, Y: 4}}
		err := plotutil.AddLinePoints(p, "Line 1", pts)
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("%s", err))
		}

		// Save the plot as an image file
		if err := p.Save(4*vg.Inch, 4*vg.Inch, "images/plot.png"); err != nil {

			return c.Status(500).SendString(fmt.Sprintf("%s", err))
		}

		// Render the HTML template
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			return c.Status(500).SendString(fmt.Sprint(err))
		}

		data := PageData{PlotImage: "images/plot.png"}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, data)
		if err != nil {
			return c.Status(500).SendString(fmt.Sprint(err))
		}

		return c.Type("html").Send(buf.Bytes())
	})

	fmt.Println("Listening on http://localhost:8000")
	if err := app.Listen(":8000"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
