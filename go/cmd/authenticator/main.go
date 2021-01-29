package main

import (
	"texas_real_foods/pkg/utils"
	auth "texas_real_foods/pkg/authenticator"
)

var (
    // create map to house environment variables
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100:5432",
        },
    )
)

func main() {

	service := auth.New(cfg.Get("postgres_url"), "0.0.0.0", 10101)
	service.Run()
}