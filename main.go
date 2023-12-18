package main

import (
	"challenge/config"
	"challenge/handlers"
	"challenge/internal/tweet"
	"challenge/storage"
	"context"
	"github.com/go-chi/chi"
	"github.com/magiconair/properties"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	app, err := Build()
	if err != nil {
		panic(err)
	}

	err = Routes(app)
	if err != nil {
		panic(err)
	}

}

func buildConfig() config.Configuration {
	cfgTest, err := buildConfigFromLocalFile()
	if err != nil {
		panic(err)
	}
	return cfgTest
}

func buildConfigFromLocalFile() (*properties.Properties, error) {
	propsFromFile := make(map[string]string)
	path, _ := filepath.Abs("./config/local.yaml")
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &propsFromFile)
	if err != nil {
		return nil, err
	}

	return properties.LoadMap(propsFromFile), nil
}

type Engine struct {
	Configs config.Configuration
}

func Build() (*Engine, error) {
	Config := buildConfig()
	return &Engine{
		Configs: Config,
	}, nil
}

func Routes(app *Engine) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	storage := storage.New(ctx, app.Configs)
	defer storage.Close(ctx)

	tweetService := tweet.New(storage, app.Configs.GetInt("limit", 0))
	handlerTweet := handlers.NewTweetHandler(tweetService)
	r := chi.NewRouter()
	r.Post("/v1/tweet", Handler(handlerTweet.Create))
	r.Post("/v1/follow/{userID}", Handler(handlerTweet.Follow))
	r.Get("/v1/timeline", Handler(handlerTweet.ViewTimeline))

	port := os.Getenv("PORT")
	if port == "" {
		port = app.Configs.GetString("port", "")
	}

	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		return err
	}

	return nil
}

func Handler(f func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(r http.ResponseWriter, w *http.Request) {
		f(r, w)
	}
}
