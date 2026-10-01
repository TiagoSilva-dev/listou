package mock

// Fictitious demo catalog. Prices and titles are illustrative only and are
// always labeled "demonstração" in the UI. Nothing here reflects real
// marketplace listings, prices, availability or ratings.
type fixture struct {
	id       string
	title    string
	brand    string
	category string
	emoji    string
	keywords string
	// price in cents per merchant code; missing merchant = not sold there
	prices map[string]int64
}

var fixtures = []fixture{
	{"AF-5L", "Fritadeira elétrica air fryer 5 litros", "Marca Demo", "Cozinha", "🍟", "air fryer airfryer fritadeira sem oleo 5l", map[string]int64{"AMAZON": 39900, "MERCADO_LIVRE": 37900, "SHOPEE": 38900}},
	{"AF-4L", "Air fryer compacta 4 litros", "Casa Demo", "Cozinha", "🍟", "air fryer airfryer fritadeira 4l compacta", map[string]int64{"AMAZON": 32990, "SHOPEE": 29990}},
	{"PAN-10", "Jogo de panelas antiaderente 10 peças", "Cozinha Demo", "Cozinha", "🍲", "jogo de panelas conjunto antiaderente", map[string]int64{"AMAZON": 54900, "MERCADO_LIVRE": 52990}},
	{"PAN-5", "Jogo de panelas inox 5 peças", "Inox Demo", "Cozinha", "🍲", "jogo de panelas inox conjunto", map[string]int64{"MERCADO_LIVRE": 69900, "SHOPEE": 67500}},
	{"FAQ-24", "Faqueiro inox 24 peças", "Mesa Demo", "Cozinha", "🍴", "faqueiro talheres inox jogo", map[string]int64{"AMAZON": 18990, "MERCADO_LIVRE": 17990, "SHOPEE": 16990}},
	{"PRA-20", "Aparelho de jantar porcelana 20 peças", "Mesa Demo", "Mesa posta", "🍽️", "jogo de pratos aparelho de jantar porcelana", map[string]int64{"AMAZON": 34900, "MERCADO_LIVRE": 32900}},
	{"LIQ-12", "Liquidificador 12 velocidades 1200W", "Eletro Demo", "Cozinha", "🥤", "liquidificador", map[string]int64{"AMAZON": 21900, "MERCADO_LIVRE": 19990, "SHOPEE": 20990}},
	{"CAF-ESP", "Cafeteira espresso 15 bar", "Café Demo", "Cozinha", "☕", "cafeteira espresso maquina de cafe", map[string]int64{"AMAZON": 89900, "MERCADO_LIVRE": 87900}},
	{"CAF-CAP", "Cafeteira de cápsulas", "Café Demo", "Cozinha", "☕", "cafeteira capsula maquina de cafe", map[string]int64{"AMAZON": 49900, "SHOPEE": 47900}},
	{"MIC-30", "Micro-ondas 30 litros", "Eletro Demo", "Cozinha", "📟", "micro-ondas microondas", map[string]int64{"AMAZON": 74900, "MERCADO_LIVRE": 72900}},
	{"BAT-PL", "Batedeira planetária 600W", "Eletro Demo", "Cozinha", "🧁", "batedeira planetaria", map[string]int64{"MERCADO_LIVRE": 59900, "SHOPEE": 56900}},
	{"PAN-ARR", "Panela elétrica de arroz", "Eletro Demo", "Cozinha", "🍚", "panela eletrica arroz", map[string]int64{"AMAZON": 18900, "SHOPEE": 17900}},
	{"CAMA-Q", "Jogo de cama queen 4 peças algodão", "Casa Demo", "Quarto", "🛏️", "jogo de cama queen lencol fronha", map[string]int64{"AMAZON": 24990, "MERCADO_LIVRE": 22990, "SHOPEE": 21990}},
	{"CAMA-K", "Jogo de cama king 400 fios", "Casa Demo", "Quarto", "🛏️", "jogo de cama king lencol 400 fios", map[string]int64{"AMAZON": 39990, "MERCADO_LIVRE": 37990}},
	{"TRAV-2", "Kit 2 travesseiros de fibra", "Sono Demo", "Quarto", "💤", "travesseiro travesseiros kit", map[string]int64{"AMAZON": 12990, "MERCADO_LIVRE": 11990, "SHOPEE": 9990}},
	{"EDR-Q", "Edredom queen dupla face", "Casa Demo", "Quarto", "🧣", "edredom coberta queen", map[string]int64{"MERCADO_LIVRE": 27990, "SHOPEE": 25990}},
	{"TOA-5", "Kit 5 toalhas de banho algodão", "Banho Demo", "Banheiro", "🧺", "toalhas kit toalha de banho rosto", map[string]int64{"AMAZON": 15990, "MERCADO_LIVRE": 14990, "SHOPEE": 13990}},
	{"TV-55", "Smart TV 55 polegadas 4K", "Tela Demo", "Sala", "📺", "smart tv televisao 55 4k", map[string]int64{"AMAZON": 279900, "MERCADO_LIVRE": 269900}},
	{"LUM-PISO", "Luminária de piso", "Luz Demo", "Sala", "💡", "luminaria abajur piso", map[string]int64{"AMAZON": 22900, "SHOPEE": 19900}},
	{"MANTA", "Manta para sofá tricô", "Casa Demo", "Sala", "🧶", "manta sofa", map[string]int64{"MERCADO_LIVRE": 8990, "SHOPEE": 7990}},
	{"ASP-VERT", "Aspirador de pó vertical sem fio", "Limpa Demo", "Limpeza", "🧹", "aspirador de po vertical", map[string]int64{"AMAZON": 59900, "MERCADO_LIVRE": 56900}},
	{"ROBO-ASP", "Robô aspirador inteligente", "Limpa Demo", "Limpeza", "🤖", "robo aspirador", map[string]int64{"AMAZON": 129900, "MERCADO_LIVRE": 124900}},
	{"ORG-KIT", "Kit organizadores de armário 6 peças", "Organiza Demo", "Organização", "📦", "organizadores organizador caixas", map[string]int64{"SHOPEE": 6990, "MERCADO_LIVRE": 7990}},
	{"GEL-FF", "Geladeira frost free 400 litros", "Frio Demo", "Cozinha", "🧊", "geladeira refrigerador frost free", map[string]int64{"AMAZON": 399900, "MERCADO_LIVRE": 389900}},
	{"FRA-P", "Pacote de fraldas tamanho P", "Bebê Demo", "Higiene", "🧷", "fraldas fralda tamanho p", map[string]int64{"AMAZON": 6990, "MERCADO_LIVRE": 6490, "SHOPEE": 5990}},
	{"FRA-M", "Pacote de fraldas tamanho M", "Bebê Demo", "Higiene", "🧷", "fraldas fralda tamanho m", map[string]int64{"AMAZON": 7490, "MERCADO_LIVRE": 6990, "SHOPEE": 6490}},
	{"LENCO", "Lenços umedecidos 4 pacotes", "Bebê Demo", "Higiene", "🧻", "lencos umedecidos", map[string]int64{"AMAZON": 3990, "SHOPEE": 3490}},
	{"BANHEIRA", "Banheira para bebê com suporte", "Bebê Demo", "Higiene", "🛁", "banheira bebe", map[string]int64{"MERCADO_LIVRE": 19990, "SHOPEE": 17990}},
	{"BABA-EL", "Babá eletrônica com câmera", "Bebê Demo", "Quarto do bebê", "📻", "baba eletronica camera", map[string]int64{"AMAZON": 34990, "MERCADO_LIVRE": 32990}},
	{"BEBE-CONF", "Bebê conforto 0 a 13 kg", "Bebê Demo", "Passeio", "🚗", "bebe conforto cadeirinha", map[string]int64{"AMAZON": 49990, "MERCADO_LIVRE": 47990}},
	{"MALA-M", "Mala de mão 10 kg rígida", "Viagem Demo", "Bagagem", "🧳", "mala de mao bagagem", map[string]int64{"AMAZON": 29990, "MERCADO_LIVRE": 27990, "SHOPEE": 25990}},
}
