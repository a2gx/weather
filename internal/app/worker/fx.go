package worker

import "errors"

type App struct{}

func Build() (*App, error) {
	return &App{}, errors.New("not implemented")
}
