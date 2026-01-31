package target

import(
	"math/rand"
	"time"
)

type Provider struct {
	words []string 
	rng   *rand.Rand
}


func New(words []string) *Provider{
	return &Provider{
		words: words,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}


func (p *Provider) Random() string {
	if len(p.words) == 0 {
		return "default"
	}
	return p.words[p.rng.Intn(len(p.words))]
}