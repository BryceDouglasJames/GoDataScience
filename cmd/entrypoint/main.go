package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// PageData holds the data to render the HTML page
type PageData struct {
	PlotImage string
}
type ServerConfig struct {
	Server  fiber.App
	Address string
	//kafka...
	Logger *zap.Logger
}

func main() {
	config := &ServerConfig{
		Address: "127.0.0.1:8000",
	}

	err_signal, err := RunService(config)
	if err != nil {
		log.Fatalf("Something went very wrong while starting the server... %s", err)
	}

	if err := <-err_signal; err != nil {
		log.Fatalf("Internal server error while running: %s", err)
	}

}

func RunService(config *ServerConfig) (chan error, error) {
	err_chan := make(chan error, 1)

	//create context for service
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	app, err := StartServer(*config)
	if err != nil {
		defer func() {
			<-ctx.Done()
			config.Logger.Info("Error when trying to start server..." + err.Error())
			_ = config.Logger.Sync()
			stop()
			close(err_chan)
		}()
		return err_chan, err
	}

	logger, err := zap.NewProduction()
	if err != nil {
		defer func() {
			<-ctx.Done()
			config.Logger.Info("Error when trying to start logger..." + err.Error())
			_ = config.Logger.Sync()
			stop()
			close(err_chan)
		}()
		return err_chan, err
	}

	//init attributes and assign config values
	ConfigLogger(logger, &app)
	config.Server = app
	config.Logger = logger

	//routine to shut down server
	go func() {
		<-ctx.Done()
		config.Logger.Info("Shutdown signal received")
		defer func() {
			_ = config.Logger.Sync()
			stop()
			close(err_chan)
		}()

		if err := config.Server.Shutdown(); err != nil {
			err_chan <- err
		}
		config.Logger.Info("Shutdown completed")
	}()

	go func(config ServerConfig) {
		config.Logger.Info("Listening and serving", zap.String("address", config.Address))
		if err := config.Server.Listen(config.Address); err != nil {
			err_chan <- err
		}
	}(*config)

	return err_chan, nil
}

func ConfigLogger(logger *zap.Logger, app *fiber.App) error {
	app.Get("/logger", func(c *fiber.Ctx) error {
		fields := []zap.Field{
			zap.Namespace("context"),
			zap.String("pid", strconv.Itoa(os.Getpid())),
			zap.Time("time", time.Now()),
		}
		logger.Info("GET",
			zap.String("url", c.BaseURL()),
		)
		logger.With(fields...)
		c.WriteString("Log endpoint hit...")
		logger.Sugar()
		return nil
	})

	app.Post("/logger", func(c *fiber.Ctx) error {
		fields := []zap.Field{
			zap.Namespace("context"),
			zap.String("pid", strconv.Itoa(os.Getpid())),
			zap.Time("time", time.Now()),
		}
		logger.Info("POST",
			zap.String("url", c.BaseURL()),
		)
		logger.With(fields...)
		logger.Sugar()
		return nil
	})

	return nil
}

func StartServer(config ServerConfig) (fiber.App, error) {

	app := fiber.New()

	// Serve the plot image file
	app.Static("/images", "images")

	// Serve the templates
	app.Static("/templates", "templates")

	//DEBUG ONLY
	//fmt.Println(app.Stack())

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

	return *app, nil
}
