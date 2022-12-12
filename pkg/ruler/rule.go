package ruler

type Rule struct {
	ApiPath   string
	MaxTokens int64
	Rate      int64 //tokens per sec
}
