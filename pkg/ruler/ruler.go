package ruler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bsm/redislock"
	"github.com/maingockien01/proxy/pkg/redis"
)

type Ruler struct {
	Filepath        string
	Rules           []*Rule
	OnFetch         func(*Ruler)
	FetchTime       time.Duration
	isFetchRepeated bool
	quitRepeater    chan (interface{})
}

var DEFAULT_RULE = &Rule{
	ApiPath:   "/",
	Rate:      5,
	MaxTokens: 5,
}

func NewRuler(filepath string, onFetch func(*Ruler), fetchTime time.Duration) *Ruler {
	ruler := &Ruler{
		Filepath:        filepath,
		OnFetch:         onFetch,
		isFetchRepeated: false,
		quitRepeater:    make(chan interface{}),
		FetchTime:       fetchTime,
		Rules:           make([]*Rule, 0),
	}

	return ruler
}

func (ruler *Ruler) FetchFile() {
	fmt.Println("Fetching rules from file")

	file, err := os.ReadFile(ruler.Filepath)

	if err != nil {
		//TODO: set up default rules here
		panic(err)
	}

	var rules []*Rule

	json.Unmarshal(file, &rules)

	if err != nil {
		//TODO: set up default rules here
		panic(err)
	}

	ruler.Rules = rules

	ruler.OnFetch(ruler)
}

func (ruler *Ruler) StopRepeater() {
	ruler.quitRepeater <- true
}

func (ruler *Ruler) FetchFileAndPushRedisInterval(client *redis.RedisClient) {
	if ruler.isFetchRepeated {
		fmt.Println("Ruler already repeats fetching!")
		return
	}
	ruler.isFetchRepeated = true
	ticker := time.NewTicker(ruler.FetchTime)

	go func() {
		for {
			select {
			case <-ticker.C:
				ruler.FetchFile()
				ruler.PushRedis(client)

			case <-ruler.quitRepeater:
				ticker.Stop()
				ruler.isFetchRepeated = false
				return
			}
		}
	}()

}

// api should have format of /path/to/api
func (ruler *Ruler) GetRule(apiPath string) *Rule {
	// search for full path
	// remove last /api
	// keep searching
	// rules should have rules for "/" path

	for {
		if len(apiPath) == 0 || !strings.HasPrefix(apiPath, "/") {
			fmt.Printf("Api Path %s does not have prefix //\n", apiPath)
			return DEFAULT_RULE
		}

		for _, rule := range ruler.Rules {
			if strings.Compare(rule.ApiPath, apiPath) == 0 {
				return rule
			}
		}

		if strings.Compare(apiPath, "/") == 0 {
			fmt.Println("Error: There should be default rule!")
			return DEFAULT_RULE
		}

		apiPath = removeLastPath(apiPath)
	}
}

func removeLastPath(apiPath string) string {
	lastSlash := strings.LastIndex(apiPath, "/")

	if lastSlash == 0 {
		return "/"
	}

	return apiPath[:lastSlash]
}

func (ruler *Ruler) FetchRedis(client *redis.RedisClient) {
	fmt.Println("Fetching rules from redis")
	lock, lockErr := client.Locker.Obtain(context.Background(), "rules-lock", 1*time.Minute, &redislock.Options{
		RetryStrategy: redislock.LinearBackoff(time.Second * 10),
	})
	defer lock.Release(context.Background())

	if lockErr != nil {
		panic(lockErr)
	}

	rulesJson, err := client.Get("rules")

	if redis.IsKeyNotExist(err) {
		ruler.PushRedis(client)
		return
	} else if err != nil {
		//TODO: set up default rules here
		panic(err)
	}

	var rules []*Rule

	json.Unmarshal([]byte(rulesJson), &rules)

	if err != nil {
		//TODO: set up default rules here
		panic(err)
	}

	ruler.Rules = rules

	ruler.OnFetch(ruler)
}

func (ruler *Ruler) FetchRedisInterval(client *redis.RedisClient) {
	if ruler.isFetchRepeated {
		fmt.Println("Ruler already repeats fetching!")
		return
	}
	ruler.isFetchRepeated = true
	ticker := time.NewTicker(ruler.FetchTime)

	go func() {
		for {
			select {
			case <-ticker.C:
				ruler.FetchRedis(client)

			case <-ruler.quitRepeater:
				ticker.Stop()
				ruler.isFetchRepeated = false
				return
			}
		}
	}()

}

func (ruler *Ruler) PushRedis(client *redis.RedisClient) {
	fmt.Println("Pushing rules to redis")
	rules, _ := json.MarshalIndent(ruler.Rules, "", "\t")
	client.Set("rules", string(rules))
}
