package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)



func main(){
	os.Remove("lotto.csv")

	channel :=make(chan []string)
	results := [][]string{}

	res,_:=http.Get("https://search.daum.net/search?q=로또+당첨+번호")
	doc, _ := goquery.NewDocumentFromReader(res.Body)
	
	searchCards := doc.Find("select.opt_select").Find("option")

	numbers := []string{}

	searchCards.Each(func(i int, card *goquery.Selection) {
		if i==0{
			return
		}

		sampleRegexp := regexp.MustCompile("(,)|((회차).+)")
		numbers = append(numbers, sampleRegexp.ReplaceAllString(card.Text(),""))
	})
	
	for i := 0; i < len(numbers); i++ {
		go getLottoNum(numbers[i], channel)
	}

	for i := 0; i < len(numbers); i++ {
		result := <-channel
		results = append(results, result);
	}

	sort.Slice(results, func(i, j int) bool {
		v1,_ := strconv.Atoi(results[i][0])
		v2,_ := strconv.Atoi(results[j][0])

		return v1 < v2 
})

	makeCSV(results)
}

func checkErr(err error){
	if err != nil{
		log.Fatalln(err)
	}
}



func makeCSV(results [][]string) {
	file, err := os.Create("lotto.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"round","num1","num2","num3","num4","num5","num6", "bonus"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _,result  := range results{
		jobSlice := result
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}

}

func getLottoNum(num string, channel chan<- []string){
	res, _:= http.Get("https://search.daum.net/search?&q="+num+"회차%20로또")
	doc, _ := goquery.NewDocumentFromReader(res.Body)

	searchCards := doc.Find("div.lottonum").Find("span.ball")

	ns := []string{num}

	searchCards.Each(func(i int, card *goquery.Selection) {
		numTxt := card.Text()

		if numTxt != "보너스"{
			ns = append(ns, numTxt)
		}
	})

	channel<- ns
}