// Package recommendations suggests categories and desires (never concrete
// products or prices) for an occasion. Today the only implementation is
// rule-based; an LLM or Jev-backed provider can implement Provider later
// behind the AI_LIST_BUILDER flag.
package recommendations

import "context"

type Importance string

const (
	Essential   Importance = "ESSENTIAL"
	Recommended Importance = "RECOMMENDED"
	Optional    Importance = "OPTIONAL"
)

type Desire struct {
	Title      string     `json:"title"`
	Emoji      string     `json:"emoji"`
	Quantity   int        `json:"quantity"`
	Importance Importance `json:"importance"`
}

type CategorySuggestion struct {
	Name    string   `json:"name"`
	Emoji   string   `json:"emoji"`
	Reason  string   `json:"reason"`
	Desires []Desire `json:"desires"`
}

type Request struct {
	EventType string
	// Free text from the user (used by AI providers; ignored by rules).
	Prompt string
}

type Provider interface {
	SuggestCategories(ctx context.Context, req Request) ([]CategorySuggestion, error)
}

type RulesProvider struct{}

func (RulesProvider) SuggestCategories(_ context.Context, req Request) ([]CategorySuggestion, error) {
	if t, ok := templates[req.EventType]; ok {
		return t, nil
	}
	return templates["WISHLIST"], nil
}

func d(title, emoji string, qty int, imp Importance) Desire {
	return Desire{Title: title, Emoji: emoji, Quantity: qty, Importance: imp}
}

var home = []CategorySuggestion{
	{Name: "Cozinha", Emoji: "🍳", Reason: "É onde a casa nova mais precisa de itens no começo.", Desires: []Desire{
		d("Air fryer", "🍟", 1, Essential), d("Jogo de panelas", "🍲", 1, Essential), d("Faqueiro", "🍴", 1, Recommended),
		d("Jogo de pratos", "🍽️", 1, Recommended), d("Liquidificador", "🥤", 1, Recommended), d("Cafeteira", "☕", 1, Optional),
	}},
	{Name: "Quarto", Emoji: "🛏️", Reason: "Roupa de cama costuma ser um dos presentes mais escolhidos.", Desires: []Desire{
		d("Jogo de cama", "🛏️", 2, Essential), d("Travesseiros", "💤", 2, Essential), d("Edredom", "🧣", 1, Recommended),
	}},
	{Name: "Banheiro", Emoji: "🛁", Reason: "Itens de uso diário que se desgastam rápido.", Desires: []Desire{
		d("Kit de toalhas", "🧺", 2, Essential), d("Tapete de banheiro", "🟫", 1, Optional),
	}},
	{Name: "Sala", Emoji: "🛋️", Reason: "Para receber os amigos na casa nova.", Desires: []Desire{
		d("Luminária", "💡", 1, Optional), d("Manta para sofá", "🧶", 1, Optional),
	}},
	{Name: "Limpeza e organização", Emoji: "🧹", Reason: "Facilita a rotina desde o primeiro dia.", Desires: []Desire{
		d("Aspirador de pó", "🧹", 1, Recommended), d("Organizadores", "📦", 1, Optional),
	}},
}

var templates = map[string][]CategorySuggestion{
	"HOUSEWARMING": home,
	"WEDDING": append([]CategorySuggestion{
		{Name: "Lua de mel", Emoji: "✈️", Reason: "Muitos convidados preferem presentear experiências.", Desires: []Desire{
			d("Jantar especial na lua de mel", "🥂", 1, Recommended), d("Passeio na lua de mel", "🌅", 1, Optional),
		}},
	}, home...),
	"BABY_SHOWER": {
		{Name: "Higiene", Emoji: "🛁", Reason: "Itens usados todos os dias desde o nascimento.", Desires: []Desire{
			d("Fraldas tamanho P", "🧷", 6, Essential), d("Fraldas tamanho M", "🧷", 6, Essential), d("Lenços umedecidos", "🧻", 6, Essential),
			d("Banheira", "🛁", 1, Recommended),
		}},
		{Name: "Quarto do bebê", Emoji: "🧸", Reason: "Conforto e segurança para as primeiras noites.", Desires: []Desire{
			d("Kit berço", "🛏️", 1, Essential), d("Babá eletrônica", "📻", 1, Recommended), d("Mobile", "🌙", 1, Optional),
		}},
		{Name: "Roupinhas", Emoji: "👕", Reason: "Bebês crescem rápido: vale variar os tamanhos.", Desires: []Desire{
			d("Bodies RN", "👶", 4, Essential), d("Macacões tamanho P", "🧸", 4, Recommended),
		}},
		{Name: "Passeio", Emoji: "🚼", Reason: "Para os primeiros passeios com segurança.", Desires: []Desire{
			d("Bebê conforto", "🚗", 1, Essential), d("Bolsa maternidade", "👜", 1, Recommended),
		}},
	},
	"TRAVEL": {
		{Name: "Bagagem", Emoji: "🧳", Reason: "O básico para qualquer viagem.", Desires: []Desire{
			d("Mala de mão", "🧳", 1, Essential), d("Kit de organizadores de mala", "📦", 1, Optional),
		}},
		{Name: "Experiências", Emoji: "🌅", Reason: "Momentos que viram memórias.", Desires: []Desire{
			d("Passeio especial", "🗺️", 1, Recommended), d("Jantar na viagem", "🍝", 1, Optional),
		}},
	},
	"WISHLIST": {
		{Name: "Desejos", Emoji: "✨", Reason: "Um lugar para tudo o que você gostaria de ganhar.", Desires: nil},
	},
}
