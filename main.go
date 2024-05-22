package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

type Acao struct {
	Nome            string
	Cotacao         float64
	VPA             float64
	LPA             float64
	MediaDividendos float64
	Bazin           float64
	Graham          float64
}

func (a *Acao) String() string {
	return fmt.Sprintf("Nome: %s;Cotacao: %.2f;VPA: %.2f;LPA: %.2f;MediaDividendos: %.2f;Bazin: %.2f;Graham: %.2f", a.Nome, a.Cotacao, a.VPA, a.LPA, a.MediaDividendos, a.Bazin, a.Graham)
}

func (a *Acao) CSV() string {
	return fmt.Sprintf("%s;%.2f;%.2f;%.2f;%.2f;%.2f;%.2f\n", a.Nome, a.Cotacao, a.VPA, a.LPA, a.MediaDividendos, a.Bazin, a.Graham)
}

func Build(nome string) Acao {
	a := Acao{
		Nome: nome,
	}
	det_c := colly.NewCollector()
	det_c.OnHTML("body > div.center > div.conteudo.clearfix > table:nth-child(2) > tbody > tr:nth-child(1) > td.data.destaque.w3 > span", func(e *colly.HTMLElement) {
		cotacao, err := strconv.ParseFloat(strings.ReplaceAll(e.Text, ",", "."), 64)
		if err != nil {
			log.Fatal(err)
		}
		a.Cotacao = cotacao
	})

	det_c.OnHTML("body > div.center > div.conteudo.clearfix > table:nth-child(4) > tbody", func(e *colly.HTMLElement) {
		lpa, err := strconv.ParseFloat(strings.ReplaceAll(e.ChildText("tr:nth-child(2) > td:nth-child(6) > span"), ",", "."), 64)
		if err != nil {
			log.Fatal(err)
		}
		a.LPA = lpa
		vpa, err := strconv.ParseFloat(strings.ReplaceAll(e.ChildText("tr:nth-child(3) > td:nth-child(6) > span"), ",", "."), 64)
		if err != nil {
			log.Fatal(err)
		}
		a.VPA = vpa
	})
	// Before making a request print "Visiting"
	det_c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on https://example.com
	err := det_c.Visit("https://www.fundamentus.com.br/detalhes.php?papel=" + nome)
	if err != nil {
		log.Fatal(err)
	}

	prov_c := colly.NewCollector()
	prov_c.OnHTML("#resultado-anual", func(e *colly.HTMLElement) {
		i := 0
		for i < 6 {
			dividendos, err := strconv.ParseFloat(strings.ReplaceAll(e.ChildText(fmt.Sprintf("tr:nth-child(%d) > td:nth-child(2)", i+2)), ",", "."), 64)
			if err != nil {
				break
			}
			a.MediaDividendos += dividendos
			i++
		}
		a.MediaDividendos /= float64(i)

	})
	err = prov_c.Visit("https://www.fundamentus.com.br/proventos.php?papel=" + nome)
	if err != nil {
		log.Fatal(err)
	}
	a.Graham = math.Sqrt(22.5 * a.LPA * a.VPA)
	a.Bazin = a.MediaDividendos / 0.06
	return a

}

func main() {
	raw_data, err := os.ReadFile("input.txt")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(raw_data), "\r\n")
	var acoes []Acao
	for _, line := range lines {
		a := Build(line)
		acoes = append(acoes, a)
	}
	os.Remove("output.txt")
	f, err := os.OpenFile("output.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	f.WriteString("Nome;Cotacao;VPA;LPA;MediaDividendos;Bazin;Graham\n")
	for _, a := range acoes {
		// fmt.Printf("%#v \n", a)
		if _, err = f.WriteString(a.CSV()); err != nil {
			panic(err)
		}
		// os.WriteFile("output.txt", []byte(), fs.FileMode(os.O_APPEND))
	}

}

// Create a function to paarse a string to a float
// func parseFloat(s string) float64 {
// 	f, err := strconv.ParseFloat(s, 64)
