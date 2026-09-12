package main

import (
	"os"
	"strings"
	"vnti/apps/api"
	"vnti/apps/workers"
	workerHandlers "vnti/apps/workers/handlers"
	"vnti/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "main",
	Short: "Run server or worker",
	Args:  cobra.ArbitraryArgs,
}

var apiServerCmd = &cobra.Command{
	Use:   "api",
	Short: "Run API server",
	Run:   runApi,
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run worker service",
	Run:   runWorker,
}

func init() {
	rootCmd.AddCommand(apiServerCmd)
	rootCmd.AddCommand(workerCmd)
}

// Preparation before start the service
//
//   - Create `media` directory
//   - Create `media` children dir like: `avatar`, etc
func preparationCheck() {
	config := internal.NewConfig()
	if err := config.Check(); err != nil {
		panic(err)
	}

	_, err := os.Stat(config.MEDIA_PATH)
	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		if err = os.Mkdir(config.MEDIA_PATH, 0755); err != nil {
			panic(err)
		}
	}

	for _, dirname := range []string{
		"avatar", // Media directory
	} {
		dir_path := config.MEDIA_PATH + string(os.PathSeparator) + dirname
		// Try to create dir, ignore errors
		_ = os.Mkdir(dir_path, 0755)
	}
}

func runApi(cmd *cobra.Command, args []string) {
	config := internal.NewConfig()
	api := api.NewApi(config)
	log.Fatal().AnErr("API", api.Listen("0.0.0.0:8080", fiber.ListenConfig{
		EnablePrefork: !config.Debug,
	}))
}

func runWorker(cmd *cobra.Command, args []string) {
	config := internal.NewConfig()
	log.Printf("%+v\n", config)
	worker := workers.New(config)
	log.Fatal().AnErr("Worker", workerHandlers.Run(worker))
}

func main() {
	preparationCheck()
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
