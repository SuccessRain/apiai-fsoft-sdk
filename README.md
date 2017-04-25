# Simple Go library for api.ai

## Install
$ go get -u github.com/SuccessRain/apiai-fsoft-sdk

import (
	"fmt"
	"github.com/meinside/api.ai-go"
)

api train -t intent -i training-api.csv -token <your-token>
api test -t intent -i test-api.csv -token <your-token>
